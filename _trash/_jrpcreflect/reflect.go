package jrpcreflect

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/daichitakahashi/httpjrpc/jrpc"
)

// ReflectHandler is
type ReflectHandler interface {
	ServeJRPC(c context.Context, params interface{}) (interface{}, *jrpc.Error)
}

// ReflectHandlerFunc is
type ReflectHandlerFunc func(c context.Context, params interface{}) (interface{}, *jrpc.Error)

// ServeJRPC is
func (rhf ReflectHandlerFunc) ServeJRPC(c context.Context, params interface{}) (interface{}, *jrpc.Error) {
	return rhf(c, params)
}

// Validator interface
type Validator interface {
	Validate() error
}

// Reflect is
func Reflect(method string, refHandler ReflectHandler, params, result interface{}) (m string, h jrpc.Handler, p, r interface{}) {
	paramType := reflect.TypeOf(params)
	resultType := reflect.TypeOf(result)

	return method, jrpc.HandlerFunc(func(ctx context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
		prm := reflect.New(paramType).Interface() // ptr
		err := json.Unmarshal(*params, prm)
		if err != nil {
			// パースエラーはInvalidParam
			return nil, &jrpc.Error{
				Message: "unmarshal error",
			}
		}

		if validator, ok := prm.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return nil, &jrpc.Error{
					Message: "validation error: " + err.Error(),
				}
			}
		}

		res, Err := refHandler.ServeJRPC(ctx, prm)
		if Err != nil {
			return nil, Err
		}

		//actualResultType := reflect.TypeOf(res)
		actualResultType := reflect.ValueOf(res).Type()
		if resultType != actualResultType {
			panic(reflectError(ctx.(jrpc.Context).Method(), resultType, actualResultType))
		}

		return res, Err
	}), params, result
}
