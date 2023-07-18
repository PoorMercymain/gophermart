package domain

import "context"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserService interface {
	Register(ctx context.Context, user *User, uniqueLoginErrorChan chan error) error
	CompareHashAndPassword(ctx context.Context, user *User) (bool, error)
	AddOrder(ctx context.Context, orderNumber string) error
	ReadOrders(ctx context.Context) ([]Order, error)
	ReadBalance(ctx context.Context) (Balance, error)
	AddWithdrawal(ctx context.Context, withdrawal Withdrawal) error
}

//go:generate mockgen -destination=mocks/repo_mock.gen.go -package=mocks . UserRepository
type UserRepository interface {
	Register(ctx context.Context, user User, uniqueLoginErrorChan chan error) error
	GetPasswordHash(ctx context.Context, login string) (string, error)
	AddOrder(ctx context.Context, orderNumber string) error
	ReadOrders(ctx context.Context) ([]Order, error)
	ReadBalance(ctx context.Context) (Balance, error)
	AddWithdrawal(ctx context.Context, withdrawal Withdrawal) error
}
