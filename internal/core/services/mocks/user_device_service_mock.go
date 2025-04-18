// Code generated by MockGen. DO NOT EDIT.
// Source: user_device_service.go
//
// Generated by this command:
//
//	mockgen -source user_device_service.go -destination mocks/user_device_service_mock.go
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

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
func (m *MockUserDeviceAPIService) CanRequestInstantOffer(ctx context.Context, tokenID uint64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanRequestInstantOffer", ctx, tokenID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CanRequestInstantOffer indicates an expected call of CanRequestInstantOffer.
func (mr *MockUserDeviceAPIServiceMockRecorder) CanRequestInstantOffer(ctx, tokenID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanRequestInstantOffer", reflect.TypeOf((*MockUserDeviceAPIService)(nil).CanRequestInstantOffer), ctx, tokenID)
}

// GetOffers mocks base method.
func (m *MockUserDeviceAPIService) GetOffers(ctx context.Context, tokenID uint64) (*models.DeviceOffer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOffers", ctx, tokenID)
	ret0, _ := ret[0].(*models.DeviceOffer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOffers indicates an expected call of GetOffers.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetOffers(ctx, tokenID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOffers", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetOffers), ctx, tokenID)
}

// GetValuations mocks base method.
func (m *MockUserDeviceAPIService) GetValuations(ctx context.Context, tokenID uint64, privJWT string) (*models.DeviceValuation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValuations", ctx, tokenID, privJWT)
	ret0, _ := ret[0].(*models.DeviceValuation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValuations indicates an expected call of GetValuations.
func (mr *MockUserDeviceAPIServiceMockRecorder) GetValuations(ctx, tokenID, privJWT any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValuations", reflect.TypeOf((*MockUserDeviceAPIService)(nil).GetValuations), ctx, tokenID, privJWT)
}

// LastRequestDidGiveError mocks base method.
func (m *MockUserDeviceAPIService) LastRequestDidGiveError(ctx context.Context, tokenID uint64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastRequestDidGiveError", ctx, tokenID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastRequestDidGiveError indicates an expected call of LastRequestDidGiveError.
func (mr *MockUserDeviceAPIServiceMockRecorder) LastRequestDidGiveError(ctx, tokenID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastRequestDidGiveError", reflect.TypeOf((*MockUserDeviceAPIService)(nil).LastRequestDidGiveError), ctx, tokenID)
}
