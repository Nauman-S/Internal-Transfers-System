package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Transfer represents a completed transfer transaction in the system
type Transfer struct {
	ID                  int             `json:"id" db:"id"`
	SourceAccountID     int             `json:"source_account_id" db:"source_account_id"`
	DestinationAccountID int            `json:"destination_account_id" db:"destination_account_id"`
	Amount              decimal.Decimal `json:"amount" db:"amount"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}
