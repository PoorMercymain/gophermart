package service

import (
	"context"

	"github.com/ShiraazMoollatjie/goluhn"
	"golang.org/x/crypto/bcrypt"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
)

type user struct {
	repo domain.UserRepository
}

func NewUser(repo domain.UserRepository) *user {
	return &user{repo: repo}
}

func (s *user) Register(ctx context.Context, user *domain.User, uniqueLoginErrorChan chan error) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		util.GetLogger().Infoln(user.Password, err)
		return err
	}
	user.Password = string(passwordHash)
	util.GetLogger().Infoln("после хэширования", *user)
	return s.repo.Register(ctx, *user, uniqueLoginErrorChan)
}

func (s *user) CompareHashAndPassword(ctx context.Context, user *domain.User) (bool, error) {
	hash, err := s.repo.GetPasswordHash(ctx, user.Login)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))
	if err != nil {
		return false, err
	}

	user.Password = hash
	return true, nil
}

func (s *user) AddOrder(ctx context.Context, orderNumber string) error {
	err := goluhn.Validate(orderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		return domain.ErrorIncorrectOrderNumber
	}
	return s.repo.AddOrder(ctx, orderNumber)
}

func (s *user) ReadOrders(ctx context.Context) ([]domain.Order, error) {
	return s.repo.ReadOrders(ctx)
}

func (s *user) ReadBalance(ctx context.Context) (domain.Balance, error) {
	return s.repo.ReadBalance(ctx)
}

func (s *user) AddWithdrawal(ctx context.Context, withdrawal domain.Withdrawal) error {
	err := goluhn.Validate(withdrawal.OrderNumber)
	if err != nil {
		util.GetLogger().Infoln(err)
		return domain.ErrorIncorrectOrderNumber
	}
	return s.repo.AddWithdrawal(ctx, withdrawal)
}

func (s *user) UpdateOrder(ctx context.Context, order domain.AccrualOrder) error {
	return s.repo.UpdateOrder(ctx, order)
}

func (s *user) GetUnprocessedBatch(ctx context.Context, batchNumber int) ([]domain.AccrualOrderWithUsername, error) {
	return s.repo.GetUnprocessedBatch(ctx, batchNumber)
}

func (s *user) ReadWithdrawals(ctx context.Context) ([]domain.WithdrawalOutput, error) {
	return s.repo.ReadWithdrawals(ctx)
}
