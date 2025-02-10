package main

import (
	"log/slog"

	"github.com/dvaxert/mdm/internal/server"
	"github.com/dvaxert/mdm/pkg/logger"
)

func main() {
	conf := server.MustLoadConfig()

	log := logger.MustSetup(conf.Env)
	log.Info("сервер запущен", slog.Any("config", conf))
}
