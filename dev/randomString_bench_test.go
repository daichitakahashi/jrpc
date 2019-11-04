package dev

import (
	"encoding/base64"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkCreateRandomString(b *testing.B) {
	length := 20
	var v string

	b.Run("by char array", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = string(randASCIIBytes(length))
		}
	})

	b.Run("by base64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = randStr(length)
		}
	})

	b.Run("fixed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = random(length)
		}
	})

	b.Run("aaa", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v = randStringBytesMaskImprSrc(length)
		}
	})

	v = ""
	if v != "" {
		panic("dummy")
	}
}

func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:len]
}

// len(encodeURL) == 64. This allows (x <= 265) x % 64 to have an even
// distribution.
const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// A helper function create and fill a slice of length n with characters from
// a-zA-Z0-9_-. It panics if there are any problems getting random bytes.
func randASCIIBytes(n int) []byte {
	output := make([]byte, n)

	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)

	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])

		// random % 64
		randomPos := random % uint8(len(encodeURL))

		// put into output
		output[pos] = encodeURL[randomPos]
	}

	return output
}

func random(n int) string {
	output := make([]byte, n)

	// read all random
	_, err := rand.Read(output)

	if err != nil {
		panic(err)
	}

	// fill output
	for pos := range output {
		// random value % 64
		randomPos := uint8(output[pos]) % uint8(len(encodeURL))

		// put into output
		output[pos] = encodeURL[randomPos]
	}

	// return string(output)
	return *(*string)(unsafe.Pointer(&output))
}

// https://qiita.com/srtkkou/items/ccbddc881d6f3549baf1

const (
	letterBytes     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-" // len(letterBytes) == 63
	letterIndexBits = 6                                                                 // 6 bits to represent a letter index (2^6 = 64)
	letterIndexMask = 1<<letterIndexBits - 1                                            // All 1-bits, as many as letterIndexBits (00000000 -> 01000000 -> 00111111)
	letterIndexMax  = 63 / letterIndexBits                                              // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano()) // not concyrrency safe

func randStringBytesMaskImprSrc(n int) string {
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
