package logging

import (
	"log/slog"
	"os"
	"path/filepath"
)

// logging context keys
const (
	ViewKey      = "view"
	PaneKey      = "pane"
	ComponentKey = "component"
	DialogKey    = "dialog"
)

const logfile = "dynamite.log"

// NOTE: emergency function for immediate debugging
func LogDebug(s string) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	logloc := dir
	path := filepath.Join(logloc, logfile)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Debug(s)
}
