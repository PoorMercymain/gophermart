// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PoorMercymain/gophermart/internal/domain (interfaces: UserRepository)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/PoorMercymain/gophermart/internal/domain"
	gomock "github.com/golang/mock/gomock"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// AddOrder mocks base method.
func (m *MockUserRepository) AddOrder(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddOrder indicates an expected call of AddOrder.
func (mr *MockUserRepositoryMockRecorder) AddOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrder", reflect.TypeOf((*MockUserRepository)(nil).AddOrder), arg0, arg1)
}

// AddWithdrawal mocks base method.
func (m *MockUserRepository) AddWithdrawal(arg0 context.Context, arg1 domain.Withdrawal) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddWithdrawal", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddWithdrawal indicates an expected call of AddWithdrawal.
func (mr *MockUserRepositoryMockRecorder) AddWithdrawal(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddWithdrawal", reflect.TypeOf((*MockUserRepository)(nil).AddWithdrawal), arg0, arg1)
}

// GetPasswordHash mocks base method.
func (m *MockUserRepository) GetPasswordHash(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPasswordHash", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPasswordHash indicates an expected call of GetPasswordHash.
func (mr *MockUserRepositoryMockRecorder) GetPasswordHash(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPasswordHash", reflect.TypeOf((*MockUserRepository)(nil).GetPasswordHash), arg0, arg1)
}

// GetUnprocessedBatch mocks base method.
func (m *MockUserRepository) GetUnprocessedBatch(arg0 context.Context, arg1 int) ([]domain.AccrualOrderWithUsername, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUnprocessedBatch", arg0, arg1)
	ret0, _ := ret[0].([]domain.AccrualOrderWithUsername)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUnprocessedBatch indicates an expected call of GetUnprocessedBatch.
func (mr *MockUserRepositoryMockRecorder) GetUnprocessedBatch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnprocessedBatch", reflect.TypeOf((*MockUserRepository)(nil).GetUnprocessedBatch), arg0, arg1)
}

// ReadBalance mocks base method.
func (m *MockUserRepository) ReadBalance(arg0 context.Context) (domain.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadBalance", arg0)
	ret0, _ := ret[0].(domain.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadBalance indicates an expected call of ReadBalance.
func (mr *MockUserRepositoryMockRecorder) ReadBalance(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadBalance", reflect.TypeOf((*MockUserRepository)(nil).ReadBalance), arg0)
}

// ReadOrders mocks base method.
func (m *MockUserRepository) ReadOrders(arg0 context.Context) ([]domain.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadOrders", arg0)
	ret0, _ := ret[0].([]domain.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadOrders indicates an expected call of ReadOrders.
func (mr *MockUserRepositoryMockRecorder) ReadOrders(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadOrders", reflect.TypeOf((*MockUserRepository)(nil).ReadOrders), arg0)
}

// Register mocks base method.
func (m *MockUserRepository) Register(arg0 context.Context, arg1 domain.User, arg2 chan error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockUserRepositoryMockRecorder) Register(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockUserRepository)(nil).Register), arg0, arg1, arg2)
}

// UpdateOrder mocks base method.
func (m *MockUserRepository) UpdateOrder(arg0 context.Context, arg1 domain.AccrualOrder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockUserRepositoryMockRecorder) UpdateOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockUserRepository)(nil).UpdateOrder), arg0, arg1)
}
