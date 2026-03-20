package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

// Setup initializes the global slog logger.
//
// Output goes to stderr. When stderr is a TTY, colored output is used via tint.
// Otherwise plain text is used.
func setupLogger() {
	level := slog.LevelInfo

	if *debug {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if isatty.IsTerminal(os.Stderr.Fd()) {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:       level,
			TimeFormat:  "15:04:05.000000",
			NoColor:     os.Getenv("NO_COLOR") != "",
			AddSource:   true,
			ReplaceAttr: replaceSourceAttr,
		})
	} else {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: replaceAttr,
		})
	}

	slog.SetDefault(slog.New(handler))
}

// replaceAttr is used by the non-TTY handler. It drops the timestamp (journald
// records its own) and trims the source file path to the last two components.
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			src.File = trimSourcePath(src.File)
		}
	case slog.TimeKey:
		return slog.Attr{}
	}

	return a
}

// replaceSourceAttr is used by the TTY handler. It trims the source path but
// leaves the timestamp alone (tint renders it via TimeFormat).
func replaceSourceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if src, ok := a.Value.Any().(*slog.Source); ok {
			src.File = trimSourcePath(src.File)
		}
	}

	return a
}

// trimSourcePath returns the last two slash-separated components of path,
// e.g. "hardware/emc230x/device.go" from a full absolute path.
func trimSourcePath(path string) string {
	idx := strings.LastIndexByte(path, '/')
	if idx < 0 {
		return path
	}

	idx = strings.LastIndexByte(path[:idx], '/')
	if idx < 0 {
		return path
	}

	return path[idx+1:]
}
