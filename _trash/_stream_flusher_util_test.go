package jrpc

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCountWriter(t *testing.T) {
	w := &countWriter{Writer: bufferpool.Get()}

	done1 := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			w.Write([]byte("1"))
		}
		close(done1)
	}()

	done2 := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			w.Write([]byte("2"))
		}
		close(done2)
	}()

	select {
	case <-done1:
		select {
		case <-done2:
		}
	}

	assert.Equal(t, 1000, int(w.Count))
}

func TestCtxCopy(t *testing.T) {

	// basics
	assert.NotNil(t, ctxCopy(nil, nil, nil, nil))

	makeSrc := func(ctx context.Context, w io.WriteCloser) {
		for {
			select {
			case <-ctx.Done():
				w.Close()
				return
			default:
				_, err := w.Write([]byte("a"))
				if err != nil {
					return
				}
			}
		}
	}

	t.Run("no cancel(success)", func(t *testing.T) {
		src, w := io.Pipe()
		srcWriter := &countWriter{Writer: w}
		dst := bufferpool.Get()

		makeCtx, c2 := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer c2()
		go makeSrc(makeCtx, srcWriter)

		err := ctxCopy(nil, dst, src, make([]byte, 200))
		assert.Nil(t, err)

		<-makeCtx.Done()

		assert.Equal(t, int64(dst.Len()), srcWriter.Count)
	})

	t.Run("with cancel", func(t *testing.T) {
		ctx, c1 := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer c1()
		src, w := io.Pipe()
		srcWriter := &countWriter{Writer: w}
		dst := bufferpool.Get()

		makeCtx, c2 := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer c2()
		go makeSrc(makeCtx, srcWriter)

		err := ctxCopy(ctx, dst, src, make([]byte, 200))
		assert.Equal(t, ctx.Err(), err)

		io.Copy(ioutil.Discard, src)
		<-makeCtx.Done()

		assert.True(t, int64(dst.Len()) < srcWriter.Count)
	})

	t.Run("reader error", func(t *testing.T) {
		ERR := errors.New("reader dummy error")

		ctx, c1 := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer c1()
		src, w := io.Pipe()
		srcWriter := &countWriter{Writer: w}
		dst := bufferpool.Get()

		makeCtx, c2 := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer c2()
		go func() {
			for {
				select {
				case <-makeCtx.Done():
					w.CloseWithError(ERR)
					return
				default:
					_, err := srcWriter.Write([]byte("a"))
					if err != nil {
						return
					}
				}
			}
		}()

		err := ctxCopy(ctx, dst, src, make([]byte, 200))
		assert.Equal(t, ERR, err)

		<-makeCtx.Done()

		assert.Equal(t, int64(dst.Len()), srcWriter.Count)
	})

	t.Run("writer error", func(t *testing.T) {
		ERR := errors.New("writer dummy error")

		ctx, c1 := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer c1()
		src, w := io.Pipe()
		srcWriter := &countWriter{Writer: w}
		dst := &errWriter{err: ERR}

		makeCtx, c2 := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer c2()
		go makeSrc(makeCtx, srcWriter)

		err := ctxCopy(ctx, dst, src, make([]byte, 200))
		assert.Equal(t, ERR, err)

		src.Close()
		<-makeCtx.Done()

		assert.Equal(t, int64(dst.c), srcWriter.Count)
	})

	t.Run("short write error", func(t *testing.T) {
		ERR := errors.New("writer dummy error")

		ctx, c1 := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer c1()
		src, w := io.Pipe()
		srcWriter := &countWriter{Writer: w}
		dst := &errWriter{err: ERR, shortFlag: true}

		makeCtx, c2 := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer c2()
		go makeSrc(makeCtx, srcWriter)

		err := ctxCopy(ctx, dst, src, make([]byte, 200))
		assert.Equal(t, io.ErrShortWrite, err)

		src.Close()
		<-makeCtx.Done()
	})
}

type errWriter struct {
	c         int
	err       error
	shortFlag bool
}

func (e *errWriter) Write(b []byte) (int, error) {
	e.c += len(b)
	if e.c > 5000 {
		if e.shortFlag {
			return 0, nil
		}
		return len(b), e.err
	}
	return ioutil.Discard.Write(b)
}
