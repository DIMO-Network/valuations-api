// Code generated by MockGen. DO NOT EDIT.
// Source: identity_api.go
//
// Generated by this command:
//
//	mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mock_gateways
//

// Package mock_gateways is a generated GoMock package.
package mock_gateways

import (
	reflect "reflect"

	models "github.com/DIMO-Network/valuations-api/internal/core/models"
	gomock "go.uber.org/mock/gomock"
)

// MockIdentityAPI is a mock of IdentityAPI interface.
type MockIdentityAPI struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityAPIMockRecorder
}

// MockIdentityAPIMockRecorder is the mock recorder for MockIdentityAPI.
type MockIdentityAPIMockRecorder struct {
	mock *MockIdentityAPI
}

// NewMockIdentityAPI creates a new mock instance.
func NewMockIdentityAPI(ctrl *gomock.Controller) *MockIdentityAPI {
	mock := &MockIdentityAPI{ctrl: ctrl}
	mock.recorder = &MockIdentityAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIdentityAPI) EXPECT() *MockIdentityAPIMockRecorder {
	return m.recorder
}

// GetDefinition mocks base method.
func (m *MockIdentityAPI) GetDefinition(definitionID string) (*models.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinition", definitionID)
	ret0, _ := ret[0].(*models.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDefinition indicates an expected call of GetDefinition.
func (mr *MockIdentityAPIMockRecorder) GetDefinition(definitionID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinition", reflect.TypeOf((*MockIdentityAPI)(nil).GetDefinition), definitionID)
}

// GetManufacturer mocks base method.
func (m *MockIdentityAPI) GetManufacturer(slug string) (*models.Manufacturer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManufacturer", slug)
	ret0, _ := ret[0].(*models.Manufacturer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManufacturer indicates an expected call of GetManufacturer.
func (mr *MockIdentityAPIMockRecorder) GetManufacturer(slug any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManufacturer", reflect.TypeOf((*MockIdentityAPI)(nil).GetManufacturer), slug)
}

// GetVehicle mocks base method.
func (m *MockIdentityAPI) GetVehicle(tokenID uint64) (*models.Vehicle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVehicle", tokenID)
	ret0, _ := ret[0].(*models.Vehicle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVehicle indicates an expected call of GetVehicle.
func (mr *MockIdentityAPIMockRecorder) GetVehicle(tokenID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVehicle", reflect.TypeOf((*MockIdentityAPI)(nil).GetVehicle), tokenID)
}
