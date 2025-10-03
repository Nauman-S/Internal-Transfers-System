package account

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/Nauman-S/Internal-Transfers-System/config"
	"github.com/Nauman-S/Internal-Transfers-System/storage"
	log "github.com/sirupsen/logrus"
)



func CreateAccount(c *gin.Context, req *CreateAccountRequest) (*CreateAccountResponse, error) {
	account, err := req.ToAccount()
	if err != nil {
		log.WithError(err).Error("Failure to parse create account request")
		return nil, err
	}

	log.WithFields(log.Fields{
		"account_id": account.ID,
		"balance":    account.InitialBalance.String(),
	}).Info("Attempting to create account")

	repo, err := getRepo(c)
	if err != nil {
		return nil, err
	}

	created, err := repo.CreateAccount(c.Request.Context(), account)
	if err != nil {
	    return nil, codes.NewWithMsg(codes.ErrSystem, "database error: %v", err)
	}

	if !created {
		log.WithField("account_id", account.ID).Warn("Account already exists")
		return nil, codes.ErrAccountExists
	}

	log.WithFields(log.Fields{
		"account_id": account.ID,
		"balance":    account.InitialBalance.String(),
	}).Info("Account created successfully")

	return &CreateAccountResponse{}, nil
}


func GetAccountByID(c *gin.Context) (*GetAccountResponse, error) {
	accountIDStr := c.Param("account_id")
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil || accountID <= 0 {
		log.WithError(err).WithField("account_id", accountIDStr).Error("Invalid account ID format")
		return nil, codes.ErrInvalidAccountID
	}

	log.WithFields(log.Fields{
		"account_id": accountID,
	}).Info("Attempting to get account")

	repo, err := getRepo(c)
	if err != nil {
		log.WithError(err).Error("Failed to get account repository from context")
		return nil, err
	}

	account, err := repo.GetAccountByID(c.Request.Context(), accountID)
	if err != nil {
		log.WithError(err).WithField("account_id", accountID).Error("Failed to get account from database")
		return nil, codes.NewWithMsg(codes.ErrSystem, "database error: %v", err)
	}

	if account == nil {
		log.WithField("account_id", accountID).Warn("Account not found")
		return nil, codes.ErrAccountNotFound
	}

	log.WithFields(log.Fields{
		"account_id": account.ID,
		"balance":    account.InitialBalance.String(),
	}).Info("Account retrieved successfully")

	return &GetAccountResponse{
		AccountID: account.ID,
		Balance:   account.InitialBalance.String(),
	}, nil
}

func getRepo(c *gin.Context) (*storage.AccountRepository, error) {
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
	
	if appConfig.AccountRepository == nil {
		log.Error("Account repository not found in app config")
		return nil, codes.NewWithMsg(codes.ErrSystem, "internal configuration error")
	}

	return appConfig.AccountRepository, nil
}