package commands

import (
	"context"
	"github.com/nats-io/nats.go"
	"testing"
)

func Test_runValuationCommandHandler_processMessage(t *testing.T) {
	// need db
	// need mocks
	h := &runValuationCommandHandler{
		DBS:                      tt.fields.DBS,
		logger:                   tt.fields.logger,
		userDeviceService:        tt.fields.userDeviceService,
		NATSSvc:                  tt.fields.NATSSvc,
		vincarioValuationService: tt.fields.vincarioValuationService,
		drivlyValuationService:   tt.fields.drivlyValuationService,
	}
	type args struct {
		msg *nats.Msg
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := h.processMessage(context.Background(), tt.args.localLog, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("processMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
