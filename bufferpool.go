// Copyright (c) 2019 daichitakahashi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package jrpc

import (
	"io"
	"strconv"
	"sync"
)

/*
TODO:

*/

var bufferpool = NewPool()

// ===== begin: from https://github.com/uber-go/zap/blob/master/buffer/pool.go =====

// A Pool is a type-safe wrapper around a sync.Pool.
type Pool struct {
	p *sync.Pool
}

// NewPool constructs a new Pool.
func NewPool() Pool {
	return Pool{p: &sync.Pool{
		New: func() interface{} {
			return &Buffer{bs: make([]byte, 0, _size)}
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p Pool) Get() *Buffer {
	buf := p.p.Get().(*Buffer)
	buf.Reset()
	buf.pool = p
	return buf
}

func (p Pool) put(buf *Buffer) {
	p.p.Put(buf)
}

// ===== end: from https://github.com/uber-go/zap/blob/master/buffer/pool.go =====

// GetForBufferSlice retrieves a Buffer from the pool and grow length up to its capacity.
func (p Pool) GetForBufferSlice() *Buffer {
	buf := p.Get()
	buf.fill()
	return buf
}

// ===== begin: from https://github.com/uber-go/zap/blob/master/buffer/pool.go =====

const _size = 1024 // by default, create 1 KiB buffers

// Buffer is a thin wrapper around a byte slice. It's intended to be pooled, so
// the only way to construct one is via a Pool.
type Buffer struct {
	bs   []byte
	pool Pool

	c int // added for Read
}

// AppendByte writes a single byte to the Buffer.
func (b *Buffer) AppendByte(v byte) {
	b.bs = append(b.bs, v)
}

// AppendString writes a string to the Buffer.
func (b *Buffer) AppendString(s string) {
	b.bs = append(b.bs, s...)
}

// AppendInt appends an integer to the underlying buffer (assuming base 10).
func (b *Buffer) AppendInt(i int64) {
	b.bs = strconv.AppendInt(b.bs, i, 10)
}

// AppendUint appends an unsigned integer to the underlying buffer (assuming
// base 10).
func (b *Buffer) AppendUint(i uint64) {
	b.bs = strconv.AppendUint(b.bs, i, 10)
}

// AppendBool appends a bool to the underlying buffer.
func (b *Buffer) AppendBool(v bool) {
	b.bs = strconv.AppendBool(b.bs, v)
}

// AppendFloat appends a float to the underlying buffer. It doesn't quote NaN
// or +/- Inf.
func (b *Buffer) AppendFloat(f float64, bitSize int) {
	b.bs = strconv.AppendFloat(b.bs, f, 'f', -1, bitSize)
}

// Len returns the length of the underlying byte slice.
func (b *Buffer) Len() int {
	return len(b.bs) - b.c
}

/*
	// Cap returns the capacity of the underlying byte slice.
	func (b *Buffer) Cap() int {
		return cap(b.bs)
	}
*/

// Bytes returns a mutable reference to the underlying byte slice.
func (b *Buffer) Bytes() []byte {
	return b.bs[b.c:]
}

// String returns a string copy of the underlying byte slice.
func (b *Buffer) String() string {
	return string(b.bs[b.c:])
}

// Reset resets the underlying byte slice. Subsequent writes re-use the slice's
// backing array.
func (b *Buffer) Reset() {
	b.bs = b.bs[:0]

	b.c = 0 // added for Read
}

// Write implements io.Writer.
func (b *Buffer) Write(bs []byte) (int, error) {
	b.bs = append(b.bs, bs...)
	return len(bs), nil
}

/*
	// TrimNewline trims any final "\n" byte from the end of the buffer.
	func (b *Buffer) TrimNewline() {
		if i := len(b.bs) - 1; i >= 0 {
			if b.bs[i] == '\n' {
				b.bs = b.bs[:i]
			}
		}
	}
*/

// Free returns the Buffer to its Pool.
//
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Free() {
	b.pool.put(b)
}

// ===== end: from https://github.com/uber-go/zap/blob/master/buffer/pool.go =====

/*
// AppendPrefix writes a single byte to the head of Buffer
func (b *Buffer) AppendPrefix(p byte) {
	if len(b.bs) == 0 {
		b.bs = append(b.bs, p)
	} else {
		//b.bs, b.bs[0] = append(b.bs[0:1], b.bs...), p
		b.bs, b.bs[b.c] = append(b.bs[b.c:b.c+1], b.bs...), p
	}
}
*/

/*
DEPRECATED: TOO SLOW
// AppendQuote appends a double-quoted Go string literal representing s to the underlying buffer.
func (b *Buffer) AppendQuote(s string) {
	b.bs = strconv.AppendQuote(b.bs, s)
}
*/

/*
// WriteTo implements io.WriterTo.
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.bs[b.c:])
	//if n != b.Len()-b.c {
	if n != b.Len() {
		return int64(n), io.ErrShortWrite
	}
	return int64(n), err
}
*/

// Read implements io.Reader.
func (b *Buffer) Read(p []byte) (int, error) {
	if b.c >= len(b.bs) {
		b.Free()
		return 0, io.EOF
	}
	n := copy(p, b.bs[b.c:])
	b.c += n
	return n, nil
}

// fill grows length of underlying byte slice up to its capacity.
func (b *Buffer) fill() {
	b.bs = b.bs[:cap(b.bs)]
}

type encoder interface {
	encode() (*Buffer, error)
	encodeTo(*Buffer) error
}

type encoderLine interface {
	encoder
	encodeLine() (*Buffer, error)
}
