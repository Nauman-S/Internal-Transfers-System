package transactions

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/Nauman-S/Internal-Transfers-System/config"
	"github.com/Nauman-S/Internal-Transfers-System/storage"
	log "github.com/sirupsen/logrus"
)

func CreateTransfer(c *gin.Context, req *TransferRequest) (*TransferResponse, error) {

	err := req.ValidateRequest()
	if err != nil {
		log.WithError(err).Error("Transfer request validation failed")
		return nil, err
	}

	repo, err := getTransferRepo(c)
	if err != nil {
		log.WithError(err).Error("Failed to get transfer repository from context")
		return nil, err
	}


	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		log.WithError(err).Error("Failed to parse transfer amount")
		return nil, codes.NewWithMsg(codes.ErrInvalidParams, "invalid amount format")
	}

	log.WithFields(log.Fields{
		"source_account_id":      req.SourceAccountID,
		"destination_account_id": req.DestinationAccountID,
		"amount":                amount.String(),
	}).Info("Processing transfer request")

	transfer, sourceBalance, destBalance, err := repo.ProcessTransfer(c.Request.Context(), req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"source_account_id":      req.SourceAccountID,
			"destination_account_id": req.DestinationAccountID,
			"amount":                amount.String(),
		}).Error("Transfer processing failed")
		return nil, err
	}

	log.WithFields(log.Fields{
		"transaction_id":         transfer.ID,
		"source_account_id":      req.SourceAccountID,
		"destination_account_id": req.DestinationAccountID,
		"amount":                amount.String(),
		"source_balance":        sourceBalance.String(),
		"destination_balance":   destBalance.String(),
	}).Info("Transfer completed successfully")

	return req.ToResponse(transfer.ID, sourceBalance, destBalance), nil
}

func getTransferRepo(c *gin.Context) (*storage.TransferRepository, error) {
	appConfigInterface, exists := c.Get("appConfig")
	if !exists {
		log.Error("App config not found in context")
		return nil, codes.NewWithMsg(codes.ErrSystem, "internal configuration error")
	}
	
	appConfig, ok := appConfigInterface.(*config.ApplicationConfig)
	if !ok {
		log.Error("Invalid app config type in context")
		return nil, codes.NewWithMsg(codes.ErrSystem, "internal configuration error")
	}
	
	if appConfig.TransferRepository == nil {
		log.Error("Transfer repository not found in app config")
		return nil, codes.NewWithMsg(codes.ErrSystem, "internal configuration error")
	}

	return appConfig.TransferRepository, nil
}



