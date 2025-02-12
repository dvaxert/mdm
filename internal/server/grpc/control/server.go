package controlgrpc

import (
	"context"

	controlv1 "github.com/dvaxert/mdm/api/gen/go/control"
	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Control interface {
	DeviceFeatures(ctx context.Context, device_uuid uuid.UUID) (models.DeviceFeatures, error)
	DeviceFeaturesList(ctx context.Context) ([]models.DeviceFeatures, error)
	DeviceInfo(ctx context.Context, device_uuid uuid.UUID) (models.Device, error)
	DeviceInfoList(ctx context.Context) ([]models.Device, error)
	DeviceList(ctx context.Context) ([]string, error)
	DeviceStatus(ctx context.Context, device_uuid uuid.UUID) (models.DeviceStatus, error)
	DeviceStatusList(ctx context.Context) ([]models.DeviceStatus, error)
	SetDeviceFeatureState(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error
}

type serverApi struct {
	controlv1.UnimplementedControlServer
	control Control
}

func Register(gRPC *grpc.Server, control Control) {
	controlv1.RegisterControlServer(gRPC, &serverApi{control: control})
}

func (s *serverApi) DeviceList(
	ctx context.Context,
	req *controlv1.DeviceListRequest,
) (*controlv1.DeviceListResponse, error) {
	list, err := s.control.DeviceList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &controlv1.DeviceListResponse{DeviceId: list}, nil
}

func (s *serverApi) DeviceInfo(
	ctx context.Context,
	req *controlv1.DeviceInfoRequest,
) (*controlv1.DeviceInfoResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	device, err := s.control.DeviceInfo(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &controlv1.DeviceInfoResponse{DeviceType: int32(device.Type)}, nil
}

func (s *serverApi) DeviceStatus(
	ctx context.Context,
	req *controlv1.DeviceStatusRequest,
) (*controlv1.DeviceStatusResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	deviceStatus, err := s.control.DeviceStatus(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &controlv1.DeviceStatusResponse{
		Location: deviceStatus.Location,
		Battery:  int32(deviceStatus.Battery),
	}, nil
}

func (s *serverApi) DeviceFeatures(
	ctx context.Context,
	req *controlv1.DeviceFeaturesRequest,
) (*controlv1.DeviceFeaturesResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	id, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	deviceFeatures, err := s.control.DeviceFeatures(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &controlv1.DeviceFeaturesResponse{
		Features: deviceFeatures.Features,
	}, nil
}

func (s *serverApi) DeviceInfoList(
	ctx context.Context,
	req *controlv1.DeviceInfoListRequest,
) (*controlv1.DeviceInfoListResponse, error) {
	list, err := s.control.DeviceInfoList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	result := make([]*controlv1.DeviceInfoListItem, 0, len(list))
	for _, item := range list {
		result = append(result, &controlv1.DeviceInfoListItem{
			DeviceId:   item.Uuid.String(),
			DeviceType: int32(item.Type),
		})
	}

	return &controlv1.DeviceInfoListResponse{
		Items: result,
	}, nil
}

func (s *serverApi) DeviceStatusList(
	ctx context.Context,
	req *controlv1.DeviceStatusListRequest,
) (*controlv1.DeviceStatusListResponse, error) {
	statusList, err := s.control.DeviceStatusList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := make([]*controlv1.DeviceStatusListItem, 0, len(statusList))
	for _, item := range statusList {
		result = append(result, &controlv1.DeviceStatusListItem{
			DeviceId: item.DeviceUuid.String(),
			Location: item.Location,
			Battery:  int32(item.Battery),
		})
	}

	return &controlv1.DeviceStatusListResponse{
		Items: result,
	}, nil
}

func (s *serverApi) SetDeviceFeatureState(
	ctx context.Context,
	req *controlv1.SetDeviceFeatureStateRequest,
) (*controlv1.SetDeviceFeatureStateResponse, error) {
	if req.DeviceId == "" {
		return nil, status.Error(codes.InvalidArgument, "device id is required")
	}

	if req.Feature == "" {
		return nil, status.Error(codes.InvalidArgument, "feature is required")
	}

	uuid, err := uuid.Parse(req.GetDeviceId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect device id")
	}

	if err = s.control.SetDeviceFeatureState(ctx, uuid, req.GetFeature(), req.GetState()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &controlv1.SetDeviceFeatureStateResponse{Success: true}, nil
}

func (s *serverApi) DeviceFeaturesList(
	ctx context.Context,
	req *controlv1.DeviceFeaturesListRequest,
) (*controlv1.DeviceFeaturesListResponse, error) {
	featuresList, err := s.control.DeviceFeaturesList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := make([]*controlv1.DeviceFeaturesListItem, 0, len(featuresList))
	for _, item := range featuresList {
		result = append(result, &controlv1.DeviceFeaturesListItem{
			DeviceId: item.DeviceUuid.String(),
			Features: item.Features,
		})
	}

	return &controlv1.DeviceFeaturesListResponse{
		Items: result,
	}, nil
}
