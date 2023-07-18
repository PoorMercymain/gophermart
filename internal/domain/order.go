package domain

import "time"

type Order struct {
	Number           string    `json:"number"`
	Status           string    `json:"status"`
	Accrual          int       `json:"accrual,omitempty"`
	UploadedAt       time.Time `json:"-"`
	UploadedAtString string    `json:"uploaded_at"`
}
