package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

func newMock() *Core {
	r := NewRepository()

	r.Register("sum", HandlerFunc(sum), []int{}, 0)
	r.Register("subtract", HandlerFunc(subtract), []int{}, 0)

	r.Namespace("wait", func(r Repository) {
		r.With(idSleepInterceptor)
		r.Register("sayHello", HandlerFunc(sayHello), nil, "")
	})

	r.Namespace("err", func(r Repository) {
		r.Register("maker", HandlerFunc(errMaker), nil, "")
		r.Register("panicMaker", HandlerFunc(panicMaker), "", nil)
		r.Register("encodeError", HandlerFunc(encodeError), nil, nil)
	})

	return r
}

func sum(_ context.Context, params *json.RawMessage) (interface{}, *Error) {
	var v []int
	err := UnmarshalParams(params, &v)
	if err != nil {
		return nil, err
	}
	var result int
	for _, n := range v {
		result += n
	}
	return result, nil
}

func subtract(_ context.Context, params *json.RawMessage) (interface{}, *Error) {
	var v []int
	err := UnmarshalParams(params, &v)
	if err != nil {
		return nil, err
	} else if len(v) != 2 {
		return nil, &Error{
			Message: "only two arguments",
		}
	}
	return v[0] - v[1], nil
}

func sayHello(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
	return "hello", nil
}

func errMaker(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
	return "dummy", &Error{
		Message: "error occurred",
	}
}

func panicMaker(_ context.Context, params *json.RawMessage) (interface{}, *Error) {
	var msg string
	UnmarshalParams(params, &msg)
	panic(msg)
}

func encodeError(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
	return &encErr{}, nil
}

type encErr struct{}

func (e *encErr) MarshalJSON() ([]byte, error) {
	return nil, errors.New("encode error")
}

func idSleepInterceptor(ctx context.Context, params *json.RawMessage, info *RequestInfo, handler Handler) (interface{}, *Error) {
	d, err := info.ID.Int64()
	if err == nil {
		val := time.Duration(d)
		time.Sleep(time.Millisecond * val)
	}
	return handler.ServeJSONRPC(ctx, params)
}
