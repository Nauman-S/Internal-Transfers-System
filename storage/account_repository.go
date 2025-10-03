package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Nauman-S/Internal-Transfers-System/models"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{
		db: db.GetPool(),
	}
}

func (r *AccountRepository) CreateAccount(ctx context.Context, acc *models.Account) (bool, error) {
	query := `
		INSERT INTO accounts (id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(ctx, query,
		acc.ID,
		acc.InitialBalance,
		acc.CreatedAt,
		acc.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return false, nil
		}
		return false, fmt.Errorf("failed to create account: %w", err)
	}

	return true, nil
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, accountID int) (*models.Account, error) {
	query := `
		SELECT id, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1`

	var acc models.Account
	err := r.db.QueryRow(ctx, query, accountID).Scan(
		&acc.ID,
		&acc.InitialBalance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &acc, nil
}

func (r *AccountRepository) AccountExists(ctx context.Context, accountID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return exists, nil
}
