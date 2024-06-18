package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/xoctopus/datatypex"
)

func AttrReplacer(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case "time":
		a.Key = "@ts"
		a.Value = slog.StringValue(time.Now().Format(datatypex.DefaultTimestampLayout))
		return a
	case "level":
		a.Key = "@lv"
		switch a.Value.Any().(slog.Level) {
		case slog.LevelDebug:
			a.Value = slog.StringValue("deb")
		case slog.LevelInfo:
			a.Value = slog.StringValue("inf")
		case slog.LevelError:
			a.Value = slog.StringValue("err")
		case slog.LevelWarn:
			a.Value = slog.StringValue("wrn")
		default:
			a.Value = slog.StringValue("deb")
		}
		return a
	case "msg":
		a.Key = "@msg"
		return a
	default:
		return a
	}
}

var logger = logr.FromSlogHandler(
	slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: AttrReplacer,
	}),
)

var Default = &Logger{Logger: &logger}
