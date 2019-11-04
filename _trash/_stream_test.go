package jrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/daichitakahashi/httpjrpc/jrpc/jrpcmock"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	type testcase struct {
		request  string
		response string
		desc     string
	}

	testCases := []testcase{
		{},
	}

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	var client, server net.Conn

	go func() {
		defer ln.Close()
		server, _ = ln.Accept()
	}()
	client, _ = net.Dial("tcp", ln.Addr().String())
	defer server.Close()
	defer client.Close()

	e := jrpcmock.Repository().Invoke(server)
	// e := newExecutor(jrpcmock.Repository(), server)
	buf := bytes.NewBuffer(nil)

	for _, c := range testCases {
		buf.Reset()
		client.Write([]byte(c.request))
		e.Execute(nil, buf)
		assert.Equal(t, c.response, buf.String(), c.desc)
	}
}

func TestExecuteRuntime(t *testing.T) {
	type testcase struct {
		request string
		test    func(client, server net.Conn, e *executor)
	}

	testCases := []testcase{
		{
			request: arr("", ""),
			test:    func(client, server net.Conn, e *executor) {},
		},
	}

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()

	done := make(chan struct{})
	cli := make(chan net.Conn, 1)
	semaphore := make(chan struct{}, 1)
	semaphore <- struct{}{}

	go func() {
		for {
			select {
			case <-done:
				return
			case semaphore <- struct{}{}:
				cl, err := net.Dial("tcp", ln.Addr().String())
				if err != nil {
					return
				}
				cli <- cl
			}
		}
	}()

	var client, server net.Conn
	mockRepo := mock.Repository()

	for _, c := range testCases {
		func() {
			<-semaphore
			server, _ = ln.Accept()
			defer server.Close()
			client <- cli
			defer client.Close()

			e := initExecutor(mockRepo, server)
			c.test(client, server, e)
		}()
	}
	close(done)
}

func arr(jsonstr ...string) string {
	return "[" + strings.Join(jsonstr, "\n,") + "]"
}

type (
	mock struct {
		test func(context.Context, *json.RawMessage)
	}

	Params struct {
		ArgsString string `json:"string"`
		ArgsInt    int    `json:"int"`
	}

	Result struct {
		Result string `json:"result"`
	}
)

func (m *mock) loadMetadata(method string) (*Metadata, bool) {
	methods := []string{"echo", "hoge", "piyo"}
	var found bool
	for _, name := range methods {
		if method == name {
			found = true
			break
		}
	}
	if !found {
		return nil, false
	}

	return &Metadata{
		Handler: HandlerFunc(func(ctx context.Context, params *json.RawMessage) (interface{}, *Error) {
			var p Params
			if params == nil {
				return nil, ErrInvalidParams()
			}
			err := json.Unmarshal(*params, &p)
			if err != nil {
				return nil, ErrInvalidParams()
			}

			if m.test != nil {
				m.test(ctx, params)
			}

			if method == "hoge" {
				time.Sleep(time.Millisecond * 2)
			} else if method == "piyo" {
				time.Sleep(time.Millisecond * 4)
			}

			result := Result{
				Result: fmt.Sprintf("method: %s, msg: %s, %d", method, p.ArgsString, p.ArgsInt),
			}
			return result, nil
		}),
		Params: Params{},
		Result: Result{},
	}, true
}

