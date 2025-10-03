package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/Nauman-S/Internal-Transfers-System/models"
	"github.com/shopspring/decimal"
)

type TransferRepository struct {
	db *pgxpool.Pool
}

func NewTransferRepository(db *DB) *TransferRepository {
	return &TransferRepository{
		db: db.pool,
	}
}


func (r *TransferRepository) ProcessTransfer(ctx context.Context, sourceAccountID, destAccountID int, amount decimal.Decimal) (*models.Transfer, decimal.Decimal, decimal.Decimal, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var sourceBalance, destBalance decimal.Decimal
	
	//This is to prevent deadlocks
	if sourceAccountID < destAccountID {

		err = tx.QueryRow(ctx, `
			SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
		`, sourceAccountID).Scan(&sourceBalance)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, decimal.Zero, decimal.Zero, codes.ErrSourceAccountNotFound
			}
			return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to lock source account: %w", err)
		}

		err = tx.QueryRow(ctx, `
			SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
		`, destAccountID).Scan(&destBalance)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, decimal.Zero, decimal.Zero, codes.ErrDestinationAccountNotFound
			}
			return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to lock source account: %w", err)
		}
	} else {
		err = tx.QueryRow(ctx, `
			SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
		`, destAccountID).Scan(&destBalance)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, decimal.Zero, decimal.Zero, codes.ErrDestinationAccountNotFound
			}
			return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to lock source account: %w", err)
		}

		err = tx.QueryRow(ctx, `
			SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
		`, sourceAccountID).Scan(&sourceBalance)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, decimal.Zero, decimal.Zero, codes.ErrSourceAccountNotFound
			}
			return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to lock source account: %w", err)
		}
	}

	if sourceBalance.LessThan(amount) {
		return nil, decimal.Zero, decimal.Zero, codes.ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, `
		UPDATE accounts SET balance = balance - $1, updated_at = NOW() WHERE id = $2
	`, amount, sourceAccountID)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to debit source account: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE id = $2
	`, amount, destAccountID)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to credit destination account: %w", err)
	}

	var transferID int
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (source_account_id, destination_account_id, amount, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, sourceAccountID, destAccountID, amount).Scan(&transferID)
	if err != nil {
		return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to create transfer record: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, decimal.Zero, decimal.Zero, fmt.Errorf("failed to commit transaction: %w", err)
	}

	newSourceBalance := sourceBalance.Sub(amount)
	newDestBalance := destBalance.Add(amount)

	transfer := &models.Transfer{
		ID:                  transferID,
		SourceAccountID:     sourceAccountID,
		DestinationAccountID: destAccountID,
		Amount:              amount,
	}

	return transfer, newSourceBalance, newDestBalance, nil
}
