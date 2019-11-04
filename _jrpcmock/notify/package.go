package notify

import "github.com/daichitakahashi/jrpc"

// Package is
func Package() jrpc.Repository {
	return jrpc.Package(func(r jrpc.Repository) {
		r.Register("sum", jrpc.HandlerFunc(notifySum), SumParams([]int{}), 0)
		r.Register("hello", jrpc.HandlerFunc(notifyHello), HelloParams([]int{}), "")
	})
}
