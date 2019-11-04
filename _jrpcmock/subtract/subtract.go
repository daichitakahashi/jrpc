package subtract

import (
	"context"
	"encoding/json"

	"github.com/daichitakahashi/httpjrpc/jrpc"
)

/*
rpc call with named parameters:

--> {"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}
<-- {"jsonrpc": "2.0", "result": 19, "id": 3}

--> {"jsonrpc": "2.0", "method": "subtract", "params": {"minuend": 42, "subtrahend": 23}, "id": 4}
<-- {"jsonrpc": "2.0", "result": 19, "id": 4}
*/

// Package is
func Package() jrpc.Repository {
	return jrpc.Package(func(r jrpc.Repository) {
		r.Register("subtract", jrpc.HandlerFunc(subtract), Params{}, Result(0))
	})
}

type (
	// Params is
	Params struct {
		Subtrahend *int
		Minuend    *int
	}

	// Result is
	Result int
)

func subtract(c context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
	var p Params
	err := jrpc.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}

	if p.Subtrahend == nil || p.Minuend == nil {
		return nil, jrpc.ErrInvalidParams()
	}

	return *(p.Minuend) - *(p.Subtrahend), nil
}
