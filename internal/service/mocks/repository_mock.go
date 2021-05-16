// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ezhdanovskiy/wallets/internal/service (interfaces: Repository)

// Package mocks is a generated GoMock package.
package mocks

import (
	dto "github.com/ezhdanovskiy/wallets/internal/dto"
	gomock "github.com/golang/mock/gomock"
	sqlx "github.com/jmoiron/sqlx"
	reflect "reflect"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// CreateWallet mocks base method
func (m *MockRepository) CreateWallet(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWallet", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateWallet indicates an expected call of CreateWallet
func (mr *MockRepositoryMockRecorder) CreateWallet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWallet", reflect.TypeOf((*MockRepository)(nil).CreateWallet), arg0)
}

// GetOperations mocks base method
func (m *MockRepository) GetOperations(arg0 dto.OperationsFilter) ([]dto.Operation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOperations", arg0)
	ret0, _ := ret[0].([]dto.Operation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOperations indicates an expected call of GetOperations
func (mr *MockRepositoryMockRecorder) GetOperations(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOperations", reflect.TypeOf((*MockRepository)(nil).GetOperations), arg0)
}

// GetWallet mocks base method
func (m *MockRepository) GetWallet(arg0 string) (*dto.Wallet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWallet", arg0)
	ret0, _ := ret[0].(*dto.Wallet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWallet indicates an expected call of GetWallet
func (mr *MockRepositoryMockRecorder) GetWallet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWallet", reflect.TypeOf((*MockRepository)(nil).GetWallet), arg0)
}

// GetWalletsForUpdateTx mocks base method
func (m *MockRepository) GetWalletsForUpdateTx(arg0 *sqlx.Tx, arg1 []string) ([]dto.Wallet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWalletsForUpdateTx", arg0, arg1)
	ret0, _ := ret[0].([]dto.Wallet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWalletsForUpdateTx indicates an expected call of GetWalletsForUpdateTx
func (mr *MockRepositoryMockRecorder) GetWalletsForUpdateTx(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWalletsForUpdateTx", reflect.TypeOf((*MockRepository)(nil).GetWalletsForUpdateTx), arg0, arg1)
}

// IncreaseWalletBalance mocks base method
func (m *MockRepository) IncreaseWalletBalance(arg0 string, arg1 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IncreaseWalletBalance", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// IncreaseWalletBalance indicates an expected call of IncreaseWalletBalance
func (mr *MockRepositoryMockRecorder) IncreaseWalletBalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncreaseWalletBalance", reflect.TypeOf((*MockRepository)(nil).IncreaseWalletBalance), arg0, arg1)
}

// RunWithTransaction mocks base method
func (m *MockRepository) RunWithTransaction(arg0 func(*sqlx.Tx) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunWithTransaction", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunWithTransaction indicates an expected call of RunWithTransaction
func (mr *MockRepositoryMockRecorder) RunWithTransaction(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunWithTransaction", reflect.TypeOf((*MockRepository)(nil).RunWithTransaction), arg0)
}

// TransferTx mocks base method
func (m *MockRepository) TransferTx(arg0 *sqlx.Tx, arg1, arg2 string, arg3 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransferTx", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// TransferTx indicates an expected call of TransferTx
func (mr *MockRepositoryMockRecorder) TransferTx(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransferTx", reflect.TypeOf((*MockRepository)(nil).TransferTx), arg0, arg1, arg2, arg3)
}
