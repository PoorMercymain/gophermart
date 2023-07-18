package domain

type Withdrawal struct {
	OrderNumber      string `json:"order"`
	WithdrawalAmount int    `json:"sum"`
}
