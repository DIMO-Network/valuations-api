// Code generated by MockGen. DO NOT EDIT.
// Source: user_device_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	grpc "github.com/DIMO-Network/devices-api/pkg/grpc"
	models "github.com/DIMO-Network/valuations-api/internal/core/models"
	gomock "go.uber.org/mock/gomock"
)

// MockUserDeviceAPIService is a mock of UserDeviceAPIService interface.
type MockUserDeviceAPIService struct {
	ctrl     *gomock.Controller
	recorder *MockUserDeviceAPIServiceMockRecorder
}

// MockUserDeviceAPIServiceMockRecorder is the mock recorder for MockUserDeviceAPIService.
type MockUserDeviceAPIServiceMockRecorder struct {
	mock *MockUserDeviceAPIService
}

// NewMockUserDeviceAPIService creates a new mock instance.
func NewMockUserDeviceAPIService(ctrl *gomock.Controller) *MockUserDeviceAPIService {
	mock := &MockUserDeviceAPIService{ctrl: ctrl}
	mock.recorder = &MockUserDeviceAPIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserDeviceAPIService) EXPECT() *MockUserDeviceAPIServiceMockRecorder {
	return m.recorder
}

// CanRequestInstantOffer mocks base method.
func (m *MockUserDeviceAPIService) CanRequestInstantOffer(ctx context.Context, userDeviceID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanRequestInstantOffer", ctx, userDeviceID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CanRequestInstantOffer indicates an expected call of CanRequestInstantOffer.
func (mr *MockUserDeviceAPIServiceMockRecorder) CanRequestInstantOffer(ctx, userDeviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanRequestInstantOffer", reflect.TypeOf((*MockUserDeviceAPIService)(nil).CanRequestInstantOffer), ctx, userDeviceID)
}

// GetAllUserDevice mocks base method.
func (m *MockUserDeviceAPIService) GetAllUserDevice(ctx context.Context, wmi string) ([]*grpc.UserDevice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllUserDevice", ctx, wmi)
	ret0, _ := ret[0].([]*grpc.UserDevice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllUserDevice indicates an expected call of GetAllUserDevice.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetAllUserDevice(ctx, wmi interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllUserDevice", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetAllUserDevice), ctx, wmi)
}

// GetUserDevice mocks base method.
func (m *MockUserDeviceAPIService) GetUserDevice(ctx context.Context, userDeviceID string) (*grpc.UserDevice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDevice", ctx, userDeviceID)
	ret0, _ := ret[0].(*grpc.UserDevice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserDevice indicates an expected call of GetUserDevice.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetUserDevice(ctx, userDeviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDevice", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetUserDevice), ctx, userDeviceID)
}

// GetUserDeviceOffers mocks base method.
func (m *MockUserDeviceAPIService) GetUserDeviceOffers(ctx context.Context, userDeviceID string) (*models.DeviceOffer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDeviceOffers", ctx, userDeviceID)
	ret0, _ := ret[0].(*models.DeviceOffer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserDeviceOffers indicates an expected call of GetUserDeviceOffers.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetUserDeviceOffers(ctx, userDeviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDeviceOffers", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetUserDeviceOffers), ctx, userDeviceID)
}

// GetUserDeviceValuations mocks base method.
func (m *MockUserDeviceAPIService) GetUserDeviceValuations(ctx context.Context, userDeviceID, countryCode string) (*models.DeviceValuation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserDeviceValuations", ctx, userDeviceID, countryCode)
	ret0, _ := ret[0].(*models.DeviceValuation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserDeviceValuations indicates an expected call of GetUserDeviceValuations.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetUserDeviceValuations(ctx, userDeviceID, countryCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserDeviceValuations", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetUserDeviceValuations), ctx, userDeviceID, countryCode)
}

// LastRequestDidGiveError mocks base method.
func (m *MockUserDeviceAPIService) LastRequestDidGiveError(ctx context.Context, userDeviceID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastRequestDidGiveError", ctx, userDeviceID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastRequestDidGiveError indicates an expected call of LastRequestDidGiveError.
func (mr *MockUserDeviceAPIServiceMockRecorder) LastRequestDidGiveError(ctx, userDeviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastRequestDidGiveError", reflect.TypeOf((*MockUserDeviceAPIService)(nil).LastRequestDidGiveError), ctx, userDeviceID)
}

// UpdateUserDeviceMetadata mocks base method.
func (m *MockUserDeviceAPIService) UpdateUserDeviceMetadata(ctx context.Context, request *grpc.UpdateUserDeviceMetadataRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserDeviceMetadata", ctx, request)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserDeviceMetadata indicates an expected call of UpdateUserDeviceMetadata.
func (mr *MockUserDeviceAPIServiceMockRecorder) UpdateUserDeviceMetadata(ctx, request interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserDeviceMetadata", reflect.TypeOf((*MockUserDeviceAPIService)(nil).UpdateUserDeviceMetadata), ctx, request)
}
