package domain

type Withdrawal struct {
	OrderNumber      int64 `json:"order"`
	WithdrawalAmount int   `json:"sum"`
}
