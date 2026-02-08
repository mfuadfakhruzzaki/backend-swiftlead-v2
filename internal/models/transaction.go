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

// CreateCategoryRequest for creating a transaction category (admin)
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=income expense"`
	Description string `json:"description,omitempty"`
}

// UpdateCategoryRequest for updating a transaction category (admin)
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// FinancialSummary represents an aggregated financial summary
type FinancialSummary struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	Balance      float64 `json:"balance"`
}

// FinancialStatement represents a financial statement for a period
type FinancialStatement struct {
	RBWID        string         `json:"rbw_id"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	TotalIncome  float64        `json:"total_income"`
	TotalExpense float64        `json:"total_expense"`
	Balance      float64        `json:"balance"`
	Incomes      []*Transaction `json:"incomes"`
	Expenses     []*Transaction `json:"expenses"`
}

// GenerateStatementRequest for generating a financial statement
type GenerateStatementRequest struct {
	RBWID     string `json:"rbw_id" validate:"required"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}
