package config

import (
	"context"
	"github.com/Nauman-S/Internal-Transfers-System/storage"
)

type ApplicationConfig struct {
	Ctx                context.Context
	DB                 *storage.DB
	AccountRepository  *storage.AccountRepository
	TransferRepository *storage.TransferRepository
}