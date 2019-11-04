package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

func main() {
	factory := &RandomIDFactory{
		Prefix:            "aggregate-",
		Digits:            10,
		BatchPrefixDigits: 5,
	}
	fmt.Println(factory.CreateID())
	fmt.Println(factory.CreateID())
	fmt.Println(factory.CreateID())
	fmt.Println(factory.CreateID())

	batchFactory := factory.BatchIDFactory()
	fmt.Println(batchFactory.CreateID().String())
	fmt.Println(batchFactory.CreateID().String())
	fmt.Println(batchFactory.CreateID().String())
	fmt.Println(batchFactory.CreateID().String())
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

const (
	letterBytes     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+" // len(letterBytes) == 63
	letterIndexBits = 6                                                                 // 6 bits to represent a letter index (2^6 = 64)
	letterIndexMask = 1<<letterIndexBits - 1                                            // All 1-bits, as many as letterIndexBits (00000000 -> 01000000 -> 00111111)
	letterIndexMax  = 63 / letterIndexBits                                              // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano()) // not concyrrency safe

func randomString(n int) string {
	// b is filled from tail to head
	b := make([]byte, n)

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

	return *(*string)(unsafe.Pointer(&b))
}

// ID is
type ID struct {
	val interface{}
}

// NewID is
func NewID(v interface{}) ID {
	id := ID{}
	if v == nil {
		id.val = nil
		return id
	}
	switch val := v.(type) {
	case int:
		id.val = int64(val)
		return id
	case int8:
		id.val = int64(val)
		return id
	case int16:
		id.val = int64(val)
		return id
	case int32:
		id.val = int64(val)
		return id
	case int64:
		id.val = val
		return id
	case uint:
		id.val = int64(val)
		return id
	case uint8:
		id.val = int64(val)
		return id
	case uint16:
		id.val = int64(val)
		return id
	case uint32:
		id.val = int64(val)
		return id
	case uint64:
		id.val = int64(val)
		return id
	case float32:
		id.val = float64(val)
		return id
	case float64:
		id.val = val
		return id
	case string:
		id.val = val
		return id
	}
	panic(errors.New("jrpc: jrpc.ID: invalid type of \"id\" member"))
}

func (id ID) String() string {
	return id.val.(string)
}
