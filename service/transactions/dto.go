package transactions

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
)

type TransferRequest struct {
	SourceAccountID      int    `json:"source_account_id" validate:"required,min=1"`
	DestinationAccountID int    `json:"destination_account_id" validate:"required,min=1"`
	Amount              string `json:"amount" validate:"required,numeric,gt=0"`
}

// TransferResponse represents the response after processing a transfer
type TransferResponse struct {
	TransactionID       int    `json:"transaction_id"`
	Status             string `json:"status"`
	SourceBalance      string `json:"source_balance"`
	DestinationBalance string `json:"destination_balance"`
	Amount             string `json:"amount"`
	CreatedAt          string `json:"created_at"`
}

func (req *TransferRequest) ValidateRequest() error {
	if req.SourceAccountID == req.DestinationAccountID {
		return codes.ErrSameAccountTransfer
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return codes.NewWithMsg(codes.ErrInvalidParams, "invalid amount format: %v", err)
	}

	if amount.IsNegative() || amount.IsZero() {
		return codes.NewWithMsg(codes.ErrInvalidParams, "amount must be positive")
	}

	return nil
}

func (req *TransferRequest) ToResponse(transactionID int, sourceBalance, destBalance decimal.Decimal) *TransferResponse {
	return &TransferResponse{
		TransactionID:       transactionID,
		Status:             "COMPLETED",
		SourceBalance:      sourceBalance.String(),
		DestinationBalance: destBalance.String(),
		Amount:             req.Amount,
		CreatedAt:          time.Now().Format(time.RFC3339),
	}
}
