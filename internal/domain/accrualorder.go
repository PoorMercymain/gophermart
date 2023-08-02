package domain

import "encoding/json"

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual AccrualAmount `json:"accrual"`
}

type AccrualAmount struct {
	Accrual int
}

func (a *AccrualAmount) UnmarshalJSON(data []byte) error {
	var accrualFloat float64

	if err := json.Unmarshal(data, &accrualFloat); err != nil {
		return err
	}

	a.Accrual = int(accrualFloat * 100)
	return nil
}

type AccrualOrderWithUsername struct {
	Accrual AccrualOrder
	Username string
}
