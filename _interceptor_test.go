package jrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var buf bytes.Buffer

func first(next InterceptorHandler) InterceptorHandler {
	return InterceptorFunc(func(ctx Context, params *json.RawMessage) (interface{}, *Error) {
		buf.WriteString("first,")
		return next.Intercept(ctx, params)
	})
}

func second(next InterceptorHandler) InterceptorHandler {
	return InterceptorFunc(func(ctx Context, params *json.RawMessage) (interface{}, *Error) {
		buf.WriteString("second,")
		return next.Intercept(ctx, params)
	})
}

func canceler(next InterceptorHandler) InterceptorHandler {
	return InterceptorFunc(func(ctx Context, params *json.RawMessage) (interface{}, *Error) {
		return nil, &Error{
			Message: "cancellation",
		}
	})
}

func TestInterceptor(t *testing.T) {

	r := newRootRepository()
	r.With(first, second)

	r.Register("interceptor.sample", HandlerFunc(func(c context.Context, params *json.RawMessage) (interface{}, *Error) {
		buf.WriteString("handler")
		return nil, nil
	}), nil, nil)

	methods := r.Methods()
	methods["interceptor.sample"].Handler.ServeJSONRPC(context.Background(), nil)

	assert.Equal(t, "first,second,handler", buf.String())

	buf = bytes.Buffer{}
	r.With(canceler)
	_, err := methods["interceptor.sample"].Handler.ServeJSONRPC(context.Background(), nil)
	assert.Error(t, err)
	assert.Equal(t, "first,second,", buf.String())
}
