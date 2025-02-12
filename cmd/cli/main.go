package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	controlv1 "github.com/dvaxert/mdm/api/gen/go/control"
	"github.com/dvaxert/mdm/internal/cli"
	"github.com/dvaxert/mdm/internal/domain/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := cli.MustLoadConfig()

	fmt.Println("starting cli")

	cc, err := grpc.NewClient(
		net.JoinHostPort(conf.Grpc.Address, conf.Grpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	client := controlv1.NewControlClient(cc)

	showCommandsList := func() {
		fmt.Println(
			`Commands:
			dlist - show list of devices
			dinfo $device_uuid - show info about device
			dstatus $device_uuid - show device status
			dfeature $device_uuid - show device features status
			ilist - show list of device info
			slist - show list of device status
			flist - show list of device features
			feature $device_id $feature_name $feature_state - change device feature state
			stop - exit program
		`,
		)
	}
	showCommandsList()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		stopped := false
		scanner := bufio.NewScanner(os.Stdin)

		for !stopped {
			fmt.Println("Enter command:")

			scanner.Scan()
			commandData := strings.Split(scanner.Text(), " ")

			switch commandData[0] {
			case "dlist":
				res, err := client.DeviceList(context.Background(), &controlv1.DeviceListRequest{})
				if err != nil {
					fmt.Printf("failed to get the device list: %s\n", err)
					continue
				}

				fmt.Println("Device list:")
				for _, device_uuid := range res.GetDeviceId() {
					fmt.Println(device_uuid)
				}

			case "dinfo":
				if len(commandData) != 2 {
					fmt.Println("the wrong arguments were passed to command")
					continue
				}

				res, err := client.DeviceInfo(
					context.Background(),
					&controlv1.DeviceInfoRequest{DeviceId: commandData[1]},
				)
				if err != nil {
					fmt.Printf("failed to get the device info: %s\n", err)
					continue
				}

				fmt.Printf("Device info: %s\n", models.DeviceType(res.GetDeviceType()).String())

			case "dstatus":
				if len(commandData) != 2 {
					fmt.Println("the wrong arguments were passed to command")
					continue
				}

				res, err := client.DeviceStatus(
					context.Background(),
					&controlv1.DeviceStatusRequest{DeviceId: commandData[1]},
				)
				if err != nil {
					fmt.Printf("failed to get the device info: %s\n", err)
					continue
				}

				fmt.Printf("Device status: location = '%s' battery = %d\n", res.GetLocation(), res.GetBattery())

			case "dfeature":
				if len(commandData) != 2 {
					fmt.Println("the wrong arguments were passed to command")
					continue
				}

				res, err := client.DeviceFeatures(
					context.Background(),
					&controlv1.DeviceFeaturesRequest{DeviceId: commandData[1]},
				)
				if err != nil {
					fmt.Printf("failed to get the device features: %s\n", err)
					continue
				}

				fmt.Println("Device features:")
				for k, v := range res.GetFeatures() {
					fmt.Printf("%s: %t\n", k, v)
				}

			case "ilist":
				res, err := client.DeviceInfoList(context.Background(), &controlv1.DeviceInfoListRequest{})
				if err != nil {
					fmt.Printf("failed to get the list of device info: %s\n", err)
					continue
				}

				fmt.Println("Devices info:")
				for _, item := range res.Items {
					fmt.Printf("%s: %s\n", item.GetDeviceId(), models.DeviceType(item.GetDeviceType()).String())
				}

			case "slist":
				res, err := client.DeviceStatusList(context.Background(), &controlv1.DeviceStatusListRequest{})
				if err != nil {
					fmt.Printf("failed to get the list of device status: %s\n", err)
					continue
				}

				fmt.Println("Devices status:")
				for _, item := range res.Items {
					fmt.Printf("%s: location='%s', battery=%d\n", item.GetDeviceId(), item.GetLocation(), item.GetBattery())
				}

			case "flist":
				res, err := client.DeviceFeaturesList(context.Background(), &controlv1.DeviceFeaturesListRequest{})
				if err != nil {
					fmt.Printf("failed to get the list of device features: %s\n", err)
					continue
				}

				fmt.Println("Devices features:")
				for _, item := range res.Items {
					fmt.Printf("device: %s\n", item.GetDeviceId())

					for k, v := range item.GetFeatures() {
						fmt.Printf("\t%s=%t\n", k, v)
					}
				}

			case "feature":
				if len(commandData) != 4 {
					fmt.Println("the wrong arguments were passed to command")
					continue
				}

				state, err := strconv.ParseBool(commandData[3])
				if err != nil {
					fmt.Println("the wrong arguments were passed to command")
					continue
				}

				// feature $device_id $feature_name $feature_state - change device feature state

				res, err := client.SetDeviceFeatureState(
					context.Background(),
					&controlv1.SetDeviceFeatureStateRequest{
						DeviceId: commandData[1],
						Feature:  commandData[2],
						State:    state,
					},
				)
				if err != nil {
					fmt.Printf("failed to get the device features: %s\n", err)
					continue
				}

				fmt.Printf("result: %t\n", res.GetSuccess())

			case "stop":
				stop <- os.Interrupt
				stopped = true

			default:
				fmt.Println("unknown command")
				showCommandsList()
			}
		}
	}()

	sig := <-stop
	fmt.Println("stopping cli", slog.String("signal", sig.String()))

	fmt.Println("cli stopped")
}