func TestExecute(t *testing.T) {
	//ctx := context.Background()
	m := &mock{}

	w, r := net.Pipe()
	executor := newExecutor(m, r)
	buf := bytes.NewBuffer(nil)

	write := func(str string) {
		go func() {
			w.Write([]byte(str))
		}()
	}
	/*
		writeClose := func(str string) {
			go func() {
				w.Write([]byte(str))
				w.Close()
			}()
		}*/

	write(`{
		"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id": 1}
	`)
	n, err := executor.Execute(nil, buf)
	assert.Nil(t, err)
	result := buf.String()
	assert.Equal(t, len([]byte(result)), n)
	result = result[:len(result)-1]
	assert.Equal(t, `{"jsonrpc":"2.0","result":{"result":"method: echo, msg: echoecho, 99"},"id":1}`, result)

	buf.Reset()
	write(`{
		"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id":null}
	`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, "", result)
	assert.Equal(t, 0, n)

	buf.Reset()
	write(`[
		{"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id":1},
		{"jsonrpc": "2.0", "method": "hoge", "params": {"string":"hogehoge","int":22}, "id":2}
]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	assert.Equal(t, `[{"jsonrpc":"2.0","result":{"result":"method: echo, msg: echoecho, 99"},"id":1}
,{"jsonrpc":"2.0","result":{"result":"method: hoge, msg: hogehoge, 22"},"id":2}
]`, result)

	buf.Reset()
	write(`[
		{"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id":null},
		{"jsonrpc": "2.0", "method": "hoge", "params": {"string":"hogehoge","int":22}}
]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	assert.Equal(t, ``, result)

	buf.Reset()
	write(`[
		{"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id":1},
		{"jsonrpc": "2.0", "method": "hoge", "params": {"string":"hogehoge","int":22}}
]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	assert.Equal(t, `[{"jsonrpc":"2.0","result":{"result":"method: echo, msg: echoecho, 99"},"id":1}
]`, result)

	buf.Reset()
	write(`[]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	result = result[:len(result)-1]
	assert.Equal(t, `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}`, result)

	buf.Reset()
	write(`[
		{"jsonrpc": "2.0", "method": "echo", "params": {"string":"echoecho","int":99}, "id":1},
		{"jsonrpc": "2.0", "method": "hoge", "params": {"string":"hogehoge","int":22}},
		{"jsonrpc": "2.0", "method": "piyo", "params": {"string":"piyopiyo","int":44}, "id":3}
]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	assert.Equal(t, `[{"jsonrpc":"2.0","result":{"result":"method: echo, msg: echoecho, 99"},"id":1}
,{"jsonrpc":"2.0","result":{"result":"method: piyo, msg: piyopiyo, 44"},"id":3}
]`, result)

	buf.Reset()
	write(`[1,2,3]`)
	n, err = executor.Execute(nil, buf)
	assert.Nil(t, err)
	result = buf.String()
	assert.Equal(t, len([]byte(result)), n)
	assert.Equal(t, `[{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}
,{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}
,{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request"},"id":null}
]`, result)

	// パースエラー
	// パースエラーの書き込みエラー
	// eof, unexpectedeofの際のエラー、空レスポンス

	// [1,2,3]

	// エラー一つで他キャンセル
}

/*
func TestParseRequestUnary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//defer cancel()

	w, r := net.Pipe()
	m := &mock{}
	executor := newExecutor(m, r)

	go func() {
		w.Write([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}{
	`))
	}()

	requests, needResponse, respErr, err := executor.ParseRequest(ctx)
	assert.Equal(t, 1, len(requests))
	assert.NotZero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	param := make([]int, 2)
	assert.Nil(t, fastjson.Unmarshal(*requests[0].Params, &param))
	assert.Equal(t, []int{42, 23}, param)

	var id int
	assert.Nil(t, fastjson.Unmarshal(*requests[0].ID, &id))
	assert.Equal(t, 1, id)

	go func() {
		w.Write([]byte(`
		"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
	`))
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 1, len(requests))
	assert.NotZero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	go func() {
		w.Write([]byte(`
		{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": null}
	`))
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 1, len(requests))
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	go func() {
		w.Write([]byte(`{`))
		w.Close()
	}()
	_, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Zero(t, needResponse)
	assert.Equal(t, ErrParse(), respErr)
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	r.Close()

	w, r = net.Pipe()
	executor = newExecutor(m, r)

	w.Close()
	_, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Equal(t, io.EOF, err)

	r.Close()
	_, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Equal(t, io.ErrClosedPipe, err)

	w, r = net.Pipe()
	executor = newExecutor(m, r)

	go func() {
		w.Write([]byte(`{jsonrpc": "2.0, "methosubtract", "params": [42, 23], "id": 1}`))
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Nil(t, requests)
	assert.Zero(t, needResponse)
	assert.Equal(t, ErrParse(), respErr)
	assert.Nil(t, err)
	w.Close()
	r.Close()

	cancel()
	_, _, respErr, err = executor.ParseRequest(ctx)
	assert.Nil(t, respErr)
	assert.Equal(t, context.Canceled, err)
}

func TestParseRequestArray(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	w, r := net.Pipe()
	m := &mock{}
	executor := newExecutor(m, r)

	go func() {
		w.Write([]byte(`[
		{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1},
		{"jsonrpc": "2.0", "method": "second", "id":null},
		{"jsonrpc": "2.0", "method": "third", "id": 3}
		]`))
	}()
	requests, needResponse, respErr, err := executor.ParseRequest(ctx)
	assert.Equal(t, 3, len(requests))
	assert.NotZero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	var i int
	for _, r := range requests {
		if r.ID != nil {
			i++
		}
	}
	assert.Equal(t, 2, i)

	go func() {
		w.Write([]byte(`[
		{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": null},
		{"jsonrpc": "2.0", "method": "second"},
		{"jsonrpc": "2.0", "method": "third"}
		]      []      [*]`))
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 3, len(requests))
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 0, len(requests))
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Nil(t, requests)
	assert.Zero(t, needResponse)
	assert.Equal(t, ErrParse(), respErr)
	assert.Nil(t, err)

	w.Close()
	r.Close()

	w, r = net.Pipe()
	executor = newExecutor(m, r)
	go func() {
		w.Write([]byte(`[
		{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1},
		{"jsonrpc": "2.0", "method": "second"},
		{"jsonrpc": "2.0", "method": "third", "id": 3}
		]
		{"jsonrpc": "2.0", "method": "second"}`))
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 3, len(requests))
	assert.NotZero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Equal(t, 1, len(requests))
	assert.Zero(t, needResponse)
	assert.Nil(t, respErr)
	assert.Nil(t, err)

	w.Close()
	r.Close()

	w, r = net.Pipe()
	executor = newExecutor(m, r)
	go func() {
		w.Write([]byte(`[
		{"jsonrpc": "2.0", "method": "first", "params": [42, 23], "id": 1},
		{"jsonrpc": "2.0", "met`))
		w.Close()
	}()
	requests, needResponse, respErr, err = executor.ParseRequest(ctx)
	assert.Nil(t, requests)
	assert.Zero(t, needResponse)
	assert.Equal(t, ErrParse(), respErr)
	assert.Equal(t, io.ErrUnexpectedEOF, err)

	r.Close()
}
*/
