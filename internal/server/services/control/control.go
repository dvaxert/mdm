package controlsrv

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/google/uuid"
)

type Control struct {
	log        *slog.Logger
	storage    StorageProvider
	management ManagementProvider
}

type ManagementProvider interface {
	SetDeviceFeatureState(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error
}

type StorageProvider interface {
	DeviceList(ctx context.Context) ([]models.Device, error)
	Device(ctx context.Context, device_id uuid.UUID) (models.Device, error)
	DeviceStatus(ctx context.Context, device_id uuid.UUID) (models.DeviceStatus, error)
	DeviceFeatures(ctx context.Context, device_id uuid.UUID) (models.DeviceFeatures, error)
	DeviceStatusList(ctx context.Context) ([]models.DeviceStatus, error)
	DeviceFeaturesList(ctx context.Context) ([]models.DeviceFeatures, error)
}

func New(log *slog.Logger, storage StorageProvider, management ManagementProvider) *Control {
	return &Control{
		log:        log,
		management: management,
		storage:    storage,
	}
}

func (c *Control) DeviceList(ctx context.Context) ([]string, error) {
	const op = "Control.DeviceList"

	log := c.log.With(slog.String("op", op))

	log.Info("attempting to prepare device list")

	log.With(slog.Any("storage", c.storage)).With(slog.Any("ctx", ctx)).Info("111")
	devices, err := c.storage.DeviceList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]string, 0, len(devices))
	for _, item := range devices {
		result = append(result, item.Uuid.String())
	}

	log.Info("device list prepared successfully")

	return result, nil
}

func (c *Control) DeviceInfo(ctx context.Context, device_id uuid.UUID) (models.Device, error) {
	const op = "Control.DeviceInfo"

	log := c.log.With(
		slog.String("op", op),
		slog.String("uuid", device_id.String()),
	)

	log.Info("attempting to prepare device info")

	device, err := c.storage.Device(ctx, device_id)
	if err != nil {
		return models.Device{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device info prepared successfully")

	return device, nil
}

func (c *Control) DeviceStatus(ctx context.Context, device_id uuid.UUID) (models.DeviceStatus, error) {
	const op = "Control.DeviceStatus"

	log := c.log.With(
		slog.String("op", op),
		slog.String("uuid", device_id.String()),
	)

	log.Info("attempting to prepare device status")

	status, err := c.storage.DeviceStatus(ctx, device_id)
	if err != nil {
		return models.DeviceStatus{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device status prepared successfully")

	return status, nil
}

func (c *Control) DeviceFeatures(ctx context.Context, device_id uuid.UUID) (models.DeviceFeatures, error) {
	const op = "Control.DeviceFeatures"

	log := c.log.With(
		slog.String("op", op),
		slog.String("uuid", device_id.String()),
	)

	log.Info("attempting to prepare device status")

	status, err := c.storage.DeviceFeatures(ctx, device_id)
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device status prepared successfully")

	return status, nil
}

func (c *Control) DeviceInfoList(ctx context.Context) ([]models.Device, error) {
	const op = "Control.DeviceInfoList"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to prepare device info list")

	list, err := c.storage.DeviceList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device info list prepared successfully")

	return list, nil
}

func (c *Control) DeviceStatusList(ctx context.Context) ([]models.DeviceStatus, error) {
	const op = "Control.DeviceStatusList"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to prepare device status list")

	list, err := c.storage.DeviceStatusList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device status list prepared successfully")

	return list, nil
}

func (c *Control) DeviceFeaturesList(ctx context.Context) ([]models.DeviceFeatures, error) {
	const op = "Control.DeviceFeaturesList"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to prepare device features list")

	list, err := c.storage.DeviceFeaturesList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device features list prepared successfully")

	return list, nil
}

func (c *Control) SetDeviceFeatureState(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error {
	const op = "Control.SetDeviceFeatureState"

	log := c.log.With(
		slog.String("op", op),
		slog.String("uuid", device_uuid.String()),
		slog.String("feature", feature),
		slog.Bool("state", state),
	)

	log.Info("attempting to set device feature state")

	if err := c.management.SetDeviceFeatureState(ctx, device_uuid, feature, state); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device feature state changed successfully")

	return nil
}
