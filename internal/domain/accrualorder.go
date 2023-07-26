package domain

import "encoding/json"

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual AccrualAmount `json:"accrual"`
}

type AccrualAmount int

func (a *AccrualAmount) UnmarshalJSON(data []byte) error {
	var accrualFloat float64

	if err := json.Unmarshal(data, &accrualFloat); err != nil {
		return err
	}

	*a = AccrualAmount(int(accrualFloat * 100))
	return nil
}
