package logger

import (
	"log/slog"
	"os"

	"github.com/go-logr/logr"
)

type Logger struct {
	Level slog.Level

	*logr.Logger `env:"-"`
}

func (l *Logger) SetDefault() {
	if l.Level == 0 {
		l.Level = slog.LevelDebug
	}
}

func (l *Logger) Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       l.Level,
		ReplaceAttr: AttrReplacer,
	})
	ll := logr.FromSlogHandler(handler)
	l.Logger = &ll
}
