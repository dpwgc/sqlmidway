package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/dpwgc/easierweb"
	"log/slog"
	"net/http"
)

func InitCore() {
	api, err := NewAPI()
	if err != nil {
		Logger.Error(err.Error())
		return
	}

	router := easierweb.New(easierweb.RouterOptions{
		ErrorHandle:    errorHandle(),
		ResponseHandle: responseHandle(),
		Logger:         Logger,
	}).Use(logMiddleware())

	router.EasyPOST("/query/:service/:method", api.Query)
	router.EasyPOST("/command/:service/:method", api.Command)
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
		ctx.WriteJSON(http.StatusBadRequest, ErrorReply{
			Error: fmt.Sprintf("%s", err),
		})
	}
}

func responseHandle() easierweb.ResponseHandle {
	return func(ctx *easierweb.Context, result any, err error) {
		if err != nil {
			ctx.WriteJSON(http.StatusBadRequest, ErrorReply{
				Error: err.Error(),
			})
		} else {
			if result == nil {
				ctx.NoContent(http.StatusNoContent)
				return
			}
			ctx.WriteJSON(http.StatusOK, result)
		}
	}
}

func logMiddleware() easierweb.Handle {
	return func(ctx *easierweb.Context) {

		ctx.Next()

		if ctx.Proto() == "/health" {
			return
		}

		path := ""
		query := ""
		body := ""
		result := ""

		if len(ctx.Path) > 0 {
			marshal, err := json.Marshal(ctx.Path)
			if err != nil {
				path = err.Error()
			} else {
				path = string(marshal)
			}
		}
		if len(ctx.Query) > 0 {
			marshal, err := json.Marshal(ctx.Query)
			if err != nil {
				query = err.Error()
			} else {
				query = string(marshal)
			}
		}
		sizeLimit := 1024 * 1024
		if len(ctx.Body) > 0 {
			if len(ctx.Body) > sizeLimit {
				body = "body is too large"
			} else {
				body = string(ctx.Body)
			}
		}
		if len(ctx.Result) > 0 {
			if len(ctx.Result) > sizeLimit {
				result = "result is too large"
			} else {
				result = string(ctx.Result)
			}
		}

		ctx.Logger.Info(ctx.Proto(), slog.String("method", ctx.Request.Method),
			slog.String("url", ctx.Request.URL.String()),
			slog.String("client", ctx.Request.RemoteAddr),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("body", body),
			slog.Int("code", ctx.Code),
			slog.String("result", result))
	}
}
