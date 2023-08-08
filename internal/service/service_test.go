package service

import (
	"context"
	"testing"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/internal/domain/mocks"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestService(t *testing.T) {
	util.InitLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	us := NewUser(mockRepo)

	testHash, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	require.NoError(t, err)
	testHashStr := string(testHash)

	mockRepo.EXPECT().AddOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetPasswordHash(gomock.Any(), gomock.Any()).Return(testHashStr, nil).AnyTimes()
	mockRepo.EXPECT().GetUnprocessedBatch(gomock.Any(), gomock.Any()).Return(make([]domain.AccrualOrderWithUsername, 0), nil).AnyTimes()
	mockRepo.EXPECT().ReadBalance(gomock.Any()).Return(domain.Balance{Balance: 100, Withdrawn: 10}, nil).AnyTimes()
	mockRepo.EXPECT().ReadOrders(gomock.Any()).Return(make([]domain.Order, 0), nil).AnyTimes()
	mockRepo.EXPECT().ReadWithdrawals(gomock.Any()).Return(make([]domain.WithdrawalOutput, 0), nil).AnyTimes()
	mockRepo.EXPECT().Register(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().UpdateOrder(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	util.GetLogger().Infoln("started add order")
	err = us.AddOrder(context.Background(), "12345")
	require.Error(t, err)
	err = us.AddOrder(context.Background(), "573956")
	require.NoError(t, err)
	err = us.AddOrder(context.Background(), "ab")
	require.Error(t, err)

	util.GetLogger().Infoln("started add withdrawal")
	err = us.AddWithdrawal(context.Background(), domain.Withdrawal{WithdrawalAmount: domain.WithdrawalAmount{Withdrawal: 12345}})
	require.NoError(t, err)

	util.GetLogger().Infoln("started compare")
	isEqual, err := us.CompareHashAndPassword(context.Background(), &domain.User{Login: "test", Password: "test"})
	require.NoError(t, err)
	require.Equal(t, true, isEqual)
	isEqual, err = us.CompareHashAndPassword(context.Background(), &domain.User{Login: "test", Password: "testing"})
	require.Error(t, err)
	require.Equal(t, false, isEqual)

	util.GetLogger().Infoln("started unprocessed batch")
	testAccrualWithUsername, err := us.GetUnprocessedBatch(context.Background(), 0)
	require.Len(t, testAccrualWithUsername, 0)
	require.NoError(t, err)

	testBalance, err := us.ReadBalance(context.Background())
	require.NoError(t, err)
	require.Equal(t, 100, testBalance.Balance)
	require.Equal(t, 10, testBalance.Withdrawn)

	testOrders, err := us.ReadOrders(context.Background())
	require.NoError(t, err)
	require.Len(t, testOrders, 0)

	testWithdrawalOutput, err := us.ReadWithdrawals(context.Background())
	require.NoError(t, err)
	require.Len(t, testWithdrawalOutput, 0)

	us.Register(context.Background(), &domain.User{Login: "test", Password: "test"}, make(chan error))
	require.NoError(t, err)

	testAccrualAmount := domain.AccrualAmount{Accrual: 100}
	err = us.UpdateOrder(context.Background(), domain.AccrualOrder{Order: "573956", Status: "PROCESSED", Accrual: testAccrualAmount})
	require.NoError(t, err)
}
