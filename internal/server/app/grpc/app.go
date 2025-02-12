package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	controlgrpc "github.com/dvaxert/mdm/internal/server/grpc/control"
	managementgrpc "github.com/dvaxert/mdm/internal/server/grpc/management"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
	mng managementgrpc.Management,
	ctl controlgrpc.Control,
) *App {
	gRPCServer := grpc.NewServer()

	controlgrpc.Register(gRPCServer, ctl)
	managementgrpc.Register(gRPCServer, mng)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is runnig", slog.String("address", listener.Addr().String()))

	if err = a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
