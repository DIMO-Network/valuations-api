package services

import (
	"context"
	"fmt"
	"time"

	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	gocache "github.com/patrickmn/go-cache"

	"google.golang.org/grpc"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go
type UserDeviceService interface {
	GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error)
}

type userDeviceService struct {
	devicesConn *grpc.ClientConn
	memoryCache *gocache.Cache
}

func NewUserDeviceService(devicesConn *grpc.ClientConn) UserDeviceService {
	c := gocache.New(8*time.Hour, 15*time.Minute)
	return &userDeviceService{devicesConn: devicesConn, memoryCache: c}
}

// GetUserDevice gets the userDevice from devices-api, checks in local cache first
func (das *userDeviceService) GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error) {
	if len(userDeviceID) == 0 {
		return nil, fmt.Errorf("user device id was empty - invalid")
	}
	var err error
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)

	var userDevice *pb.UserDevice
	get, found := das.memoryCache.Get("ud_" + userDeviceID)
	if found {
		userDevice = get.(*pb.UserDevice)
	} else {
		userDevice, err = deviceClient.GetUserDevice(ctx, &pb.GetUserDeviceRequest{
			Id: userDeviceID,
		})
		if err != nil {
			return nil, err
		}
		das.memoryCache.Set("ud_"+userDeviceID, userDevice, time.Hour*24)
	}

	return userDevice, nil
}
