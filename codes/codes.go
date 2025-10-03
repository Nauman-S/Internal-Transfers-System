package codes

import (
	"errors"
	"fmt"
)

var (
	//General Application Codes
	Success          = CodeError{Code: 0, Msg: "success"}
	ErrSystem        = CodeError{Code: 1, Msg: "system error"}
	ErrInvalidParams = CodeError{Code: 2, Msg: "invalid params"}
	ErrTimeout       = CodeError{Code: 3, Msg: "request timeout"}

	//Account Codes
	ErrNegativeBalance = CodeError{
		Code:    4,
		Msg: "initial balance cannot be negative",
	}

	ErrAccountExists = CodeError{
		Code:    5,
		Msg: "account with this ID already exists",
	}
	
	ErrInvalidAccountID = CodeError{
		Code:    6,
		Msg: "account ID must be a positive integer",
	}

	ErrAccountNotFound = CodeError{
		Code:    7,
		Msg: "account not found",
	}

	//Transaction Codes
	ErrSameAccountTransfer = CodeError{
		Code: 8,
		Msg:  "cannot transfer to the same account",
	}
	ErrInsufficientFunds = CodeError{
		Code: 9,
		Msg:  "insufficient funds for transfer",
	}
	ErrSourceAccountNotFound = CodeError{
		Code: 10,
		Msg:  "source account not found",
	}
	ErrDestinationAccountNotFound = CodeError{
		Code: 11,
		Msg:  "destination account not found",
	}
)

type CodeError struct {
	Code int
	Msg  string
}

func (e CodeError) Error() string {
	return e.Msg
}

func New(code int, msg string) CodeError {
	return CodeError{Code: code, Msg: msg}
}

func NewWithMsg(err CodeError, format string, args ...any) CodeError {
	return CodeError{Code: err.Code, Msg: fmt.Sprintf(format, args...)}
}

func GetCode(err error) int {
	if err == nil {
		return Success.Code
	}

	var cerr CodeError
	if errors.As(err, &cerr) {
		return cerr.Code
	}

	return ErrSystem.Code
}

func GetMsg(err error) string {
	if err == nil {
		return Success.Msg
	}

	var cerr CodeError
	if errors.As(err, &cerr) {
		return cerr.Msg
	}
	return ErrSystem.Msg
}