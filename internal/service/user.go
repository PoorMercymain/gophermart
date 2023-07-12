package service

import (
	"context"

	"github.com/PoorMercymain/gophermart/internal/domain"
	"github.com/PoorMercymain/gophermart/pkg/util"
	"golang.org/x/crypto/bcrypt"
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
		util.LogInfoln(user.Password, err)
		return err
	}
	user.Password = string(passwordHash)
	util.LogInfoln("после хэширования", *user)
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

func (s *user) AddOrder(ctx context.Context, orderNumber int) error {
	//TODO: add Luhn check
	return s.repo.AddOrder(ctx, orderNumber)
}

func (s *user) ReadOrders(ctx context.Context) ([]domain.Order, error) {
	return s.repo.ReadOrders(ctx)
}

func (s *user) ReadBalance(ctx context.Context) (domain.Balance, error) {
	return s.repo.ReadBalance(ctx)
}
