package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/francoispqt/gojay"
)

/*
TODO:

*/

func BenchmarkJSONUnaryDecode(b *testing.B) {

	fd, _ := os.Open("sample.pdf")
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('"')
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	io.Copy(enc, fd)
	fd.Close()
	buf.WriteByte('"')

	//encoded := `"012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"`
	encoded := buf.String()
	//encodedLen := len(encoded)

	src := `
	{
		"jsonrpc":"2.0",
		"method":"jsonrpc.decode.bench",
		"params":` + encoded + `,
		"id":"noid"
	}`
	r := strings.NewReader(src)

	b.Run("std-----------", func(b *testing.B) {
		stdReq := &stdRequest{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0) // ==================================
			dec := json.NewDecoder(r)
			err := dec.Decode(stdReq)
			if err != nil {
				panic(err)
			}
		}
	})

	/*
		b.Run("fastjson------", func(b *testing.B) {
			stdReq := &stdRequest{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := strings.NewReader(src)
				dec := fastjson.NewDecoder(r)
				err := dec.Decode(stdReq)
				if err != nil {
					panic(err)
				}
			}
		})
	*/

	b.Run("std+token-----", func(b *testing.B) {
		stdReq := &stdRequest{}

		b.ResetTimer()
		//rawParams := make(json.RawMessage, encodedLen)
		for i := 0; i < b.N; i++ {
			r := strings.NewReader(src)
			dec := json.NewDecoder(r)

			t, err := dec.Token()
			if err != nil {
				panic(err)
			} else if t != json.Delim('{') {
				panic(t)
			}

			for dec.More() {
				t, err = dec.Token()
				if err != nil {
					panic(err)
				}
				switch t {
				case "jsonrpc":
					t, err = dec.Token()
					if err != nil {
						panic(err)
					}
					stdReq.Version = t.(string)
				case "method":
					t, err = dec.Token()
					if err != nil {
						panic(err)
					}
					stdReq.Method = t.(string)
				case "params":
					var rawParams json.RawMessage
					err = dec.Decode(&rawParams)
					if err != nil {
						panic(err)
					}
					stdReq.Params = &rawParams
				case "id":
					//stdReq.ID = &json.RawMessage{} // 48856452 B/op	      82 allocs/op
					var rawID json.RawMessage // 48856451 B/op	      82 allocs/op
					err = dec.Decode(&rawID)
					if err != nil {
						panic(err)
					}
					stdReq.ID = &rawID
				default:
					panic(t)
				}
			}

			t, err = dec.Token()
			if err != nil {
				panic(err)
			} else if t != json.Delim('}') {
				panic(t)
			}
		}
	})

	b.Run("std+raw-------", func(b *testing.B) {
		stdReq := &stdRequest{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := json.NewDecoder(r)
			var raw json.RawMessage
			err := dec.Decode(&raw)
			if err != nil {
				panic(err)
			}
			if raw[0] != '{' {
				panic(raw[0])
			}
			json.Unmarshal(raw, stdReq)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("std+custom----", func(b *testing.B) {
		//stdReq := &stdRequest{}
		reqs := &receivedRequest{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := json.NewDecoder(r)
			err := dec.Decode(&reqs)
			if err != nil || reqs.req == nil {
				panic(err)
			}
		}
	})

	b.Run("gojay---------", func(b *testing.B) {
		gojayReq := &gojayRequest{
			Params: &gojay.EmbeddedJSON{},
			ID:     &gojay.EmbeddedJSON{},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			// dec := gojay.BorrowDecoder(r)
			dec := gojay.NewDecoder(r)
			err := dec.Decode(gojayReq)
			if err != nil {
				panic(err)
			}
			// dec.Release()
		}
	})

	b.Run("gojay+embedded", func(b *testing.B) {
		gojayReq := &gojayRequest{
			//Params: &gojay.EmbeddedJSON{},
			ID: &gojay.EmbeddedJSON{},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			// dec := gojay.BorrowDecoder(r)
			dec := gojay.NewDecoder(r)
			var raw gojay.EmbeddedJSON //= make([]byte, 0, len(src))
			err := dec.Decode(&raw)
			if err != nil {
				panic(err)
			}
			if raw[0] != '{' {
				panic(raw[0])
			}
			gojayReq.Params = &raw // reuse (in unmarshal, data is to be copied)
			err = gojay.UnmarshalJSONObject(raw, gojayReq)
			if err != nil {
				panic(err)
			}
			// dec.Release()
		}
	})
}

type receivedRequest struct {
	req *stdRequest
}

func (rr *receivedRequest) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '{':
		rr.req = &stdRequest{}
		return json.Unmarshal(data, rr.req)
	case '[':
		// return json.Unmarshal(data, rr.batchResp)
		fallthrough
	default:
		return errors.New("invalid JSON-RPC Request object")
	}
}
