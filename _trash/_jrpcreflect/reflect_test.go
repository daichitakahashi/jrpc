package jrpcreflect

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/daichitakahashi/httpjrpc/jrpc"
	"github.com/stretchr/testify/assert"
)

func TestReflectHandlerFunc(t *testing.T) {

	var called bool
	emptyReflectHandler := ReflectHandlerFunc(func(c context.Context, params interface{}) (interface{}, *jrpc.Error) {
		called = true
		return nil, nil
	})
	emptyReflectHandler.ServeJRPC(context.Background(), nil)
	assert.True(t, called)

}

type counterHandler struct {
	count int
}

func (c *counterHandler) ServeJRPC(ctx context.Context, params interface{}) (interface{}, *jrpc.Error) {
	p := params.(*CounterParam)

	if *p.Num < 0 {
		return nil, &jrpc.Error{
			Message: "param needs to be upper",
		}
	}

	c.count += *p.Num

	return CounterResult{
		Num: c.count,
	}, nil
}

type CounterParam struct {
	Num *int `json:"num"`
}

func (c *CounterParam) Validate() error {
	if c.Num == nil {
		return errors.New("Validation error")
	}
	return nil
}

type CounterResult struct {
	Num int `json:"num"`
}

type PseudoResult struct {
	Pseudo string `json:"pseudo"`
}

func TestReflectCall(t *testing.T) {

	h := &counterHandler{}
	r := jrpc.NewRootRepository(jrpc.NewCore())

	r.Register(Reflect("success", h, CounterParam{}, CounterResult{}))
	r.Register(Reflect("failure", h, CounterParam{}, PseudoResult{}))
	methods := r.Methods()

	trueParams := json.RawMessage(`{"num":25}`)
	falseParams := json.RawMessage(`{"msg":"false param"}`)

	assert.NotPanics(t, func() {
		r, err := methods["success"].Handler.ServeJSONRPC(context.Background(), &trueParams)
		assert.Nil(t, err)
		assert.Equal(t, 25, h.count)
		assert.Equal(t, r.(CounterResult).Num, h.count)

		r, err = methods["success"].Handler.ServeJSONRPC(context.Background(), &falseParams)
		assert.Nil(t, r)
		assert.NotEmpty(t, err)
	})

	assert.Panics(t, func() {
		methods["failure"].Handler.ServeJSONRPC(context.Background(), &trueParams)
	})

	irregularParams := json.RawMessage(`{"num":-1}`)
	res, err := methods["success"].Handler.ServeJSONRPC(context.Background(), &irregularParams)
	assert.Nil(t, res)
	assert.NotEmpty(t, err)

	invalidJSONParams := json.RawMessage(`{key":val"ue}`)
	res, err = methods["success"].Handler.ServeJSONRPC(context.Background(), &invalidJSONParams)
	assert.Nil(t, res)
	assert.NotEmpty(t, err)

}
