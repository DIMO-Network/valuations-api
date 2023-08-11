package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	gocache "github.com/patrickmn/go-cache"

	"google.golang.org/grpc"
)

//go:generate mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go
type UserDeviceAPIService interface {
	GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error)
	GetAllUserDevice(ctx context.Context, wmi string) ([]*pb.UserDevice, error)
	UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error
}

type userDeviceAPIService struct {
	devicesConn *grpc.ClientConn
	memoryCache *gocache.Cache
}

func NewUserDeviceService(devicesConn *grpc.ClientConn) UserDeviceAPIService {
	c := gocache.New(8*time.Hour, 15*time.Minute)
	return &userDeviceAPIService{devicesConn: devicesConn, memoryCache: c}
}

// GetUserDevice gets the userDevice from devices-api, checks in local cache first
func (das *userDeviceAPIService) GetUserDevice(ctx context.Context, userDeviceID string) (*pb.UserDevice, error) {
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

func (das *userDeviceAPIService) UpdateUserDeviceMetadata(ctx context.Context, request *pb.UpdateUserDeviceMetadataRequest) error {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)
	_, err := deviceClient.UpdateUserDeviceMetadata(ctx, request)
	return err
}

// GetAllUserDevice gets all userDevices from devices-api
func (das *userDeviceAPIService) GetAllUserDevice(ctx context.Context, wmi string) ([]*pb.UserDevice, error) {
	deviceClient := pb.NewUserDeviceServiceClient(das.devicesConn)
	all, err := deviceClient.GetAllUserDevice(ctx, &pb.GetAllUserDeviceRequest{Wmi: wmi})
	if err != nil {
		return nil, err
	}

	var useDevices []*pb.UserDevice
	for {
		response, err := all.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while receiving response: %v", err)
		}

		useDevices = append(useDevices, response)
	}

	return useDevices, nil
}
