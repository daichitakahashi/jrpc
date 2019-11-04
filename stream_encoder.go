package jrpc

import (
	"bufio"
	"context"
	"io"
	"sync"
)

/*
TODO:

*/

// Encoder is
type Encoder struct {
	dst *bufio.Writer
	m   sync.Mutex
}

// NewEncoder is
func NewEncoder(dst io.Writer) *Encoder {
	return &Encoder{
		dst: bufio.NewWriter(dst),
	}
}

// NewEncoderSize is
func NewEncoderSize(dst io.Writer, bufSize int) *Encoder {
	return &Encoder{
		dst: bufio.NewWriterSize(dst, bufSize),
	}
}

// Encode is
func (enc *Encoder) Encode(resps []*Response, batch bool) error {
	return enc.EncodeContext(nil, resps, batch)
}

// EncodeContext is
func (enc *Encoder) EncodeContext(ctx context.Context, resps []*Response, batch bool) (err error) {
	enc.m.Lock()
	buf := bufferpool.Get()
	defer func() {
		buf.Free()
		enc.m.Unlock()
	}()
	if len(resps) == 0 {
		return nil
	}

	if batch {
		buf.AppendByte('[')
	} else {
		resps = resps[:1]
	}

	lastIdx := len(resps) - 1
	for i := range resps {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		err = resps[i].encodeTo(buf)
		if err != nil {
			return
		}

		if batch {
			if i != lastIdx {
				buf.AppendByte(',')
			} else {
				buf.AppendString("]\n")
			}
		} else {
			buf.AppendByte('\n')
		}
		_, err = enc.dst.Write(buf.Bytes())
		buf.Reset()
		if err != nil {
			return
		}
	}
	return enc.dst.Flush()
}

// Reset is
func (enc *Encoder) Reset(dst io.Writer) {
	enc.dst.Reset(dst)
}
