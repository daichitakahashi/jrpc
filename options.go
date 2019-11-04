package jrpc

type (
	options struct {
		namespaceSeparator    string
		disableConcurrentCall bool
		panicHandler          func(req *Request, recovered interface{})
	}

	// Option is
	Option interface {
		apply(opts *options)
	}

	optionFunc func(opts *options)
)

func (of optionFunc) apply(opts *options) {
	of(opts)
}

var defaultOptions = options{
	namespaceSeparator: ".",
}

// WithNamespaceSeparator is
// namespaceSeparator used for join namespace. Default is comma.
func WithNamespaceSeparator(sep byte) Option {
	return optionFunc(func(opts *options) {
		opts.namespaceSeparator = string(sep)
	})
}

// WithDisableConcurrentCall is
func WithDisableConcurrentCall() Option {
	return optionFunc(func(opts *options) {
		opts.disableConcurrentCall = true
	})
}

// WithPanicHandler register panic handler function.
// Core always return Response object with Internal Error when panic occurred during call of JSON-RPC method.
// You can get detailed information of panic in your panicHandler.
// If you allow concurrent call of JSON-RPC method, panicHandler must be concurrent-safe.
func WithPanicHandler(panicHandler func(req *Request, recovered interface{})) Option {
	return optionFunc(func(opts *options) {
		opts.panicHandler = panicHandler
	})
}
