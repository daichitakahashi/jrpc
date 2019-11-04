package jrpc

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

/*
TODO:

*/

type (
	clientOptions struct {
		idFactory IDFactory
	}

	// ClientOption is
	ClientOption interface {
		apply(otps *clientOptions)
	}

	clientOptionFunc func(opts *clientOptions)
)

func (cof clientOptionFunc) apply(opts *clientOptions) {
	cof(opts)
}

// EmptyClientOption is xxx . It can be embedded in
// another structure to build custom options
type EmptyClientOption struct{}

func (EmptyClientOption) apply(*clientOptions) {}

var defaultClientOption = clientOptions{
	idFactory: &RandomIDFactory{
		Digits:            10,
		BatchPrefixDigits: 5,
	},
}

// WithIDFactory is
func WithIDFactory(factory IDFactory) ClientOption {
	return clientOptionFunc(func(opts *clientOptions) {
		opts.idFactory = factory
	})
}

// IDFactory is
type IDFactory interface {
	CreateID() ID
	BatchIDFactory() IDFactory
}

// RandomIDFactory is
type RandomIDFactory struct {
	Prefix            string
	Digits            int
	BatchPrefixDigits int
}

// CreateID makes
func (rif *RandomIDFactory) CreateID() ID {
	return NewID(rif.Prefix + randomString(rif.Digits))
}

// BatchIDFactory is
func (rif *RandomIDFactory) BatchIDFactory() IDFactory {
	return &RandomIDFactory{
		Prefix: rif.Prefix + randomString(rif.BatchPrefixDigits) + "-",
		Digits: rif.Digits,
	}
}

// from https://qiita.com/srtkkou/items/ccbddc881d6f3549baf1
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

const (
	letterBytes     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+" // len(letterBytes) == 63
	letterIndexBits = 6                                                                 // 6 bits to represent a letter index (2^6 = 64)
	letterIndexMask = 1<<letterIndexBits - 1                                            // All 1-bits, as many as letterIndexBits (00000000 -> 01000000 -> 00111111)
	letterIndexMax  = 63 / letterIndexBits                                              // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano()) // not concyrrency safe
var randMutex sync.Mutex

func randomString(n int) string {
	// b is filled from tail to head
	b := make([]byte, n)

	randMutex.Lock()
	for i, cache, remain := n-1, src.Int63(), letterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIndexMax
		}
		if index := int(cache & letterIndexMask); index < len(letterBytes) {
			b[i] = letterBytes[index]
			i--
		}
		cache >>= letterIndexBits // consume used 6 bits
		remain--
	}

	randMutex.Unlock()
	return *(*string)(unsafe.Pointer(&b))
}

var _ IDFactory = (*RandomIDFactory)(nil)
