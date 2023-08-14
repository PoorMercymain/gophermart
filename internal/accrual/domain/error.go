package domain

import (
	"errors"
	"time"
)

var (
	ErrorOrderNotRegistered     = errors.New("order not registered")
	ErrorOrderAlreadyProcessing = errors.New("order already processing")
	ErrorRequestsLimitExceeded  = errors.New("requests limit exceeded")
	ErrorRequestFormatIncorrect = errors.New("incorrect request format")
	ErrorMatchAlreadyRegistered = errors.New("match already registered")
)

var RepeatedAttemptsIntervals = [7]*time.Duration{
	getTimeInterval(1),
	getTimeInterval(3),
	getTimeInterval(5),
	getTimeInterval(15),
	getTimeInterval(25),
	getTimeInterval(35),
	getTimeInterval(45),
}

func getTimeInterval(seconds float64) *time.Duration {
	value := time.Duration(seconds) * time.Second
	return &value
}
