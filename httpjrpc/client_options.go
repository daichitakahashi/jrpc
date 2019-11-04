package httpjrpc

import (
	"net/http"

	"github.com/daichitakahashi/jrpc"
)

/*
TODO:

*/

type (
	httpClientOptions struct {
		requestInterceptors []func(*http.Request)
		client              **http.Client
	}

	// HTTPClientOption is implemented jtpc.ClientOption but do nothing
	HTTPClientOption struct {
		jrpc.EmptyClientOption
		applyHTTP func(opts *httpClientOptions)
	}
)

func defaultHTTPClientOptions(clientPt **http.Client) *httpClientOptions {
	return &httpClientOptions{
		requestInterceptors: []func(*http.Request){},
		client:              clientPt,
	}
}

// WithBasicAuth is
func WithBasicAuth(user, password string) *HTTPClientOption {
	return &HTTPClientOption{
		applyHTTP: func(opts *httpClientOptions) {
			opts.requestInterceptors = append(opts.requestInterceptors, func(r *http.Request) {
				r.SetBasicAuth(user, password)
			})
		},
	}
}

// WithUserAgent is
func WithUserAgent(ua string) *HTTPClientOption {
	return &HTTPClientOption{
		applyHTTP: func(opts *httpClientOptions) {
			opts.requestInterceptors = append(opts.requestInterceptors, func(r *http.Request) {
				r.Header.Set("User-Agent", ua)
			})
		},
	}
}

// WithHTTPClient is
// proof: https://play.golang.org/p/RCLHTp-qY80
func WithHTTPClient(hc *http.Client) *HTTPClientOption {
	return &HTTPClientOption{
		applyHTTP: func(opts *httpClientOptions) {
			*opts.client = hc
		},
	}
}

// WithRoundTripper option
func WithRoundTripper(rt http.RoundTripper) *HTTPClientOption {
	return &HTTPClientOption{
		applyHTTP: func(opts *httpClientOptions) {
			(*(opts.client)).Transport = rt
		},
	}
}
