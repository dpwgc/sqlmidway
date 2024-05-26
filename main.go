package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/dpwgc/easierweb"
	"log/slog"
	"net/http"
	"strings"
)

func main() {
	InitConfig()
	InitLog()
	api, err := NewAPI()
	if err != nil {
		Logger.Error(err.Error())
		return
	}

	router := easierweb.New(easierweb.RouterOptions{
		ErrorHandle:       errorHandle(),
		Logger:            Logger,
		RequestHandle:     requestHandle(),
		ResponseHandle:    responseHandle(),
		CloseConsolePrint: true,
	}).Use(middleware())

	router.EasyPOST("/query/:db/:group/:api", api.Query)
	router.EasyPOST("/command/:db/:group/:api", api.Command)
	router.EasyGET("/info", api.Info)
	router.EasyGET("/health", api.Health)

	host := fmt.Sprintf("%s:%v", Config().Server.Addr, Config().Server.Port)
	if Config().Server.TLS {
		err = router.RunTLS(host, Config().Server.CertFile, Config().Server.KeyFile, &tls.Config{})
	} else {
		err = router.Run(host)
	}
	Logger.Error(err.Error())
}

func errorHandle() easierweb.ErrorHandle {
	return func(ctx *easierweb.Context, err any) {
		errStr := fmt.Sprintf("%v", err)
		ctx.Logger.Error(errStr, slog.String("url", ctx.Request.URL.String()),
			slog.String("client", ctx.Request.RemoteAddr),
			slog.String("body", string(ctx.Body)))
		ctx.WriteJSON(http.StatusBadRequest, ErrorReply{
			Error: errStr,
		})
	}
}

func requestHandle() easierweb.RequestHandle {
	return func(ctx *easierweb.Context, reqObj any) error {
		if len(ctx.Body) > 0 {
			decoder := json.NewDecoder(strings.NewReader(string(ctx.Body)))
			decoder.UseNumber() // UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.
			if err := decoder.Decode(reqObj); err != nil {
				return err
			}
		}
		return nil
	}
}

func responseHandle() easierweb.ResponseHandle {
	return func(ctx *easierweb.Context, result any, err error) {
		if err != nil {
			if result != nil {
				errStr := fmt.Sprintf("%v", err)
				ctx.Logger.Error(errStr, slog.String("url", ctx.Request.URL.String()),
					slog.String("client", ctx.Request.RemoteAddr),
					slog.String("body", string(ctx.Body)))
				ctx.WriteJSON(http.StatusBadRequest, result)
				return
			}
			panic(err)
		}
		if result == nil {
			ctx.NoContent(http.StatusNoContent)
			return
		}
		ctx.WriteJSON(http.StatusOK, result)
	}
}

func middleware() easierweb.Handle {
	return func(ctx *easierweb.Context) {

		if Config().Server.Auth && ctx.Route != "/health" {
			ok := false
			for _, a := range Config().Server.Accounts {
				if ctx.Header.Get("Username") == a.Username && ctx.Header.Get("Password") == a.Password {
					ok = true
					break
				}
				if ctx.Header.Get("username") == a.Username && ctx.Header.Get("password") == a.Password {
					ok = true
					break
				}
			}
			if ok {
				ctx.Next()
			} else {
				errStr := "authentication failed"
				ctx.Logger.Error(errStr, slog.String("url", ctx.Request.URL.String()),
					slog.String("client", ctx.Request.RemoteAddr),
					slog.Any("header", ctx.Header))
				ctx.WriteJSON(http.StatusForbidden, &ErrorReply{
					Error: errStr,
				})
				ctx.Abort()
			}
		} else {
			ctx.Next()
		}

		if ctx.Route == "/health" {
			return
		}

		body := ""
		sizeLimit := 1024 * 1024
		if len(ctx.Body) > 0 {
			if len(ctx.Body) > sizeLimit {
				body = "body is too large"
			} else {
				body = string(ctx.Body)
			}
		}
		ctx.Logger.Info(fmt.Sprintf("%s -> %s", ctx.Request.RemoteAddr, ctx.Request.URL.String()),
			slog.String("body", body),
			slog.Int("code", ctx.Code))
	}
}
