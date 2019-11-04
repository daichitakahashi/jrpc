package notify

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/daichitakahashi/jrpc"
)

// HelloParams is
type HelloParams []int

func notifyHello(c context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
	var p HelloParams
	err := jrpc.Unmarshal(params, &p)
	if err != nil {
		return nil, err
	}

	var s string

	for _, n := range p {
		s += strconv.Itoa(n)
	}

	return s, nil
}
