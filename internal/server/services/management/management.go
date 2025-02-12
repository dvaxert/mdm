package managementsrv

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/google/uuid"
)

type Management struct {
	log     *slog.Logger
	storage StorageProvider
	states  map[uuid.UUID]bool // хранилище отображает для каких девайсов было изменено состояние
}

type StorageProvider interface {
	Close() error
	Device(ctx context.Context, device_uuid uuid.UUID) (models.Device, error)
	DeviceFeatures(ctx context.Context, device_uuid uuid.UUID) (models.DeviceFeatures, error)
	DeviceStatus(ctx context.Context, device_uuid uuid.UUID) (models.DeviceStatus, error)
	RegisterDevice(ctx context.Context, device_uuid uuid.UUID, device_type models.DeviceType) (int64, error)
	UpdateDeviceFeature(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error
	UpdateDeviceStatus(ctx context.Context, device_uuid uuid.UUID, location string, battery int) error
}

func New(log *slog.Logger, storage StorageProvider) *Management {
	return &Management{
		log:     log,
		storage: storage,
		states:  make(map[uuid.UUID]bool),
	}
}

func (m *Management) DeviceRegister(ctx context.Context, device_uuid uuid.UUID, device_type models.DeviceType) error {
	const op = "Management.DeviceRegister"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uuid", device_uuid.String()),
		slog.String("type", device_type.String()),
	)

	log.Info("attempting to register device")

	_, err := m.storage.RegisterDevice(ctx, device_uuid, device_type)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("device registered successfully")

	return nil
}

func (m *Management) DevicePing(ctx context.Context, device_uuid uuid.UUID, location string, battery int) (bool, error) {
	const op = "Management.DevicePing"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uuid", device_uuid.String()),
	)

	log.Info("attempting to process a ping from the device")

	err := m.storage.UpdateDeviceStatus(ctx, device_uuid, location, battery)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("ping processed successfully")

	stateChanged, ok := m.states[device_uuid]
	if !ok {
		return false, nil
	}

	return stateChanged, nil
}

func (m *Management) DeviceState(ctx context.Context, device_uuid uuid.UUID) (models.DeviceFeatures, error) {
	const op = "Management.DeviceState"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uuid", device_uuid.String()),
	)

	log.Info("attempt to prepare the state of the devices features")

	features, err := m.storage.DeviceFeatures(ctx, device_uuid)
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	delete(m.states, device_uuid)

	log.Info("state of device features successfully prepared")

	return features, nil
}

func (m *Management) SetDeviceFeatureState(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error {
	const op = "Management.SetDeviceFeatureState"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uuid", device_uuid.String()),
		slog.String("feature", feature),
		slog.Bool("state", state),
	)

	log.Info("attempt to change state device feature")

	err := m.storage.UpdateDeviceFeature(ctx, device_uuid, feature, state)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	m.states[device_uuid] = true

	log.Info("state of device feature successfully changed")

	return nil
}
