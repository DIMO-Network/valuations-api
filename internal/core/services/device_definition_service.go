package services

import (
	"context"
	"fmt"

	pb "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"google.golang.org/grpc"
)

//go:generate mockgen -source device_definitions_service.go -destination mocks/device_definitions_service_mock.go
type DeviceDefinitionsAPIService interface {
	GetDeviceDefinitionByID(ctx context.Context, id string) (*pb.GetDeviceDefinitionItemResponse, error)
	GetDeviceStyleByExternalID(ctx context.Context, id string) (*pb.DeviceStyle, error)
	UpdateDeviceDefAttrs(ctx context.Context, deviceDef *pb.GetDeviceDefinitionItemResponse, vinInfo map[string]any) error
}

func NewDeviceDefinitionsAPIService(ddConn *grpc.ClientConn) DeviceDefinitionsAPIService {
	return &deviceDefinitionsAPIService{deviceDefinitionsConn: ddConn}
}

type deviceDefinitionsAPIService struct {
	deviceDefinitionsConn *grpc.ClientConn
}

func (dda *deviceDefinitionsAPIService) GetDeviceDefinitionByID(ctx context.Context, id string) (*pb.GetDeviceDefinitionItemResponse, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("device definition id was empty - invalid")
	}
	definitionsClient := pb.NewDeviceDefinitionServiceClient(dda.deviceDefinitionsConn)

	def, err := definitionsClient.GetDeviceDefinitionByID(ctx, &pb.GetDeviceDefinitionRequest{
		Ids: []string{id},
	})
	if err != nil {
		return nil, err
	}

	return def.DeviceDefinitions[0], nil
}

func (dda *deviceDefinitionsAPIService) GetDeviceStyleByExternalID(ctx context.Context, id string) (*pb.DeviceStyle, error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("device definition id was empty - invalid")
	}
	definitionsClient := pb.NewDeviceDefinitionServiceClient(dda.deviceDefinitionsConn)

	style, err := definitionsClient.GetDeviceStyleByExternalID(ctx, &pb.GetDeviceStyleByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return style, nil
}

func (dda *deviceDefinitionsAPIService) UpdateDeviceDefAttrs(ctx context.Context, deviceDef *pb.GetDeviceDefinitionItemResponse, vinInfo map[string]any) error {
	deviceAttributes := buildDeviceAttributes(deviceDef.DeviceAttributes, vinInfo)

	definitionsClient := pb.NewDeviceDefinitionServiceClient(dda.deviceDefinitionsConn)

	_, err := definitionsClient.UpdateDeviceDefinition(ctx, &pb.UpdateDeviceDefinitionRequest{
		DeviceDefinitionId: deviceDef.DeviceDefinitionId,
		DeviceAttributes:   deviceAttributes,
	})

	if err != nil {
		return err
	}

	return nil
}

// buildDeviceAttributes returns list of set attributes based on what already exists and vinInfo pulled from drivly. based on a predetermined list
func buildDeviceAttributes(existingDeviceAttrs []*pb.DeviceTypeAttribute, vinInfo map[string]any) []*pb.DeviceTypeAttributeRequest {
	// TODO: replace seekAttributes with a better solution based on device_types.attributes
	seekAttributes := map[string]string{
		// {device attribute, must match device_types.properties}: {vin info from drivly}
		"mpg_city":               "mpgCity",
		"mpg_highway":            "mpgHighway",
		"mpg":                    "mpg",
		"base_msrp":              "msrpBase",
		"fuel_tank_capacity_gal": "fuelTankCapacityGal",
		"fuel_type":              "fuel",
		"wheelbase":              "wheelbase",
		"generation":             "generation",
		"number_of_doors":        "doors",
		"manufacturer_code":      "manufacturerCode",
		"driven_wheels":          "drive",
	}

	addedAttrCount := 0
	// build array of already present device_attributes and remove any already set satisfactorily from seekAttributes map
	var deviceAttributes []*pb.DeviceTypeAttributeRequest //nolint
	for _, attr := range existingDeviceAttrs {
		deviceAttributes = append(deviceAttributes, &pb.DeviceTypeAttributeRequest{
			Name:  attr.Name,
			Value: attr.Value,
		})
		// todo: 0 value attributes could be decimal form in string eg. 0.00000 . Convert value to int, and then compare to 0 again?
		if _, exists := seekAttributes[attr.Name]; exists && attr.Value != "" && attr.Value != "0" {
			// already set, no longer seeking it
			delete(seekAttributes, attr.Name)
		}
	}
	// iterate over remaining attributes
	for k, attr := range seekAttributes {
		if v, ok := vinInfo[attr]; ok && v != nil {
			val := fmt.Sprintf("%v", v)
			// lookup the existing device attribute and set it if exists
			existing := false
			for _, attribute := range deviceAttributes {
				if attribute.Name == k {
					attribute.Value = val
					existing = true
					break
				}
			}
			if !existing {
				deviceAttributes = append(deviceAttributes, &pb.DeviceTypeAttributeRequest{
					Name:  k,
					Value: val,
				})
			}
			addedAttrCount++
		}
	}
	if addedAttrCount == 0 {
		deviceAttributes = nil
	}
	return deviceAttributes
}
