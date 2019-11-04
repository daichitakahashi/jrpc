package jrpc

import (
	"context"
	"encoding/json"
)

type (
	// Handler is
	Handler interface {
		ServeJSONRPC(ctx context.Context, params *json.RawMessage) (interface{}, *Error)
	}

	// HandlerFunc is
	HandlerFunc func(context.Context, *json.RawMessage) (interface{}, *Error)
)

// ServeJSONRPC is
func (f HandlerFunc) ServeJSONRPC(ctx context.Context, params *json.RawMessage) (interface{}, *Error) {
	return f(ctx, params)
}

type (
	// Interceptor is
	Interceptor func(context.Context, *json.RawMessage, *RequestInfo, Handler) (interface{}, *Error)

	// Interceptors is
	Interceptors []Interceptor

	// RequestInfo is
	RequestInfo struct {
		MethodFullName string // 機能を定義する側は、自身がどうマウントされたのかは知り得ないため、有用
		IsBatch        bool   // バッチリクエストかどうかの情報が、何に使えるのか。ロギング？
		ID             ID
	} // params は、json.Compactを使おう
)

func (ints Interceptors) chained(ctx context.Context, params *json.RawMessage, info *RequestInfo, handler Handler) (interface{}, *Error) {
	return ints.chain(ctx, params, info, handler, 0)
}

func (ints Interceptors) chain(ctx context.Context, params *json.RawMessage, info *RequestInfo, handler Handler, i int) (interface{}, *Error) {
	if i == len(ints) {
		return handler.ServeJSONRPC(ctx, params)
	}
	return ints[i](ctx, params, info, HandlerFunc(func(ctx2 context.Context, params2 *json.RawMessage) (interface{}, *Error) {
		return ints.chain(ctx2, params2, info, handler, i+1)
	}))
}

func (ints Interceptors) wrapChain(sub Interceptor) Interceptor {
	return func(ctx context.Context, params *json.RawMessage, info *RequestInfo, handler Handler) (interface{}, *Error) {
		return ints.chain(ctx, params, info, HandlerFunc(func(ctx2 context.Context, params2 *json.RawMessage) (interface{}, *Error) {
			return sub(ctx2, params2, info, handler)
		}), 0)
	}
}
