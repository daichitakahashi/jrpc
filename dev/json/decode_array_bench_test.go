package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/francoispqt/gojay"
)

/*
TODO:

*/

func BenchmarkJSONArrayDecode(b *testing.B) {
	fd, _ := os.Open("sample.pdf")
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('"')
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	io.Copy(enc, fd)
	fd.Close()
	buf.WriteByte('"')

	//encoded := `"012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"`
	encoded := buf.String()

	src := `
	[
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"},
		{"jsonrpc":"2.0","method":"jsonrpc.decode.bench","params":` + encoded + `,"id":"noid"}
	]`
	r := strings.NewReader(src)

	b.Run("std---------------------------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			br := bufio.NewReader(r)
			dec := json.NewDecoder(br)

			r, err := trimReader(br)
			if err != nil {
				panic(err)
			} else if r != '[' {
				panic("not array")
			}

			var reqs []stdRequest
			err = dec.Decode(&reqs)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("std+RawMessage----------------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			br := bufio.NewReader(r)
			dec := json.NewDecoder(br)

			r, err := trimReader(br)
			if err != nil {
				panic(err)
			} else if r != '[' {
				panic("not array")
			}

			var raws []*json.RawMessage
			err = dec.Decode(&raws)
			if err != nil {
				panic(err)
			}

			for _, raw := range raws {
				req := stdRequest{}
				err = json.Unmarshal(*raw, &req)
				if err != nil {
					panic(err)
				}
			}
		}
	})

	b.Run("std+Token---------------------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			br := bufio.NewReader(r)
			dec := json.NewDecoder(br)

			t, err := dec.Token()
			if err != nil {
				panic(err)
			} else if t != json.Delim('[') {
				panic(t)
			}

			for dec.More() {
				req := stdRequest{}
				err = dec.Decode(&req)
				if err != nil {
					panic(err)
				}
			}

			t, err = dec.Token()
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("gojay+UnmarshalArray----------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			br := bufio.NewReader(r)
			dec := gojay.BorrowDecoder(br)

			r, err := trimReader(br)
			if err != nil {
				panic(err)
			} else if r != '[' {
				panic("not array")
			}

			var reqs gojayBatchRequest
			err = dec.DecodeArray(&reqs)
			if err != nil {
				panic(err)
			}
			dec.Release()
		}
	})

	b.Run("gojay+Embedded----------------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			arr := []*gojayRequest{}
			r.Seek(0, 0)
			br := bufio.NewReader(r)
			dec := gojay.BorrowDecoder(br)

			r, err := trimReader(br)
			if err != nil {
				panic(err)
			} else if r != '[' {
				panic("not array")
			}

			var embs embeddedArray
			err = dec.Decode(&embs)
			if err != nil {
				panic(err)
			}
			dec.Release()

			for _, emb := range embs {
				gojayReq := &gojayRequest{
					Params: emb, // 689901680 B/op
					//Params: &gojay.EmbeddedJSON{}, // 689903256 B/op
					ID: &gojay.EmbeddedJSON{},
				}
				err = gojay.UnmarshalJSONObject(*emb, gojayReq)
				if err != nil {
					panic(err)
				}
				arr = append(arr, gojayReq)
			}
		}
	})

	b.Run("gojay+UnmarshalArray+redecode-", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := gojay.BorrowDecoder(r)

			var raw gojay.EmbeddedJSON
			err := dec.Decode(&raw)
			if err != nil {
				panic(err)
			} else if raw[0] != '[' {
				panic(raw[0])
			}
			dec.Release()

			var reqs gojayBatchRequest
			err = gojay.UnmarshalJSONArray(raw, &reqs)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("gojay+Embedded+redecode-------", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := gojay.BorrowDecoder(r)

			var raw gojay.EmbeddedJSON
			err := dec.Decode(&raw)
			if err != nil {
				panic(err)
			} else if raw[0] != '[' {
				panic(raw[0])
			}
			dec.Release()

			var embs embeddedArray
			err = gojay.UnmarshalJSONArray(raw, &embs)
			if err != nil {
				panic(err)
			}

			for _, emb := range embs {
				gojayReq := &gojayRequest{
					Params: emb,
					ID:     &gojay.EmbeddedJSON{},
				}
				err = gojay.UnmarshalJSONObject(*emb, gojayReq)
				if err != nil {
					panic(err)
				}
			}
		}
	})
}

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
