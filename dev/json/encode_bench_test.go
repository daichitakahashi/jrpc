package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/francoispqt/gojay"
	"github.com/intel-go/fastjson"
)

func BenchmarkJSONEncode(b *testing.B) {

	fd, _ := os.Open("sample.pdf")
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('"')
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	io.Copy(enc, fd)
	fd.Close()
	buf.WriteByte('"')

	//params := []byte(`"012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"`)
	params := buf.Bytes()
	id := []byte(`"noid"`)

	//buf.Reset()

	var rawParams json.RawMessage = params
	var rawID json.RawMessage = id

	b.Run("std", func(b *testing.B) {
		stdReq := &stdRequest{
			Version: "2.0",
			Method:  "add",
			Params:  &rawParams,
			ID:      &rawID,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(stdReq)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("fastjson", func(b *testing.B) {
		stdReq := &stdRequest{
			Version: "2.0",
			Method:  "add",
			Params:  &rawParams,
			ID:      &rawID,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := fastjson.Marshal(stdReq)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("buf", func(b *testing.B) {
		bufReq := &bufRequest{
			Version: "2.0",
			Method:  "add",
			Params:  &rawParams,
			ID:      &rawID,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// _, err := json.Marshal(bufReq)
			/*_, err := bufReq.MarshalJSON()
			if err != nil {
				panic(err)
			}*/
			b := bufReq.encode()
			b.Free()
		}
	})

	var embededParams gojay.EmbeddedJSON = params
	var embededID gojay.EmbeddedJSON = id

	b.Run("gojay", func(b *testing.B) {
		gojayReq := &gojayRequest{
			Version: "2.0",
			Method:  "add",
			Params:  &embededParams,
			ID:      &embededID,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := gojay.Marshal(gojayReq)
			if err != nil {
				panic(err)
			}
		}
	})
}
