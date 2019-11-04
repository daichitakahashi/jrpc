package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDecoder_Decode(t *testing.T) {
	type TestCase struct {
		src              string
		checkReturnValue func([]*Request, bool, error)
		checkStoredError func(error)
		needReset        bool
		desc             string
	}

	dec := NewDecoder(nil)

	t.Run("unary", func(t *testing.T) {
		r, w := io.Pipe()
		dec.Reset(r)

		testcases := []TestCase{
			{
				src: `{"jsonrpc":"2.0","method":"sample.text","id":123}`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, NewID(123), reqs[0].ID)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.NoError(t, err)
				},
				desc: "unary success",
			}, {
				src: `{"jsonrpc":"2.0","method":999999,"id":123}`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcInvalidRequest, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.NoError(t, err)
				},
				desc: "unary type error",
			}, {
				src: `{"jsonrpc":"2.0","method":sample.method,"id":123}`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcParseError, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.IsType(t, &json.SyntaxError{}, err)
				},
				desc: "unary parse error",
			},
		}
		for _, testcase := range testcases {
			t.Run(testcase.desc, func(t *testing.T) {
				go func() {
					w.Write([]byte(testcase.src))
				}()
				testcase.checkReturnValue(
					dec.Decode(make([]*Request, 0, 1)),
				)
				testcase.checkStoredError(dec.Err())
			})
		}
		w.Close()
		r.Close()

		t.Run("stored error", func(t *testing.T) {
			reqs, batch, err := dec.Decode(make([]*Request, 0, 1))
			require.Len(t, reqs, 0)
			require.False(t, batch)
			require.Error(t, err)

			require.IsType(t, &json.SyntaxError{}, err)
		})
	})

	t.Run("batch", func(t *testing.T) {
		r, w := io.Pipe()
		dec.Reset(r)

		testcases := []TestCase{
			{
				src: `[
					{"jsonrpc":"2.0","method":"sample.batch","id":1},
					{"jsonrpc":"2.0","method":"sample.batch","id":2}
				][
					{"jsonrpc":"2.0","method":"999999","id":9999}
				]`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 2)
					require.Equal(t, NewID(1), reqs[0].ID)
					require.Equal(t, NewID(2), reqs[1].ID)
					require.True(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.NoError(t, err)
				},
				desc: "batch success",
			}, {
				src: ``,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, NewID(9999), reqs[0].ID)
					require.True(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.NoError(t, err)
				},
				desc: "batch success but len(reqs) == 1",
			}, {
				src: `  `,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 0)
					require.False(t, batch)
					require.Error(t, err)
				},
				checkStoredError: func(err error) {
					require.Equal(t, err, io.EOF)
				},
				needReset: true,
				desc:      "batch success but empty",
			}, {
				src: `[
					{"jsonrpc":"2.0","method":"1111111111","id":1},
					{"jsonrpc":"2.0","method":"2222222222","id":2},
					{"jsonrpc":"2.0","method":3333333333,"id":3},
					{"jsonrpc":"2.0","method":"4444444444","id":4},
					{"jsonrpc":"2.0","method":"5555555555","id":5}
				]`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 5)
					require.Equal(t, rpcInvalidRequest, reqs[2].Method)
					require.True(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.NoError(t, err)
				},
				desc: "batch success but one type error",
			}, {
				src: `[
					{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1},
					{"jsonrpc":"2.0","method":"ssssssssss","id":2},
					{"jsonrpc":"2.0","method":dddddddddd,"id":3},
					{"jsonrpc":"2.0","method":"ffffffffff","id":4},
					{"jsonrpc":"2.0","method":"gggggggggg","id":5}
				]`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcParseError, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.IsType(t, &json.SyntaxError{}, err)
				},
				needReset: true,
				desc:      "batch parse error",
			}, {
				src: `[
					{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1},
					{"jsonrpc":"2.0","method":"ssssssssss","id":2},
				)`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcParseError, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.IsType(t, &json.SyntaxError{}, err)
				},
				needReset: true,
				desc:      "batch parse error(not finished correctly)",
			}, {
				src: `[
					{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1},
					{"jsonrpc":"2.0","method":"ssssssssss","id":2}
				`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcParseError, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.Equal(t, io.EOF, err)
				},
				needReset: true,
				desc:      "batch parse error(not finished correctly, ver. io.EOF)",
			}, {
				src: `[
					{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1},
					{"jsonrpc":"2.0","method":
				`,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 1)
					require.Equal(t, rpcParseError, reqs[0].Method)
					require.False(t, batch)
					require.NoError(t, err)
				},
				checkStoredError: func(err error) {
					require.Equal(t, err, io.ErrUnexpectedEOF)
				},
				needReset: true,
				desc:      "batch parse error(not finished correctly, ver. io.ErrUnexpectedEOF)",
			}, {
				src: ``,
				checkReturnValue: func(reqs []*Request, batch bool, err error) {
					require.Len(t, reqs, 0)
					require.False(t, batch)
					require.Error(t, err)
				},
				checkStoredError: func(err error) {
					require.Equal(t, err, io.EOF)
				},
				needReset: true,
				desc:      "EOF",
			},
		}
		for _, testcase := range testcases {
			t.Run(testcase.desc, func(t *testing.T) {
				go func() {
					w.Write([]byte(testcase.src))
					if testcase.needReset {
						w.Close()
					}
				}()
				time.Sleep(time.Millisecond)
				testcase.checkReturnValue(
					dec.Decode(make([]*Request, 0, 3)),
				)
				testcase.checkStoredError(dec.Err())
				if testcase.needReset {
					r, w = io.Pipe()
					dec.Reset(r)
				}
			})
		}
		r.Close()
	})
}

