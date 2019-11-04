package dev

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerConnRead(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:8888")
	defer ln.Close()
	connect := func(t *testing.T, serverFunc func(t *testing.T, server net.Conn)) (client net.Conn) {

		go func() {
			server, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			defer server.Close()
			serverFunc(t, server)
		}()
		client, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			panic(err)
		}
		return
	}

	type testcase struct {
		desc       string
		serverFunc func(*testing.T, net.Conn)
		clientFunc func(*testing.T, net.Conn)
	}

	testcases := []testcase{
		{desc: "straight - net.Conn.Read will await",
			serverFunc: func(t *testing.T, server net.Conn) {
				time.Sleep(time.Millisecond)
				server.Write([]byte("after 1 millisecond\n"))
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				buf := make([]byte, 20)
				n, err := client.Read(buf) //io.ReadFull(client, buf)
				assert.Equal(t, 20, n)
				assert.Nil(t, err)
				assert.Equal(t, []byte("after 1 millisecond\n"), buf[:n])
			},
		}, {desc: "straight - bufio.Reader.ReadRune will await",
			serverFunc: func(t *testing.T, server net.Conn) {
				time.Sleep(time.Millisecond)
				server.Write([]byte("after 1 millisecond\n"))
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				br := bufio.NewReader(client)
				r, _, err := br.ReadRune()
				assert.Equal(t, 'a', r)
				assert.Nil(t, err)
			},
		}, {desc: "straight - json.Decoder.Decode will await",
			serverFunc: func(t *testing.T, server net.Conn) {
				time.Sleep(time.Millisecond)
				server.Write([]byte(`{"after": "1 millisecond"}`))
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				v := make(map[string]string)
				err := json.NewDecoder(client).Decode(&v)
				assert.Nil(t, err)
				assert.Equal(t, "1 millisecond", v["after"])
			},
		}, {desc: "separated - net.Conn.Read will await",
			serverFunc: func(t *testing.T, server net.Conn) {
				time.Sleep(time.Millisecond)
				server.Write([]byte("after 1 millisecond\n"))
				time.Sleep(time.Millisecond)
				server.Write([]byte("after 2 millisecond\n"))
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				buf := make([]byte, 40)
				n, err := client.Read(buf) //io.ReadFull(client, buf)
				assert.Equal(t, 20, n)
				assert.Nil(t, err)
				n, err = client.Read(buf[n:]) //io.ReadFull(client, buf)
				assert.Equal(t, 20, n)
				assert.Nil(t, err)
				assert.Equal(t, []byte("after 1 millisecond\nafter 2 millisecond\n"), buf)
			},
		}, {desc: "separated - json.Decoder.Decode will await",
			serverFunc: func(t *testing.T, server net.Conn) {
				time.Sleep(time.Millisecond)
				server.Write([]byte(`{"after": "1 millisecond",`))
				time.Sleep(time.Millisecond * 20)
				server.Write([]byte(`"result": "separated"}`))
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				//client.SetReadDeadline(time.Now().Add(time.Millisecond * 5))
				v := make(map[string]string)
				err := json.NewDecoder(client).Decode(&v)
				assert.Nil(t, err)
				assert.Equal(t, "1 millisecond", v["after"])
				assert.Equal(t, "separated", v["result"])
			},
		}, {desc: "unfinished - json.Decoder.Decode will return io.ErrUnexpectedEOF",
			serverFunc: func(t *testing.T, server net.Conn) {
				//assert.Nil(t, server.SetDeadline(time.Now().Add(time.Millisecond*500)))
				server.Write([]byte(`{"after": "1 millisecond", `))
				time.Sleep(time.Millisecond * 200)
			},
			clientFunc: func(t *testing.T, client net.Conn) {
				v := make(map[string]string)
				err := json.NewDecoder(client).Decode(&v)
				assert.Equal(t, io.ErrUnexpectedEOF, err)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			client := connect(t, testcase.serverFunc)
			defer client.Close()
			testcase.clientFunc(t, client)
		})
	}
}
