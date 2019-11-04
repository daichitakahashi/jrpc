package sum

import (
	"context"
	"encoding/json"

	"github.com/daichitakahashi/httpjrpc/jrpc"
)

// Package is
func Package() jrpc.Repository {
	return jrpc.Package(func(r jrpc.Repository) {
		r.Register("sum", jrpc.HandlerFunc(sum), Params([]int{}), Result(0))
	})
}

type (
	// Params is
	Params []int

	// Result is
	Result int
)

func sum(c context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
	var p Params
	err := jrpc.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}

	var s int

	for _, n := range p {
		s += n
	}

	return s, nil
}
