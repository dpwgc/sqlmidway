package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"time"
)

var Logger *slog.Logger

func InitLog() {
	r := &lumberjack.Logger{
		Filename:   Config().Log.Path + "/runtime.log",
		LocalTime:  true,
		MaxSize:    Config().Log.Size,
		MaxAge:     Config().Log.Age,
		MaxBackups: Config().Log.Backups,
		Compress:   false,
	}
	Logger = slog.New(slog.NewTextHandler(r, &slog.HandlerOptions{
		AddSource: true, // 输出日志语句的位置信息
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // 格式化 key 为 "time" 的属性值
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}
			}
			return a
		},
	}))
}
