package domain

const OrderStatusRegistered = "REGISTERED"
const OrderStatusInvalid = "INVALID"
const OrderStatusProcessing = "PROCESSING"
const OrderStatusProcessed = "PROCESSED"

type OrderGoods struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Order struct {
	Number string        `json:"order"`
	Goods  []*OrderGoods `json:"goods"`
}

type OrderRecord struct {
	Number  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
