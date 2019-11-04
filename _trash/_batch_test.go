package jrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveBatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	m := &mock{}
	w, r := net.Pipe()
	e := newExecutor(m, r)

	reset := func() {
		w.Close()
		r.Close()
		w, r = net.Pipe()
		e = newExecutor(m, r)
	}

	// unary
	go func() {
		w.Write([]byte(`{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1}`))
	}()
	raws, isBatch, err := e.ResolveBatch(ctx)
	assert.Nil(t, raws)
	assert.False(t, isBatch)
	assert.Nil(t, err)

	reset()

	// usual array
	go func() {
		w.Write([]byte(`[
			{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1},
			{"jsonrpc": "2.0", "method": "second", "params": [42, 23], "id": 2},
			{"jsonrpc": "2.0", "method": "third", "params": [42, 23], "id": 3}
		]`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Equal(t, 3, len(raws))
	assert.Equal(t, "first", raws[0].Method)
	assert.Equal(t, "second", raws[1].Method)
	assert.Equal(t, "third", raws[2].Method)
	assert.True(t, isBatch)
	assert.Nil(t, err)

	reset()

	// array but single
	go func() {
		w.Write([]byte(`[{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1}]`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Equal(t, 1, len(raws))
	assert.True(t, isBatch)
	assert.Nil(t, err)

	reset()

	// empty array
	go func() {
		w.Write([]byte(`[]`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Equal(t, 0, len(raws))
	assert.True(t, isBatch)
	assert.Nil(t, err)

	reset()

	// ivalid type
	go func() {
		w.Write([]byte(`[1,2,3]`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Equal(t, 3, len(raws))
	assert.True(t, isBatch)
	assert.Nil(t, err)

	reset()

	// EOF
	go func() {
		w.Close()
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Nil(t, raws)
	assert.False(t, isBatch)
	assert.Equal(t, io.EOF, err)

	reset()

	// broken array
	go func() {
		w.Write([]byte(`[1,2,`))
		w.Close()
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Nil(t, raws)
	assert.False(t, isBatch)
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	reset()

	// broken json
	go func() {
		w.Write([]byte(`[1,2,{"json`))
		w.Close()
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Nil(t, raws)
	assert.False(t, isBatch)
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	reset()

	go func() {
		w.Write([]byte(`[1,2,{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1}]`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Equal(t, 3, len(raws))
	assert.Nil(t, raws[0])
	assert.Nil(t, raws[1])
	assert.Equal(t, "first", raws[2].Method)
	assert.True(t, isBatch)
	assert.Nil(t, err)

	reset()

	// broken array
	go func() {
		w.Write([]byte(`[1,2,3}`))
	}()
	raws, isBatch, err = e.ResolveBatch(ctx)
	assert.Nil(t, raws)
	assert.False(t, isBatch)
	assert.IsType(t, &json.SyntaxError{}, err)
}

func TestTrimReader(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("          \n       {\"json\":\"content\"}"))
	firstRune, err := trimReader(br)
	assert.Equal(t, '{', firstRune)
	assert.Nil(t, err)

	p := make([]byte, 30)
	n, err := br.Read(p)
	assert.Nil(t, err)
	assert.Equal(t, "{\"json\":\"content\"}", string(p[:n]))

	_, err = trimReader(br)
	assert.Equal(t, io.EOF, err)

	br = bufio.NewReader(strings.NewReader("          \n       [{\"json\":\"content\"}]"))
	firstRune, err = trimReader(br)
	assert.Equal(t, '[', firstRune)
	assert.Nil(t, err)
}
