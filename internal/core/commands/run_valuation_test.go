package commands

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	mock_services "github.com/DIMO-Network/valuations-api/internal/core/services/mocks"
	"github.com/DIMO-Network/valuations-api/internal/infrastructure/dbtest"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
)

const migrationsDirRelPath = "../../infrastructure/db/migrations"

func Test_runValuationCommandHandler_processMessage(t *testing.T) {
	// need db
	// need mocks
	ctx := context.Background()
	pdb, _ := dbtest.StartContainerDatabase(ctx, "valuations_api", t, migrationsDirRelPath)
	logger := dbtest.Logger()
	mockCtrl := gomock.NewController(t)

	userDeviceSvc := mock_services.NewMockUserDeviceAPIService(mockCtrl)
	vincarioSvc := mock_services.NewMockVincarioValuationService(mockCtrl)
	drivlySvc := mock_services.NewMockDrivlyValuationService(mockCtrl)
	userDeviceID := ksuid.New().String()
	ddID := ksuid.New().String()
	vin := "VINDIESEL12312322"

	h := &runValuationCommandHandler{
		DBS:                      pdb.DBS,
		logger:                   *logger,
		userDeviceService:        userDeviceSvc,
		NATSSvc:                  nil,
		vincarioValuationService: vincarioSvc,
		drivlyValuationService:   drivlySvc,
	}
	type args struct {
		msgBody string
		setup   func()
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{msgBody: "",
				setup: func() {
					userDeviceSvc.EXPECT().GetUserDevice(gomock.Any(), userDeviceID).Times(1).Return(&pb.UserDevice{Id: userDeviceID, UserId: "123", CountryCode: "USA", DeviceDefinitionId: ddID}, nil)
					drivlySvc.EXPECT().PullValuation(gomock.Any(), userDeviceID, ddID, vin)
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()

			msg := nats.NewMsg("dd_valuation_tasks")
			msg.Data = []byte(fmt.Sprintf(`{"vin": "%s", "userDeviceId": "%s" }`, vin, userDeviceID))
			msg.Reply = ""

			if err := h.processMessage(ctx, *logger, msg); (err != nil) != tt.wantErr {
				t.Errorf("processMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			// could add database expectations
		})
	}
}
