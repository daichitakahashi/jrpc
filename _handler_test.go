package jrpc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoMethod(t *testing.T) {
	m := &mock{}
	executor := newExecutor(m, strings.NewReader(""))

	ctx := context.Background()
	param := json.RawMessage(`{"string":"test string", "int":9}`)

	request1 := &Request{
		Version: "1.5",
		Method:  "echo",
		Params:  &param,
	}
	resp1 := executor.DoMethod(ctx, request1)
	assert.Equal(t, ErrInvalidRequest(), resp1.Error)

	request2 := &Request{
		Version: "2.0",
		Method:  "",
		Params:  &param,
	}
	resp2 := executor.DoMethod(ctx, request2)
	assert.Equal(t, ErrInvalidRequest(), resp2.Error)

	request3 := &Request{
		Version: "2.0",
		Method:  "imaginary.method",
		Params:  &param,
	}
	resp3 := executor.DoMethod(ctx, request3)
	assert.Equal(t, ErrMethodNotFound(), resp3.Error)

	request4 := &Request{
		Version: "2.0",
		Method:  "echo",
		Params:  nil,
	}
	resp4 := executor.DoMethod(ctx, request4)
	assert.Equal(t, ErrInvalidParams(), resp4.Error)

	request5 := &Request{
		Version: "2.0",
		Method:  "echo",
		Params:  &param,
	}

	m.test = func(c context.Context, _ *json.RawMessage) {
		val := c.Value(RequestKey{})
		assert.NotNil(t, val)
		req := val.(*Request)
		assert.Equal(t, request5, req)
		return
	}

	resp5 := executor.DoMethod(ctx, request5)
	assert.Equal(t, "method: echo, msg: test string, 9", resp5.Result.(Result).Result)
}
