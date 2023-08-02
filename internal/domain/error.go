package domain

import "errors"

var (
	ErrorAlreadyRegistered              = errors.New("already registred by the user")
	ErrorAlreadyRegisteredByAnotherUser = errors.New("already registered by another user")
	ErrorNotEnoughPoints                = errors.New("not enough points to withdraw")
	ErrorIncorrectOrderNumber           = errors.New("incorrect order number")
)
