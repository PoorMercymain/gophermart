package domain

import (
	"fmt"
)

type Accrual struct {
	Money int
}

func (a *Accrual) MarshalJSON() ([]byte, error) {
	moneyBeforePoint, moneyAfterPoint := getBeforeAndAfterPoint(a.Money)
	accrualString := fmt.Sprintf("%d", moneyBeforePoint)
	if moneyAfterPoint > 0 {
		if moneyAfterPoint > 9 {
			accrualString += fmt.Sprintf(".%d", moneyAfterPoint)
		} else {
			accrualString += fmt.Sprintf(".0%d", moneyAfterPoint)
		}
	}
	return []byte(accrualString), nil
}

type Balance struct {
	Balance   int
	Withdrawn int
}

func (b *Balance) Marshal() []byte {
	balance, balanceAfterPoint := getBeforeAndAfterPoint(b.Balance)
	withdrawn, withdrawnAfterPoint := getBeforeAndAfterPoint(b.Withdrawn)

	balanceString := fmt.Sprintf("{\"current\": %d", balance)
	if balanceAfterPoint > 0 {
		if balanceAfterPoint > 9 {
			balanceString += fmt.Sprintf(".%d", balanceAfterPoint)
		} else {
			balanceString += fmt.Sprintf(".0%d", balanceAfterPoint)
		}
	}

	balanceString += fmt.Sprintf(",\"withdrawn\": %d", withdrawn)
	if withdrawnAfterPoint > 0 {
		if withdrawnAfterPoint > 9 {
			balanceString += fmt.Sprintf(".%d", withdrawnAfterPoint)
		} else {
			balanceString += fmt.Sprintf(".0%d", withdrawnAfterPoint)
		}
	}
	balanceString += "}"

	return []byte(balanceString)
}

func getBeforeAndAfterPoint(amount int) (int, int) {
	amountAfterPoint := amount % 100
	amount -= amountAfterPoint
	amount = amount / 100
	return amount, amountAfterPoint
}
