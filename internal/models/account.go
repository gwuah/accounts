package models

import "time"

type Model struct {
	ID        int        `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Account struct {
	Model
	UserID        int    `json:"user_id"`
	AccountNumber string `json:"account_number"`
}

type Transaction struct {
	Model
	Reference string `json:"reference"`
}

type TransactionLine struct {
	Model
	TransactionID int `json:"transaction_id"`
	AccountID     int `json:"account_id"`
	Amount        int `json:"amount"`
}

type User struct {
	Model
	Email    string     `json:"email"`
	Accounts []*Account `json:"accounts"`
}
