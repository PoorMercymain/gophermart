package domain

import "encoding/json"

type AccrualOrder struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual accrual `json:"accrual"`
}

type accrual int

func (a *accrual) UnmarshalJSON(data []byte) error {
	var accrualFloat float64

	if err := json.Unmarshal(data, &accrualFloat); err != nil {
		return err
	}

	*a = accrual(int(accrualFloat * 100))
	return nil
}
