package models

type Type string

var (
	Credit Type = "credit"
	Debit  Type = "debit"
)

type Transaction struct {
	Model
	From        int    `json:"from"`
	To          int    `json:"to"`
	Type        Type   `json:"type"`
	Description string `json:"description"`
	Reference   string `json:"reference"`
	AccountID   int    `json:"account_id"`
	Amount      int64  `json:"amount"`
}
