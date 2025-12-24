package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID              string    `json:"id"`
	RBWID           string    `json:"rbw_id"`
	CategoryID      string    `json:"category_id"`
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	Description     *string   `json:"description,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       *string   `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TransactionType constants
const (
	TransactionTypeIncome  = "income"
	TransactionTypeExpense = "expense"
)

// TransactionCategory represents a transaction category
type TransactionCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateTransactionRequest for creating a transaction
type CreateTransactionRequest struct {
	RBWID           string    `json:"rbw_id"`
	CategoryID      string    `json:"category_id"`
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	Description     string    `json:"description,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
}

// UpdateTransactionRequest for updating a transaction
type UpdateTransactionRequest struct {
	CategoryID  *string  `json:"category_id,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Description *string  `json:"description,omitempty"`
}
