package managementgrpc

import (
	"context"

	managementv1 "github.com/dvaxert/mdm/api/gen/go/management"
	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Management interface {
	DevicePing(ctx context.Context, device_uuid uuid.UUID, location string, battery int) (bool, error)
	DeviceRegister(ctx context.Context, device_uuid uuid.UUID, device_type models.DeviceType) error
	DeviceState(ctx context.Context, device_uuid uuid.UUID) (models.DeviceFeatures, error)
}

type serverApi struct {
	managementv1.UnimplementedDeviceManagementServer
	management Management
}

func Register(gRPC *grpc.Server, management Management) {
	managementv1.RegisterDeviceManagementServer(gRPC, &serverApi{management: management})
}

func (s *serverApi) DeviceRegister(
	ctx context.Context,
	req *managementv1.DeviceRegisterRequest,
) (*managementv1.DeviceRegisterResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	dType := models.DeviceType(req.GetDeviceType())
	if dType >= models.DeviceTypeCount {
		return nil, status.Error(codes.InvalidArgument, "incorrect device type")
	}

	err = s.management.DeviceRegister(ctx, id, dType)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &managementv1.DeviceRegisterResponse{Success: true}, nil
}

func (s *serverApi) DevicePing(
	ctx context.Context,
	req *managementv1.DevicePingRequest,
) (*managementv1.DevicePingResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	if req.Location == "" {
		return nil, status.Error(codes.InvalidArgument, "location is required")
	}

	if req.Battery < 0 || req.Battery > 100 {
		return nil, status.Error(codes.InvalidArgument, "incorrect battery state")
	}

	stateChanged, err := s.management.DevicePing(ctx, id, req.GetLocation(), int(req.GetBattery()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &managementv1.DevicePingResponse{StateChanged: stateChanged}, nil
}

func (s *serverApi) DeviceState(
	ctx context.Context,
	req *managementv1.DeviceStateRequest,
) (*managementv1.DeviceStateResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	features, err := s.management.DeviceState(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &managementv1.DeviceStateResponse{Features: features.Features}, nil
}
