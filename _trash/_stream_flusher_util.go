package jrpc

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
)

type countWriter struct {
	Count int64
	io.Writer
}

func (w *countWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.Count += int64(n)
	return
}

func (w *countWriter) Close() error {
	if c, ok := w.Writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func ctxCopy(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) error {
	if len(buf) == 0 {
		return errors.New("jrpc: internal: empty buffer for copy")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var cancel int32
	e := make(chan error)

	go func() {
		var rn, wn int
		var err error

		for {
			if atomic.LoadInt32(&cancel) == 1 {
				close(e)
				return
			}

			rn, err = src.Read(buf)
			if err != nil {
				if err != io.EOF {
					e <- err
					return
				}
				break
			}

			wn, err = dst.Write(buf[:rn])
			if err != nil {
				e <- err
				return
			}

			if rn != wn {
				e <- io.ErrShortWrite
				return
			}
		}
		e <- nil
	}()

	select {
	case err := <-e:
		return err
	case <-ctx.Done():
		atomic.StoreInt32(&cancel, 1)
		<-e
		return ctx.Err()
	}
}
