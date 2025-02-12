package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	managementv1 "github.com/dvaxert/mdm/api/gen/go/management"
	"github.com/dvaxert/mdm/internal/device"
	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/dvaxert/mdm/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := device.MustLoadConfig()

	state := models.DefaultFeatures

	log := logger.MustSetup(conf.Env).With(slog.Any("state", state))
	log.Info("starting device", slog.Any("config", conf))

	cc, err := grpc.NewClient(
		net.JoinHostPort(conf.Grpc.Address, conf.Grpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	client := managementv1.NewDeviceManagementClient(cc)

	res, err := client.DeviceRegister(
		context.Background(),
		&managementv1.DeviceRegisterRequest{
			DeviceId:   conf.Uuid,
			DeviceType: int32(conf.DeviceType),
		},
	)
	if err != nil {
		panic(err)
	}

	if !res.Success {
		panic("failed to register on the server")
	}

	go func() {
		for {
			time.Sleep(conf.PingPeriod)

			log.Info("attempting to send device ping")

			pingRes, err := client.DevicePing(
				context.Background(),
				&managementv1.DevicePingRequest{
					DeviceId: conf.Uuid,
					Location: conf.Location,
					Battery:  int32(conf.Battery),
				},
			)
			if err != nil {
				log.Error("error when sending a ping to the server", slog.Any("error", err))
			}

			if pingRes.StateChanged {
				log.Info("device state change detected, request new state")

				stateRes, err := client.DeviceState(
					context.Background(),
					&managementv1.DeviceStateRequest{
						DeviceId: conf.Uuid,
					},
				)
				if err != nil {
					log.Error("error when requesting a new device state", slog.Any("error", err))
				}

				state = stateRes.Features
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop
	log.Info("stopping device", slog.String("signal", sig.String()))

	log.Info("device stopped")
}
