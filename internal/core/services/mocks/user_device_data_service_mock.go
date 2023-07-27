// Code generated by MockGen. DO NOT EDIT.
// Source: user_device_data_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	grpc "github.com/DIMO-Network/device-data-api/pkg/grpc"
	gomock "github.com/golang/mock/gomock"
)

// MockUserDeviceDataAPIService is a mock of UserDeviceDataAPIService interface.
type MockUserDeviceDataAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockUserDeviceDataAPIServiceMockRecorder
}

// MockUserDeviceDataAPIServiceMockRecorder is the mock recorder for MockUserDeviceDataAPIService.
type MockUserDeviceDataAPIServiceMockRecorder struct {
	mock *MockUserDeviceDataAPIService
}

// NewMockUserDeviceDataAPIService creates a new mock instance.
func NewMockUserDeviceDataAPIService(ctrl *gomock.Controller) *MockUserDeviceDataAPIService {
	mock := &MockUserDeviceDataAPIService{ctrl: ctrl}
	mock.recorder = &MockUserDeviceDataAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserDeviceDataAPIService) EXPECT() *MockUserDeviceDataAPIServiceMockRecorder {
	return m.recorder
}

// GetUserDeviceData mocks base method.
func (m *MockUserDeviceDataAPIService) GetUserDeviceData(ctx context.Context, id, ddID string) (*grpc.UserDeviceDataResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDeviceData", ctx, id, ddID)
	ret0, _ := ret[0].(*grpc.UserDeviceDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserDeviceData indicates an expected call of GetUserDeviceData.
func (mr *MockUserDeviceDataAPIServiceMockRecorder) GetUserDeviceData(ctx, id, ddID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDeviceData", reflect.TypeOf((*MockUserDeviceDataAPIService)(nil).GetUserDeviceData), ctx, id, ddID)
}
