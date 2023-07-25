package domain

type AccrualOrder struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual string `json:"accrual"`
}
