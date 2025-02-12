package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dvaxert/mdm/internal/server"
	serverapp "github.com/dvaxert/mdm/internal/server/app"
	"github.com/dvaxert/mdm/pkg/logger"
)

func main() {
	conf := server.MustLoadConfig()

	log := logger.MustSetup(conf.Env)
	log.Info("starting application", slog.Any("config", conf))

	application := serverapp.New(log, conf.Grpc.Port, conf.StoragePath)
	go application.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop
	log.Info("stopping application", slog.String("signal", sig.String()))

	application.Stop()

	log.Info("application stopped")
}
