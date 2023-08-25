// Code generated by MockGen. DO NOT EDIT.
// Source: vincario_valuation_service.go

// Package mock_services is a generated GoMock package.
package mock_services

import (
	context "context"
	reflect "reflect"

	services "github.com/DIMO-Network/valuations-api/internal/core/services"
	gomock "github.com/golang/mock/gomock"
)

// MockVincarioValuationService is a mock of VincarioValuationService interface.
type MockVincarioValuationService struct {
	ctrl     *gomock.Controller
	recorder *MockVincarioValuationServiceMockRecorder
}

// MockVincarioValuationServiceMockRecorder is the mock recorder for MockVincarioValuationService.
type MockVincarioValuationServiceMockRecorder struct {
	mock *MockVincarioValuationService
}

// NewMockVincarioValuationService creates a new mock instance.
func NewMockVincarioValuationService(ctrl *gomock.Controller) *MockVincarioValuationService {
	mock := &MockVincarioValuationService{ctrl: ctrl}
	mock.recorder = &MockVincarioValuationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVincarioValuationService) EXPECT() *MockVincarioValuationServiceMockRecorder {
	return m.recorder
}

// PullValuation mocks base method.
func (m *MockVincarioValuationService) PullValuation(ctx context.Context, userDeiceID, deviceDefinitionID, vin string) (services.DataPullStatusEnum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullValuation", ctx, userDeiceID, deviceDefinitionID, vin)
	ret0, _ := ret[0].(services.DataPullStatusEnum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PullValuation indicates an expected call of PullValuation.
func (mr *MockVincarioValuationServiceMockRecorder) PullValuation(ctx, userDeiceID, deviceDefinitionID, vin interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullValuation", reflect.TypeOf((*MockVincarioValuationService)(nil).PullValuation), ctx, userDeiceID, deviceDefinitionID, vin)
}