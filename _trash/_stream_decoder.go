package jrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/pkg/errors"
)

/*
TODO:
- Resetの実装
	- Bufferedは全て破棄でよい？
- Decodeを、すべて戻り値とする仕様に変更
	- internal error, readruneに関わるところ、排除したい
	- そもそもまず、ReadByteに置き換えたい

*/

func newErrorRequest(errtype string, err error) *Request {
	return &Request{
		Version: "2.0",
		Method:  errtype,
		err:     err,
	}
}

// Decoder is
type Decoder struct {
	ctx   context.Context
	m     sync.Mutex
	err   error
	dirty bool
	buf   []byte

	r   *bufio.Reader
	dec *json.Decoder
}

// NewDecoder is
func NewDecoder(r io.Reader) *Decoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Decoder{
		ctx: context.Background(),
		r:   br,
		dec: json.NewDecoder(br),
		buf: make([]byte, peekLength),
	}
}

// Decode is
func (d *Decoder) Decode(ctx context.Context) ([]*Request, bool, error) {
	if d.err != nil {
		return nil, false, d.err
	}

	if ctx == nil {
		ctx = context.Background()
	} else {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		default:
		}
	}

	var requests = make([]*Request, 0, defaultBatchCapacity)
	var batch bool
	var err error
	defer func() { d.err = err }() // do not store context error

	done := make(chan struct{})
	go func() {
		d.m.Lock()
		defer func() {
			close(done)
			d.m.Unlock()
		}()

		batch, err = d.confirmIsBatch()
		if err != nil {
			// Unread contention
			if ue, ok := err.(*unreadError); ok {
				requests = append(requests, newErrorRequest(rpcInternalError, ue.err))
				err = errors.Wrap(err, "jrpc") // エラーレスポンスを返して、エラーも返す。パッケージ使用者はどのように対応するのか？
				// エラーが返されても、len(requests) > 0 ならレスポンスの必要があるということにする？
			} else {
				// or stream error
				requests = nil
			}
			return
		}
		if batch {
			_, err = d.dec.Token()
			if err != nil {
				// Token error
				if se, ok := err.(*json.SyntaxError); ok {
					requests = append(requests, newErrorRequest(rpcParseError, se))
					err = nil
				} else {
					// or stream error
					requests = nil
				}
				return
			}

			var req *Request
			for d.dec.More() {
				req = &Request{}
				err = d.dec.Decode(&req)
				if err != nil {
					switch err.(type) {
					// invalid syntax(maybe not be called)
					case *json.SyntaxError:
						requests = append(requests[:0], newErrorRequest(rpcParseError, err))
						batch = false
						return
					// individual type error
					case *json.UnmarshalTypeError:
						requests = append(requests, newErrorRequest(rpcInvalidRequest, err))
						err = nil
					default:
						if err == io.ErrUnexpectedEOF {
							requests = append(requests[:0], newErrorRequest(rpcParseError, err))
							batch = false
							err = nil
						} else {
							// or other stream error
							requests = nil
							return
						}
					}
					continue
				}
				requests = append(requests, req)
			}
			_, err = d.dec.Token()
			if err != nil {
				// Token error
				if se, ok := err.(*json.SyntaxError); ok {
					requests = append(requests, newErrorRequest(rpcParseError, se))
					err = nil
				} else {
					// or stream error
					requests = nil
				}
				return
			}
		} else {
			req := &Request{}
			err = d.dec.Decode(req)
			if err != nil {
				switch err.(type) {
				case *json.SyntaxError:
					req = newErrorRequest(rpcParseError, err)
				case *json.UnmarshalTypeError:
					req = newErrorRequest(rpcInvalidRequest, err)
				default:
					if err == io.ErrUnexpectedEOF {
						req = newErrorRequest(rpcParseError, err)
					} else {
						// or stream error
						requests = nil
						return
					}
				}
				err = nil
			}
			requests = append(requests, req)
		}
	}()

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case <-done:
		return requests, batch, err
	}
	return nil, false, nil
}

func (d *Decoder) confirmIsBatch() (isBatch bool, err error) {
	var r rune
	if d.dirty {
		buf := d.dec.Buffered()
		rs, ok := buf.(io.RuneScanner)
		if !ok {
			rs = bufio.NewReader(buf)
		}
		r, err = trimReader(rs)
		if err == io.EOF { // bytes.Reader or buio.Reader: error is EOF or ErrInvalidUnreadRune
			d.dirty = false
		} else if err != nil { // Unread error
			err = &unreadError{err: bufio.ErrInvalidUnreadRune}
			return
		}
	}
	if !d.dirty {
		r, err = trimReader(d.r) // ここでeofが出るとして、素直に返してしまう
		if err != nil {
			return
		}
		d.dirty = true
	}

	isBatch = r == '['
	return
}

// skip whitespaces, get first rune
func trimReader(rs io.RuneScanner) (r rune, err error) {
	for {
		r, _, err = rs.ReadRune()
		if err != nil {
			return // EOF or other connection errors (matter of streaming)
		} else if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			continue
		}
		err = rs.UnreadRune()
		if err != nil {
			err = &unreadError{err: err}
			return // Internal Error (bufio.Reader operation contention)
		}
		break
	}
	return
}

type unreadError struct {
	err error
}

func (ue *unreadError) Error() string {
	return ue.err.Error()
}

// Buffered returns a reader of the data remaining in the internal json.Decoder's buffer. The reader is valid until the next call to Decode.
func (d *Decoder) Buffered() io.Reader {
	return d.dec.Buffered()
}

// Reset not implemented yet
func (d *Decoder) Reset(r io.Reader) {
	d.err = nil
	d.dirty = false
	d.r.Reset(r)
	d.dec = json.NewDecoder(r)
}
