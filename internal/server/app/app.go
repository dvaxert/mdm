package serverapp

import (
	"io"
	"log/slog"

	grpcapp "github.com/dvaxert/mdm/internal/server/app/grpc"
	controlsrv "github.com/dvaxert/mdm/internal/server/services/control"
	managementsrv "github.com/dvaxert/mdm/internal/server/services/management"
	"github.com/dvaxert/mdm/internal/server/storage/sqlite"
)

type App struct {
	gRPCSrv *grpcapp.App
	storage io.Closer
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	managementSrv := managementsrv.New(log, storage)
	controlSrv := controlsrv.New(log, storage, managementSrv)

	grpcApp := grpcapp.New(log, grpcPort, managementSrv, controlSrv)

	return &App{
		gRPCSrv: grpcApp,
		storage: storage,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	return a.gRPCSrv.Run()
}

func (a *App) Stop() {
	a.gRPCSrv.Stop()
	a.storage.Close()
}
