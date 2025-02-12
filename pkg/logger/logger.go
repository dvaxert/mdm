package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func MustSetup(env string) *slog.Logger {
	log, err := Setup(env)
	if err != nil {
		panic(err)
	}

	return log
}

func Setup(env string) (*slog.Logger, error) {
	var handler slog.Handler

	switch env {
	case envLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		return nil, fmt.Errorf("invalid env value passed to logger.Setup")
	}

	return slog.New(handler), nil
}
