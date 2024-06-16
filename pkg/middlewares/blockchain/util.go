package blockchain

import (
	"log/slog"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/xoctopus/datatypex"
)

type stacks struct {
	errors []error
}

func (s *stacks) Append(err error, message string, args ...any) {
	if err != nil {
		err = errors.Wrapf(err, message, args...)
	}
	s.errors = append(s.errors, err)
}

func (s *stacks) TrimLast() {
	if len(s.errors) > 0 {
		s.errors = s.errors[0 : len(s.errors)-1]
	}
}

func (s *stacks) Final() error {
	var final error
	for _, err := range s.errors {
		if err != nil {
			if final == nil {
				final = err
			} else {
				final = errors.Wrap(err, final.Error())
			}
		}
	}
	return final
}

var logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
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
	},
}))
