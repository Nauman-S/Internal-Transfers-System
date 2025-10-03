package account

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/Nauman-S/Internal-Transfers-System/models"
)

type CreateAccountRequest struct {
	AccountID      int    `json:"account_id" validate:"required,min=0"`
	InitialBalance string `json:"initial_balance" validate:"required,numeric"`
}

type CreateAccountResponse struct{}

type GetAccountResponse struct {
	AccountID int    `json:"account_id"`
	Balance   string `json:"balance"`
}

func (req *CreateAccountRequest) ToAccount() (*models.Account, error) {
	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		return nil, err
	}
	
	if balance.IsNegative() {
		return nil, codes.ErrNegativeBalance
	}
	
	now := time.Now()
	return &models.Account{
		ID:             req.AccountID,
		InitialBalance: balance,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