func TestDecoder_DecodeContext(t *testing.T) {
	dec := NewDecoder(nil)

	reqs := make([]*Request, 0, 1)
	var batch bool
	var err error

	t.Run("success", func(t *testing.T) {
		dec.Reset(strings.NewReader(`
			{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1}
		`))
		reqs, batch, err = dec.DecodeContext(context.Background(), reqs)
		require.Len(t, reqs, 1)
		require.False(t, batch)
		require.NoError(t, err)
	})

	t.Run("canceled in advance", func(t *testing.T) {
		dec.Reset(strings.NewReader(`
			{"jsonrpc":"2.0","method":"aaaaaaaaaa","id":1}
		`))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		reqs, batch, err = dec.DecodeContext(ctx, reqs)
		require.Len(t, reqs, 0)
		require.False(t, batch)
		require.Equal(t, err, ctx.Err())
	})

	t.Run("canceled in halfway", func(t *testing.T) {
		r, w := io.Pipe()
		dec.Reset(r)

		go func() {
			w.Write([]byte(`{"jsonrpc":"2.0",`))
			time.Sleep(time.Millisecond * 10)
			w.Write([]byte(`"method":"aaaaaaaaaa","id":1}`))
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
		defer cancel()
		reqs, batch, err = dec.DecodeContext(ctx, reqs)
		require.Len(t, reqs, 0)
		require.False(t, batch)
		require.Equal(t, err, ctx.Err())
	})
}

func TestDecoder_HandleError(t *testing.T) {
	dec := NewDecoder(nil)
	reqs := make([]*Request, 0, 5)
	var batch bool
	var err error

	t.Run("*json.SyntaxError", func(t *testing.T) {
		reqs = append(reqs[:0], &Request{}, &Request{})
		batch = true
		err = &json.SyntaxError{}

		dec.Reset(nil)
		reqs, batch, err = dec.handleError(err, reqs, batch)

		require.Len(t, reqs, 1)
		require.False(t, batch)
		require.Nil(t, err)

		require.Equal(t, rpcParseError, reqs[0].Method)

		require.Error(t, dec.Err())
	})

	t.Run("*json.UnmarshalTypeError", func(t *testing.T) {
		reqs = append(reqs[:0], &Request{}, &Request{})
		batch = true
		err = &json.UnmarshalTypeError{}

		dec.Reset(nil)
		reqs, batch, err = dec.handleError(err, reqs, batch)

		require.Len(t, reqs, 3)
		require.True(t, batch)
		require.NoError(t, err)

		require.Equal(t, rpcInvalidRequest, reqs[2].Method)

		require.NoError(t, dec.Err())
	})

	t.Run("io.ErrUnexpectedEOF", func(t *testing.T) {
		reqs = append(reqs[:0], &Request{}, &Request{})
		batch = true
		err = io.ErrUnexpectedEOF

		dec.Reset(nil)
		reqs, batch, err = dec.handleError(err, reqs, batch)

		require.Len(t, reqs, 1)
		require.False(t, batch)
		require.NoError(t, err)

		require.Equal(t, rpcParseError, reqs[0].Method)

		require.Error(t, dec.Err())
	})

	t.Run("other", func(t *testing.T) {
		reqs = append(reqs[:0], &Request{}, &Request{})
		batch = true
		err = errors.New("dummy error")

		dec.Reset(nil)
		reqs, batch, err = dec.handleError(err, reqs, batch)

		require.Len(t, reqs, 0)
		require.False(t, batch)
		require.Error(t, err)

		require.Equal(t, err, dec.Err())
	})
}

func TestDecoder_FirstByte(t *testing.T) {
	// ONLY ERROR OR MALFORMED PATTERN
	t.Run("EOF", func(t *testing.T) {
		src := strings.NewReader(`{"jsonrpc":"2.0","method":"sample.text","id":123}                                       `)
		dec := NewDecoder(src)

		char, err := dec.firstByte()
		require.Equal(t, byte('{'), char)
		require.NoError(t, err)

		require.True(t, dec.dirty)

		dec.Decode(make([]*Request, 0, 0))

		char, err = dec.firstByte()
		require.Equal(t, byte(0), char)
		require.Equal(t, err, io.EOF)
	})

	t.Run("too long space", func(t *testing.T) {
		src := strings.NewReader(`{"jsonrpc":"2.0","method":"too.long.space","id":123}                    ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
                                                                                                               ` + `
		`)
		dec := NewDecoder(src)

		char, err := dec.firstByte()
		require.Equal(t, byte('{'), char)
		require.NoError(t, err)

		require.True(t, dec.dirty)

		dec.Decode(make([]*Request, 0, 0))

		char, err = dec.firstByte()
		require.Equal(t, byte(0), char)
		require.Equal(t, err, io.EOF)
	})
}

func TestCalibrate(t *testing.T) {
	t.Run("under capacity", func(t *testing.T) {
		reqs := []*Request{&Request{}, &Request{}, &Request{}, &Request{}, &Request{}}
		capacity := 10

		reqs = Calibrate(reqs, capacity)

		require.Len(t, reqs, 0)
		require.Equal(t, 5, cap(reqs))
	})

	t.Run("over capacity", func(t *testing.T) {
		reqs := []*Request{&Request{}, &Request{}, &Request{}, &Request{}, &Request{}, &Request{}, &Request{}, &Request{}, &Request{}, &Request{}}
		capacity := 5

		reqs = Calibrate(reqs, capacity)

		require.Len(t, reqs, 0)
		require.Equal(t, 5, cap(reqs))
	})
}
