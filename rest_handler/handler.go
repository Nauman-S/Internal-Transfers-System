package rest_handler

import (
	"bytes"
	"errors"
	"io"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/Nauman-S/Internal-Transfers-System/codes"
)

type FrontFilter func(ctx *gin.Context) error

type RequestFilter func(ctx *gin.Context, req any) error

type Handler struct {
	frontFilters   []FrontFilter
	requestFilters []RequestFilter
	respAdaptor    RespAdapter
}

type handlerFun any

func NewHandler(frontFilters []FrontFilter, requestFilters []RequestFilter) *Handler {
	return &Handler{
		frontFilters:   frontFilters,
		requestFilters: requestFilters,
		respAdaptor:    &StdRespAdapter{},
	}
}

func (h *Handler) SetResponseAdaptor(respAdaptor RespAdapter) *Handler {
	h.respAdaptor = respAdaptor
	return h
}

func callHandleFunc(fun handlerFun, args ...any) []any {
	fv := reflect.ValueOf(fun)

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}

	rs := fv.Call(params)
	result := make([]any, len(rs))
	for i, r := range rs {
		result[i] = r.Interface()
	}
	return result
}

func createHandleReqArg(fun handlerFun, ctx *gin.Context) (any, error) {
	ft := reflect.TypeOf(fun)
	if ft.NumIn() == 1 {
		return nil, nil
	}

	argType := ft.In(1).Elem()
	reqArg := reflect.New(argType).Interface()

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, codes.NewWithMsg(codes.ErrSystem, "read request body err: %v", err)
	}
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err = jsoniter.Unmarshal(bodyBytes, reqArg); err != nil {
		return nil, codes.NewWithMsg(codes.ErrInvalidParams, "unmarshal request body err: %v", err)
	}

	validate := validator.New()
	if err = validate.Struct(reqArg); err != nil {
		return nil, codes.NewWithMsg(codes.ErrInvalidParams, "validator param err: %v", err)
	}
	return reqArg, nil
}

func (h *Handler) HandleMiddleware(handleFunc any) func(*gin.Context) {
	if err := ValidateFuncType(handleFunc); err != nil {
		log.WithError(err).Fatal("validate function type failed")
	}

	return func(ctx *gin.Context) {
		for _, filter := range h.frontFilters {
			if err := filter(ctx); err != nil {
				h.respAdaptor.PackError(ctx, err)
				return
			}
		}
		h.handleRequest(ctx, handleFunc)
	}
}

func (h *Handler) handleRequest(ctx *gin.Context, fun handlerFun) {
	args, err := h.buildHandleFuncArgs(fun, ctx)
	if err != nil {
		h.respAdaptor.PackError(ctx, err)
		return
	}

	result := callHandleFunc(fun, args...)
	if err := result[len(result)-1]; err != nil {
		h.respAdaptor.PackError(ctx, err.(error))
		return
	}

	if len(result) == 1 {
		h.respAdaptor.PackData(ctx, struct{}{})
		return
	}
	h.respAdaptor.PackData(ctx, result[0])
}

func (h *Handler) buildHandleFuncArgs(fun handlerFun, ctx *gin.Context) ([]any, error) {
	req, err := createHandleReqArg(fun, ctx)
	if err != nil {
		return nil, err
	}

	for _, filter := range h.requestFilters {
		if err = filter(ctx, req); err != nil {
			return nil, err
		}
	}

	args := []any{ctx}
	if req != nil {
		args = append(args, req)
	}
	return args, nil
}

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*gin.Context)(nil))
)

func ValidateFuncType(fun handlerFun) error {
	ft := reflect.TypeOf(fun)
	if ft.Kind() != reflect.Func || ft.IsVariadic() {
		return errors.New("need non-variadic func in " + ft.String())
	}

	if ft.NumIn() < 1 || ft.NumIn() > 3 {
		return errors.New("need one or two or three parameters in " + ft.String())
	}

	if ft.In(0) != contextType {
		return errors.New("the first parameter must point of context in " + ft.String())
	}

	if ft.NumIn() == 2 && ft.In(1).Kind() != reflect.Ptr {
		return errors.New("the second parameter must point in " + ft.String())
	}

	if ft.NumOut() < 1 || ft.NumOut() > 2 {
		return errors.New("the size of return value must one or two in " + ft.String())
	}

	if !ft.Out(ft.NumOut() - 1).Implements(errorType) {
		return errors.New("the last return value must error in " + ft.String())
	}
	return nil
}