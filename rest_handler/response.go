package rest_handler

import (
	"errors"
	"net/http"

	"github.com/Nauman-S/Internal-Transfers-System/codes"
	"github.com/gin-gonic/gin"
)

const (
	RespCtxCodeLabel = "app-ctx-resp-code"
	RespCtxMsgLabel  = "app-ctx-resp-msg"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type RespAdapter interface {
	PackData(c *gin.Context, data any)
	PackError(c *gin.Context, err error)
}

type StdRespAdapter struct{}

func (h *StdRespAdapter) PackData(c *gin.Context, data any) {
	c.Set(RespCtxCodeLabel, codes.Success.Code)
	c.Set(RespCtxMsgLabel, codes.Success.Msg)
	c.AbortWithStatusJSON(http.StatusOK, data)
}

func (h *StdRespAdapter) PackError(c *gin.Context, err error) {
	code := codes.GetCode(err)
	msg := codes.GetMsg(err)
	c.Set(RespCtxCodeLabel, code)
	c.Set(RespCtxMsgLabel, msg)
	
	statusCode := mapErrorToStatusCode(err)
	
	c.AbortWithStatusJSON(statusCode, Response{
		Code:    code,
		Message: msg,
	})
}

func mapErrorToStatusCode(err error) int {
	var codeErr codes.CodeError
	if !errors.As(err, &codeErr) {
		return http.StatusInternalServerError // Default for non-CodeError
	}
	
	switch codeErr.Code {
	case codes.ErrSystem.Code:
		return http.StatusInternalServerError
	case codes.ErrInvalidParams.Code:
		return http.StatusBadRequest
	case codes.ErrTimeout.Code:
		return http.StatusRequestTimeout
		
	// Account Codes
	case codes.ErrNegativeBalance.Code:
		return http.StatusBadRequest
	case codes.ErrAccountExists.Code:
		return http.StatusConflict
	case codes.ErrInvalidAccountID.Code:
		return http.StatusBadRequest
	case codes.ErrAccountNotFound.Code:
		return http.StatusNotFound
		
	// Transaction Codes
	case codes.ErrSameAccountTransfer.Code:
		return http.StatusBadRequest
	case codes.ErrInsufficientFunds.Code:
		return http.StatusBadRequest
	case codes.ErrSourceAccountNotFound.Code:
		return http.StatusNotFound
	case codes.ErrDestinationAccountNotFound.Code:
		return http.StatusNotFound
		
	default:
		return http.StatusInternalServerError
	}
}

func (r *Response) Error() string {
	return r.Message
}