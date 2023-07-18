package domain

import "errors"

var (
	ErrorAlreadyRegistered              = errors.New("already registred by the user")
	ErrorAlreadyRegisteredByAnotherUser = errors.New("already registered by another user")
)
