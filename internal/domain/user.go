package domain

import "context"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserService interface {
	Register(ctx context.Context, user *User, uniqueLoginErrorChan chan error) error
	CompareHashAndPassword(ctx context.Context, user *User) (bool, error)
	AddOrder(ctx context.Context, orderNumber int) error
}

type UserRepository interface {
	Register(ctx context.Context, user User, uniqueLoginErrorChan chan error) error
	GetPasswordHash(ctx context.Context, login string) (string, error)
	AddOrder(ctx context.Context, orderNumber int) error
}
