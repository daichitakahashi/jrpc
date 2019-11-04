package jrpc

import (
	"errors"
	"sync"
)

/*
TODO:

*/

type (
	// Repository is
	Repository interface {
		Register(method string, handler Handler, params, result interface{}) error
		With(interceptors ...Interceptor)
		Interceptors() Interceptors
		Namespace(namespace string, fn func(Repository))
		registerMethod(method string, handler Handler, interceptorChain Interceptor, params, result interface{})
	}

	// Core is
	Core struct {
		*MethodRepository
		methods sync.Map
		options options
	}

	// Metadata is
	Metadata struct {
		Handler          Handler     // raw handler
		InterceptorChain Interceptor // interceptor stack
		Params           interface{}
		Result           interface{}
	}
)

// NewRepository is
func NewRepository(opts ...Option) *Core {
	repository := &Core{
		MethodRepository: &MethodRepository{},
		options:          defaultOptions,
	}
	for _, opt := range opts {
		opt.apply(&repository.options)
	}
	repository.MethodRepository.parent = repository
	repository.MethodRepository.sep = repository.options.namespaceSeparator
	return repository
}

func (c *Core) registerMethod(method string, handler Handler, interceptorChain Interceptor, params, result interface{}) {
	/*
		if interceptorChain == nil {
			panic("jrpc: internal: interceptorChain is empty") // should not be called
		}
	*/
	c.methods.Store(method, &Metadata{
		Handler:          handler,
		InterceptorChain: interceptorChain,
		Params:           params,
		Result:           result,
	})
}

// Methods returns all registered method info
func (c *Core) Methods() map[string]Metadata {
	_copy := make(map[string]Metadata)
	c.methods.Range(func(key, value interface{}) bool {
		_copy[key.(string)] = *(value.(*Metadata))
		return true
	})
	return _copy
}

var _ Repository = (*Core)(nil)

// MethodRepository is
type MethodRepository struct {
	namespace    string
	sep          string
	interceptors Interceptors
	parent       Repository
}

// Register is
func (mr *MethodRepository) Register(method string, handler Handler, params, result interface{}) error {
	method = mr.trimSeparator(method)
	if method == "" || handler == nil {
		return errors.New("jrpc: method name and function should not be empty")
	}
	methodFullName := mr.appendNamespace(mr.namespace, method)
	mr.registerMethod(methodFullName, handler, nil, params, result)
	return nil
}

func (mr *MethodRepository) registerMethod(methodFullName string, handler Handler, interceptorChain Interceptor, params, result interface{}) {
	if interceptorChain == nil {
		interceptorChain = mr.interceptors.chained
	} else {
		interceptorChain = mr.interceptors.wrapChain(interceptorChain)
	}
	mr.parent.registerMethod(methodFullName, handler, interceptorChain, params, result)
}

// With is
func (mr *MethodRepository) With(interceptors ...Interceptor) {
	if mr.interceptors == nil {
		mr.interceptors = make(Interceptors, len(interceptors))
		copy(mr.interceptors, interceptors)
	} else {
		mr.interceptors = append(mr.interceptors, interceptors...)
	}
}

// Interceptors is
func (mr *MethodRepository) Interceptors() Interceptors {
	if mr.interceptors == nil {
		return Interceptors{}
	}
	interceptors := make(Interceptors, len(mr.interceptors))
	copy(interceptors, mr.interceptors)
	return interceptors
}

// Namespace is
func (mr *MethodRepository) Namespace(namespace string, assignFunc func(Repository)) {
	namespace = mr.trimSeparator(namespace)
	assignFunc(&MethodRepository{
		namespace: mr.appendNamespace(mr.namespace, namespace),
		sep:       mr.sep,
		parent:    mr,
	})
}

var _ Repository = (*MethodRepository)(nil)

func (mr *MethodRepository) trimSeparator(namespace string) string {
	for len(namespace) > 0 && namespace[0] == mr.sep[0] {
		namespace = namespace[1:]
	}
	for {
		if l := len(namespace); l > 0 && namespace[l-1] == mr.sep[0] {
			namespace = namespace[:l-1]
		} else {
			break
		}
	}
	return namespace
}

func (mr *MethodRepository) appendNamespace(base, namespace string) string {
	if base == "" {
		return namespace
	} else if namespace == "" {
		return base
	}
	return base + mr.sep + namespace
}
