package jrpc

import (
	"context"
	"io"
)

/*
TODO:

*/

// Encoder is
type Encoder struct {
	buf     []byte
	bufSize int
	dst     io.Writer
	state   struct {
		encoded *Buffer
		cur     int
		len     int
	}
}

// NewEncoder is
func NewEncoder(dst io.Writer, bufSize int) *Encoder {
	return &Encoder{
		buf:     make([]byte, bufSize),
		bufSize: bufSize,
		dst:     dst,
	}
}

// Encode is
func (enc *Encoder) Encode(resps []*Response, batch bool) error {
	return enc.EncodeContext(nil, resps, batch)
}

// EncodeContext is
func (enc *Encoder) EncodeContext(ctx context.Context, resps []*Response, batch bool) error {
	var bufCur int
	var err error

	state := struct {
		encoded *Buffer
		cur     int
		len     int
	}{}

	if !batch {
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
		state.encoded, err = resps[i].encode()
		if err != nil {
			return err
		}

		if batch {
			switch i {
			case 0:
				state.encoded.AppendPrefix('[')
			case lastIdx:
				state.encoded.AppendString("]\n")
				fallthrough
			default:
				state.encoded.AppendPrefix(',')
			}
		} else {
			state.encoded.AppendString("]\n")
		}

		state.cur = 0
		state.len = state.encoded.Len()

		fill := func() (overflow bool) {
			resp := state.encoded.Bytes()[state.cur:]
			avail := enc.bufSize - bufCur
			if len(resp) > avail {
				n := copy(enc.buf[bufCur:], resp)
				state.cur += n
				bufCur = 0
				overflow = true
			} else {
				n := copy(enc.buf[bufCur:], resp)
				state.encoded.Free()
				if avail == n {
					bufCur = 0
				} else {
					bufCur += n
				}
			}
			return
		}

	LOOP:
		for {
			if fill() {
				_, err = enc.dst.Write(enc.buf)
				if err != nil {
					state.encoded.Free()
					return err
				}
			} else {
				break LOOP
			}
		}
	}
	if bufCur > 0 {
		_, err = enc.dst.Write(enc.buf[:bufCur])
		if err != nil {
			return err
		}
	}
	return nil
}

// Reset is
func (enc *Encoder) Reset(dst io.Writer) {
	enc.dst = dst
}
