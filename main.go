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
		CloseConsolePrint: true,
	})

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
			slog.String("request", string(ctx.Body)))
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
