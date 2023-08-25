// Code generated by MockGen. DO NOT EDIT.
// Source: drivly_valuation_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/valuations-api/internal/core/services"
	gomock "github.com/golang/mock/gomock"
)

// MockDrivlyValuationService is a mock of DrivlyValuationService interface.
type MockDrivlyValuationService struct {
	ctrl     *gomock.Controller
	recorder *MockDrivlyValuationServiceMockRecorder
}

// MockDrivlyValuationServiceMockRecorder is the mock recorder for MockDrivlyValuationService.
type MockDrivlyValuationServiceMockRecorder struct {
	mock *MockDrivlyValuationService
}

// NewMockDrivlyValuationService creates a new mock instance.
func NewMockDrivlyValuationService(ctrl *gomock.Controller) *MockDrivlyValuationService {
	mock := &MockDrivlyValuationService{ctrl: ctrl}
	mock.recorder = &MockDrivlyValuationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDrivlyValuationService) EXPECT() *MockDrivlyValuationServiceMockRecorder {
	return m.recorder
}

// PullValuation mocks base method.
func (m *MockDrivlyValuationService) PullValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (services.DataPullStatusEnum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullValuation", ctx, userDeiceID, deviceDefinitionID, vin)
	ret0, _ := ret[0].(services.DataPullStatusEnum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PullValuation indicates an expected call of PullValuation.
func (mr *MockDrivlyValuationServiceMockRecorder) PullValuation(ctx, userDeiceID, deviceDefinitionID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullValuation", reflect.TypeOf((*MockDrivlyValuationService)(nil).PullValuation), ctx, userDeiceID, deviceDefinitionID, vin)
}