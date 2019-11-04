package notify

import (
	"context"
	"encoding/json"

	"github.com/daichitakahashi/jrpc"
)

// SumParams is
type SumParams []int

func notifySum(c context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
	var p SumParams
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
