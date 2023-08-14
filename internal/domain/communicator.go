package domain

import "net/http"

//go:generate mockgen -destination=mocks/communicator_mock.gen.go -package=mocks . Communicator
type Communicator interface {
	GetOrderAccrual(orderNumber string) (*http.Response, error)
}

type accrualCommunicator struct {
	accrualURL string
}

func NewAccrualCommunicator(accrualURL string) *accrualCommunicator {
	return &accrualCommunicator{accrualURL: accrualURL}
}

func (ac *accrualCommunicator) GetOrderAccrual(orderNumber string) (*http.Response, error) {
	return http.Get(ac.accrualURL + "/api/orders/" + orderNumber)
}
