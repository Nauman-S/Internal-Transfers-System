package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/timeout"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/Nauman-S/Internal-Transfers-System/config"
	"github.com/Nauman-S/Internal-Transfers-System/service/account"
	"github.com/Nauman-S/Internal-Transfers-System/service/transactions"
	"github.com/Nauman-S/Internal-Transfers-System/rest_handler"
)

func InitRouter(appConfig *config.ApplicationConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(timeoutHF(5 * time.Second))
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	
	r.Use(func(c *gin.Context) {
		c.Set("appConfig", appConfig)
		c.Next()
	})

	handler := rest_handler.NewHandler(
		[]rest_handler.FrontFilter{},
		[]rest_handler.RequestFilter{},
	)

	accountsAPI := r.Group("/accounts")
	{
		accountsAPI.POST("/", handler.HandleMiddleware(account.CreateAccount))
		accountsAPI.GET("/:account_id", handler.HandleMiddleware(account.GetAccountByID))
	}

	transactionsAPI := r.Group("/transactions")
	{
		transactionsAPI.POST("/", handler.HandleMiddleware(transactions.CreateTransfer))
	}


	return r
}

func timeoutHF(ttl time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(ttl),
		timeout.WithResponse(func(c *gin.Context) {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"code": codes.ErrTimeout.Code,
				"msg":  codes.ErrTimeout.Msg,
			})
		}),
	)
}