package logging

import (
	"log/slog"

	u "github.com/wolfwfr/dynamite/pkg/util"
)

const (
	LevelTrace = slog.Level(-8)
)

var LevelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
}

func ReplaceLevelName(a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		label, exists := LevelNames[level]
		u.Ternary(label, level.String(), exists)
		a.Value = slog.StringValue(label)
	}
	return a
}
