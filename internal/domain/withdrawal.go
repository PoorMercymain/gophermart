package domain

import "encoding/json"

type Withdrawal struct {
	OrderNumber      string     `json:"order"`
	WithdrawalAmount WithdrawalAmount `json:"sum"`
}

type WithdrawalAmount struct {
	Withdrawal int
}

func (w *WithdrawalAmount) UnmarshalJSON(data []byte) error {
	var withdrawalFloat float64

	if err := json.Unmarshal(data, &withdrawalFloat); err != nil {
		return err
	}

	w.Withdrawal = int(withdrawalFloat * 100)
	return nil
}
