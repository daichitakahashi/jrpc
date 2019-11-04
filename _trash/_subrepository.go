package jrpc

/*
import (
	"context"
	"encoding/json"
	"errors"
)

// MethodRepository is
type MethodRepository struct {
	parents      []Repository
	methods      map[string]Metadata
	interceptors Interceptors
	core         *Core
}

// Package makes
func Package(fn func(repository Repository)) Repository {
	r := newSubRepository()
	fn(r)
	return r
}

func newSubRepository() *MethodRepository {
	return &MethodRepository{
		parents:      make([]Repository, 0),
		interceptors: make(Interceptors, 0), // キャパシティ要検討
	}
}

// Register is
func (mr *MethodRepository) Register(method string, handler Handler, params, result interface{}) error {
	if method == "" || handler == nil {
		return errors.New("jrpc: method name and function should not be empty")
	}
	methodData := Metadata{
		Handler: handler,
		Params:  params,
		Result:  result,
	}
	mr.methods[method] = methodData // sync でなくていい？

	if len(mr.parents) != 0 {
		for _, parent := range mr.parents {
			parent.Register(method, handler, params, result)
		}
	}
	return nil
}

// With is
func (mr *MethodRepository) With(interceptors ...Interceptor) Repository {
	// ここでlazyにアロケーションという手もある
	// len(nil)が0値を返すことを利用する
	mr.interceptors = append(mr.interceptors, interceptors...)
	return mr
}

// Interceptors is
func (mr *MethodRepository) Interceptors() Interceptors {
	return mr.interceptors
}

// Namespace is
func (mr *MethodRepository) Namespace(namespace string, fn func(Repository)) Repository {
	child := newSubRepository()
	fn(child)
	child.assign(newNamespace(mr, namespace))
	return mr
}

// Assign is
func (mr *MethodRepository) Assign(namespace string, repository Repository) Repository {
	repository.assign(newNamespace(mr, namespace))
	return nil
}

func (mr *MethodRepository) assign(parent Repository) {
	for name, methodData := range mr.methods {
		parent.Register(name, HandlerFunc(func(c context.Context, params *json.RawMessage) (interface{}, *Error) {
			return mr.interceptors.apply(methodData.Handler).ServeJSONRPC(c, params)
		}), methodData.Params, methodData.Result)
	}
	mr.parents = append(mr.parents, parent)
}
*/
