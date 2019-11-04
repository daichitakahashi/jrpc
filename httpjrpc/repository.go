package httpjrpc

import (
	"net/http"

	"github.com/daichitakahashi/jrpc"
)

/*
TODO:
- Repositoryの実装

*/

// Repository is
type Repository struct {
	*jrpc.Core
}

func (r *Repository) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// concurrency safeのためには、毎回作り直すべき
	dec := jrpc.NewDecoder(req.Body)
	// poolを使うのがいい
	requests := make([]*jrpc.Request, 0, 10)     // であれば、レスポンスを返す場合には、エラーは返さず、内部に保存しておくこととする？
	requests, batch, err := dec.Decode(requests) // if dec.Err() != nil { dec.Reset(conn) }
	if len(requests) == 0 {
		if err != nil {
			// connection error
		} else {
			// invalid request
		}
	}

	resps, err := r.Execute(req.Context(), requests, batch)
	if err != nil {
		//
	}

	enc := jrpc.NewEncoder(w)
	err = enc.Encode(resps, batch)
	if err != nil {
		//
	}
}
