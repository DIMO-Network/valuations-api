package services

import (
	"context"

	pb "github.com/DIMO-Network/device-data-api/pkg/grpc"
	"google.golang.org/grpc"
)

//go:generate mockgen -source user_device_data_service.go -destination mocks/user_device_data_service_mock.go
type UserDeviceDataAPIService interface {
	GetUserDeviceData(ctx context.Context, userDeviceID string, ddID string) (*pb.UserDeviceDataResponse, error)
	GetVehicleRawData(ctx context.Context, userDeviceID string) (*pb.RawDeviceDataResponse, error)
}

func NewUserDeviceDataAPIService(ddConn *grpc.ClientConn) UserDeviceDataAPIService {
	return &userDeviceDataAPIService{deviceDataConn: ddConn}
}

type userDeviceDataAPIService struct {
	deviceDataConn *grpc.ClientConn
}

func (dda *userDeviceDataAPIService) GetUserDeviceData(ctx context.Context, userDeviceID string, ddID string) (*pb.UserDeviceDataResponse, error) {
	deviceClient := pb.NewUserDeviceDataServiceClient(dda.deviceDataConn)
	userDeviceData, err := deviceClient.GetUserDeviceData(ctx, &pb.UserDeviceDataRequest{
		UserDeviceId:       userDeviceID,
		DeviceDefinitionId: ddID,
	})
	if err != nil {
		return nil, err
	}

	return userDeviceData, nil
}

func (dda *userDeviceDataAPIService) GetVehicleRawData(ctx context.Context, userDeviceID string) (*pb.RawDeviceDataResponse, error) {
	deviceClient := pb.NewUserDeviceDataServiceClient(dda.deviceDataConn)
	deviceData, err := deviceClient.GetRawDeviceData(ctx, &pb.RawDeviceDataRequest{
		UserDeviceId: userDeviceID,
	})
	if err != nil {
		return nil, err
	}

	return deviceData, nil
}
