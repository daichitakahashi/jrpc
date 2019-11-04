package jrpc

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func BenchmarkFlush(b *testing.B) {
	makeEntry := func(s string) *Buffer {
		buf := bufferpool.Get()
		buf.AppendString(`{"json":`)
		for i := 0; i < 10; i++ {
			buf.AppendString(s)
		}
		buf.AppendByte('}')
		return buf
	}

	doWithoutFlush := func(dst io.Writer) {
		contents := []*Buffer{
			makeEntry("1"),
			makeEntry("2"),
			makeEntry("3"),
			makeEntry("4"),
			makeEntry("5"),
			makeEntry("6"),
			makeEntry("7"),
			makeEntry("8"),
			makeEntry("9"),
		}
		for i, buf := range contents {
			if i == 0 {
				dst.Write([]byte{'['})
			} else {
				dst.Write([]byte{','})
			}

			_, err := dst.Write(buf.Bytes())
			if err != nil {
				panic(err)
			}
			buf.Free()
		}
		dst.Write([]byte{']', '\n'})

	}

	flushBuffer := make([]byte, 1024*2)
	doFlush := func(dst io.Writer, buf []byte) {
		contents := []*Buffer{
			makeEntry("1"),
			makeEntry("2"),
			makeEntry("3"),
			makeEntry("4"),
			makeEntry("5"),
			makeEntry("6"),
			makeEntry("7"),
			makeEntry("8"),
			makeEntry("9"),
		}
		f := NewFlusherFromBuffers(dst, contents, true)
		f.SetBuffer(buf)
		//
		_, err := f.Flush(context.Background())
		if err != nil {
			panic(err)
		}
	}

	b.Run("discard_without_flush", func(b *testing.B) {
		b.ResetTimer()
		for c := 0; c < b.N; c++ {
			doWithoutFlush(ioutil.Discard)
		}
	})

	b.Run("discard_flush", func(b *testing.B) {
		b.ResetTimer()
		for c := 0; c < b.N; c++ {
			doFlush(ioutil.Discard, flushBuffer)
		}
	})

	b.Run("file_without_flush", func(b *testing.B) {
		fd, err := os.Create("xx_test_noflush.txt")
		defer fd.Close()
		if err != nil {
			panic(err)
		}
		b.ResetTimer()
		for c := 0; c < b.N; c++ {
			doWithoutFlush(fd)
		}
		os.Remove("xx_test_noflush.txt")
	})

	b.Run("file_flush", func(b *testing.B) {
		fd, err := os.Create("xx_test_flush.txt")
		defer fd.Close()
		if err != nil {
			panic(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doFlush(fd, flushBuffer)
		}
		os.Remove("xx_test_flush.txt")
	})

	serverBuffer := make([]byte, 1024)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()

	b.Run("conn_without_flush", func(b *testing.B) {
		done := make(chan struct{})
		go func() {
			server, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			defer server.Close()
			for {
				select {
				case <-done:
					return
				default:
					server.Read(serverBuffer)
				}
			}
		}()
		client, err := net.Dial("tcp", ln.Addr().String())
		defer client.Close()
		if err != nil {
			panic(err)
		}
		b.ResetTimer()
		for c := 0; c < b.N; c++ {
			doWithoutFlush(client)
		}
	})

	b.Run("conn_flush", func(b *testing.B) {
		done := make(chan struct{})
		go func() {
			server, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			defer server.Close()
			for {
				select {
				case <-done:
					return
				default:
					server.Read(serverBuffer)
				}
			}
		}()
		client, err := net.Dial("tcp", ln.Addr().String())
		defer client.Close()
		if err != nil {
			panic(err)
		}
		b.ResetTimer()
		for c := 0; c < b.N; c++ {
			doFlush(client, flushBuffer)
		}
	})

}
