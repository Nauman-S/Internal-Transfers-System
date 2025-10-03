package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Account represents a bank account in the system
type Account struct {
	ID             int             `json:"account_id" db:"id"`
	InitialBalance decimal.Decimal `json:"initial_balance" db:"balance"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}
