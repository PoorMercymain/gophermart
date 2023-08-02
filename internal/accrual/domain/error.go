package domain

import "errors"

var (
	ErrorOrderNotRegistered     = errors.New("order not registered")
	ErrorOrderAlreadyProcessing = errors.New("order already processing")
	ErrorRequestsLimitExceeded  = errors.New("requests limit exceeded")
	ErrorRequestFormatIncorrect = errors.New("incorrect request format")
	ErrorMatchAlreadyRegistered = errors.New("match already registered")
)
