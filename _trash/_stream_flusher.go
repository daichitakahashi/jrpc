package jrpc

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Flusher is
type Flusher struct {
	dst   io.Writer
	resps array

	index int // index of resps
	rb    int // position of resps[i].Bytes

	buf    []byte
	bufLen int // buffer size
	bb     int // position of buf
}

// NewFlusher is
func NewFlusher(dst io.Writer, resps []*Response, batch bool) *Flusher {
	buffers := make([]*Buffer, len(resps))
	for i, r := range resps {
		buf, err := r.encode()
		if err != nil {
			r.Result = nil
			r.Error = ErrInternal(nil)
			buf, _ = r.encode()
		}
		buffers[i] = buf
	}
	return NewFlusherFromBuffers(dst, buffers, batch)
}

// NewFlusherFromBuffers is. This not make sure if content of buffer is valid json
func NewFlusherFromBuffers(dst io.Writer, buffers []*Buffer, batch bool) *Flusher {
	f := &Flusher{
		dst: dst,
	}
	if batch {
		f.resps = newJSONArray(buffers)
	} else {
		f.resps = newSingleJSON(buffers[0])
	}
	return f
}

const defaultBufferSize = 1024 * 4

// Flush is
func (f *Flusher) Flush(ctx context.Context) (n int64, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	default:
	}
	if f.resps == nil {
		return 0, errors.New("jrpc: responses already run out")
	} else if f.resps.len() == 0 {
		return
	}

	w := &countWriter{Writer: f.dst}
	defer func() {
		n = w.Count
	}()

	if f.buf == nil {
		f.buf = make([]byte, defaultBufferSize)
		f.bufLen = defaultBufferSize
	}

	var wn int

LOOP:
	for f.resps.len() > f.index {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			if f.fill() {
				wn, err = w.Write(f.buf)
				if wn != f.bufLen || err != nil {
					if err == nil {
						err = io.ErrShortWrite
					}
					f.bb = copy(f.buf, f.buf[wn:])
					return
				}
			} else if f.index == f.resps.len() {
				break LOOP
			} else {
				continue
			}
		}
	}

	// flush the rest
	if f.bb > 0 {
		wn, err = w.Write(f.buf[:f.bb])
		if wn != f.bb || err != nil {
			if err == nil {
				err = io.ErrShortWrite
			}
			f.bb = copy(f.buf, f.buf[wn:])
			return
		}
	}
	// done
	f.resps.free()
	f.resps = nil
	f.buf = nil
	return
}

// Encode is option to catch invalid requests when encoding
func Encode(resps []*Response) ([]*Buffer, Reports) {
	buffers := make([]*Buffer, len(resps))
	var report Reports

	for i, r := range resps {
		buf, err := r.encode()
		if err != nil {
			r.Result = nil
			r.Error = ErrInternal(nil)
			report.append(i, err)
			buf, _ = r.encode()
		}
		buffers = append(buffers, buf)
	}
	return buffers, report
}

func (f *Flusher) available() int {
	return f.bufLen - f.bb
}

func (f *Flusher) fill() (overflow bool) {
	resp := f.resps.at(f.index)[f.rb:]
	avail := f.available()
	if len(resp) > avail {
		n := copy(f.buf[f.bb:], resp)
		f.rb += n
		f.bb = 0
		overflow = true
	} else {
		n := copy(f.buf[f.bb:], resp)
		// f.resps[f.index].Free() // do not Free() for rewind
		f.rb = 0
		if avail == n {
			f.bb = 0
		} else {
			f.bb += n
		}
		f.index++
	}
	return
}

// SetBuffer is
func (f *Flusher) SetBuffer(b []byte) {
	if b == nil {
		return
	}
	newLength := len(b)
	if newLength == 0 {
		panic("jrpc: zero buffer")
	}
	if f.bb == 0 {
		f.buf = b
		f.bufLen = newLength
		return
	}

	copy(b, f.buf)
	if f.bb > newLength { // buffered　> new buffer
		f.rewind(f.bb - newLength)
		f.buf = b
	} else { // buffered < new buffer
		f.buf = b
	}
}

func (f *Flusher) rewind(n int) {
	if f.rb > n {
		f.rb -= n
		return
	}
	n -= f.rb
	f.rb = 0
	var l int
	for n > 0 {
		f.index--
		l = f.resps.lenAt(f.index)
		if l > n {
			f.rb = l - n
			return
		}
		n -= l
	}
	return
}

var (
	openBracket   = []byte{'['}
	comma         = []byte{','}
	closedBracket = []byte{']'}
	newLine       = []byte{'\n'}
)

type array interface {
	len() int
	at(n int) []byte
	lenAt(n int) int
	free()
}

type singleJSON struct {
	buf *Buffer
}

func newSingleJSON(buf *Buffer) *singleJSON {
	return &singleJSON{
		buf: buf,
	}
}

func (s *singleJSON) len() int {
	return 2
}

func (s *singleJSON) at(n int) []byte {
	switch n {
	case 0:
		return s.buf.Bytes()
	case 1:
		return newLine
	default:
		return nil
	}
}

func (s *singleJSON) lenAt(n int) int {
	if n == 0 {
		return s.buf.Len()
	}
	return 1
}

func (s *singleJSON) free() {
	s.buf.Free()
}

var _ array = (*singleJSON)(nil)

type jsonArray struct {
	arr    []*Buffer
	length int
}

func newJSONArray(arr []*Buffer) *jsonArray {
	return &jsonArray{
		arr:    arr,
		length: len(arr)*2 + 2,
	}
}

func (a *jsonArray) len() int {
	return a.length
}

func (a *jsonArray) at(n int) []byte {
	if n&1 == 1 { // odd
		if n == a.length-1 { // last
			return newLine
		}
		return a.arr[(n-1)/2].Bytes()
	}
	// even
	if n == 0 {
		return openBracket
	} else if n == a.length-2 {
		return closedBracket
	}
	return comma
}

func (a *jsonArray) lenAt(n int) int {
	if n&1 == 1 { // odd
		if n == a.length-1 { // last
			return 1
		}
		return a.arr[(n-1)/2].Len()
	}
	// even
	return 1
}

func (a *jsonArray) free() {
	for _, buf := range a.arr {
		buf.Free()
	}
}

func (a *jsonArray) freeAt(n int) {
	for _, buf := range a.arr {
		buf.Free()
	}
}

var _ array = (*jsonArray)(nil)

type (
	// Report is
	Report struct {
		Index int
		Err   error
	}

	// Reports is
	Reports []*Report
)

func (r *Report) Error() string {
	return fmt.Sprintf("index: %d, error: %s", r.Index, r.Err.Error())
}

func (rs *Reports) append(index int, err error) {
	if rs == nil {
		*rs = make(Reports, 0, 5)
	}
	*rs = append(*rs, &Report{
		Index: index,
		Err:   err,
	})
}
