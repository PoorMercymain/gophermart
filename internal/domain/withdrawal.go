package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Withdrawal struct {
	OrderNumber      string     `json:"order"`
	WithdrawalAmount WithdrawalAmount `json:"sum"`
}

type WithdrawalAmount struct {
	Withdrawal int
}

type WithdrawalOutput struct {
	OrderNumber string `json:"order"`
	WithdrawnPoints WithdrawalAmount `json:"sum"`
	ProcessedAt time.Time `json:"-"`
	ProcessedAtString string `json:"processed_at"`
}

func (w *WithdrawalAmount) UnmarshalJSON(data []byte) error {
	var withdrawalFloat float64

	if err := json.Unmarshal(data, &withdrawalFloat); err != nil {
		return err
	}

	w.Withdrawal = int(withdrawalFloat * 100)
	return nil
}

func (w *WithdrawalAmount) MarshalJSON() ([]byte, error) {
	pointsBeforePoint, pointsAfterPoint := getBeforeAndAfterPoint(w.Withdrawal)
	withdrawalString := fmt.Sprintf("%d", pointsBeforePoint)
	if pointsAfterPoint > 0 {
		if pointsAfterPoint > 9 {
			withdrawalString += fmt.Sprintf(".%d", pointsAfterPoint)
		} else {
			withdrawalString += fmt.Sprintf(".0%d", pointsAfterPoint)
		}
	}
	return []byte(withdrawalString), nil
}
