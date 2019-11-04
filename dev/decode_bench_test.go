package dev

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/daichitakahashi/jrpc"
)

func makeJSONArrayRaw(n int) io.Reader {
	buf := bytes.NewBuffer(nil)

	buf.Write([]byte{'['})

	for i := 0; i < n; i++ {
		if i != 0 {
			buf.Write([]byte{','})
		}
		buf.Write([]byte(`{"jsonrpc":"2.0", "method": "eth_getStorageAt", "params": ["0x295a70b2de5e3953354a6a8344e616ed314d7251", "0x0", "latest"], "id": 1}`))
	}

	buf.Write([]byte{']'})

	return buf
}

func iferr(err error) {
	if err != nil {
		panic(err)
	}
}

func BenchmarkContinuousBuffer(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	var raws []*json.RawMessage

	iferr(json.NewDecoder(r).Decode(&raws))

	buf := bytes.NewBuffer(nil)
	dec := json.NewDecoder(buf)

	wg := &sync.WaitGroup{}
	rm := sync.Mutex{}

	for _, raw := range raws {
		wg.Add(1)
		rm.Lock()
		buf.Write(*raw)
		rm.Unlock()
		var dummy *jrpc.Request
		go func() {
			func() {
				rm.Lock()
				iferr(dec.Decode(&dummy))
				rm.Unlock()
			}()
			time.Sleep(time.Millisecond * 4)
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkResetReader(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	var raws []*json.RawMessage

	iferr(json.NewDecoder(r).Decode(&raws))

	br := bytes.NewReader(nil)
	dec := json.NewDecoder(br)

	wg := &sync.WaitGroup{}

	var dummy *jrpc.Request
	for _, raw := range raws {
		br.Reset(*raw)
		iferr(dec.Decode(&dummy))
		wg.Add(1)
		go func() {
			time.Sleep(time.Millisecond * 4)
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkDecodeToken(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	dec := json.NewDecoder(r)

	t, err := dec.Token()
	if t != json.Delim('[') || err != nil {
		panic(t)
	}

	var dummy *jrpc.Request
	array := make([]*jrpc.Request, 1000)

	for dec.More() {
		iferr(dec.Decode(&dummy))
		array = append(array, dummy)
	}
	t, err = dec.Token()
	if t != json.Delim(']') || err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < len(array); i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Millisecond * 4)
			wg.Done()
		}()
	}

	wg.Wait()
}

/*
func BenchmarkContinuousBuffer(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	var raws []*json.RawMessage

	iferr(json.NewDecoder(r).Decode(&raws))

	buf := bytes.NewBuffer(nil)
	dec := json.NewDecoder(buf)

	var dummy interface{}
	for _, raw := range raws {
		buf.Write(*raw)
		iferr(dec.Decode(&dummy))
	}
}

func BenchmarkContinuousBufferFastjson(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	var raws []*fastjson.RawMessage

	iferr(fastjson.NewDecoder(r).Decode(&raws))

	buf := bytes.NewBuffer(nil)
	dec := fastjson.NewDecoder(buf)

	var dummy interface{}
	for _, raw := range raws {
		buf.Write(*raw)
		iferr(dec.Decode(&dummy))
	}
}

func BenchmarkNoWrappedReaderFastjson(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	var raws []*fastjson.RawMessage

	iferr(fastjson.NewDecoder(r).Decode(&raws))

	br := bytes.NewReader(nil)
	dec := fastjson.NewDecoder(br)

	var dummy interface{}
	for _, raw := range raws {
		br.Reset(*raw)
		iferr(dec.Decode(&dummy))
	}
}

func BenchmarkDecodeTokenFastjson(b *testing.B) {
	r := makeJSONArrayRaw(b.N / 100)

	b.ResetTimer()

	dec := fastjson.NewDecoder(r)

	var dummy interface{}
	t, err := dec.Token()
	if t != fastjson.Delim('[') || err != nil {
		panic(t)
	}
	for dec.More() {
		iferr(dec.Decode(&dummy))
	}
	t, err = dec.Token()
	if t != fastjson.Delim(']') || err != nil {
		panic(err)
	}
}
*/

/*
func makeJSONArray(n int) []*RawMessage {
	var array []*RawMessage

	for i := 0; i < n; i++ {
		json := RawMessage(`{"jsonrpc":"2.0","data":"dummytext"}`)
		array = append(array, &json)
	}

	return array
}

func BenchmarkIndividualBuffer(b *testing.B) {
	array := makeJSONArray(b.N / 100)

	b.ResetTimer()

	var dummy interface{}
	for _, raw := range array {
		iferr(json.NewDecoder(bytes.NewReader(*raw)).Decode(&dummy))
	}
}

func BenchmarkResetBuffer(b *testing.B) {
	array := makeJSONArray(b.N / 100)

	b.ResetTimer()
	r := bytes.NewReader(nil)

	var dummy interface{}
	for _, raw := range array {
		r.Reset(*raw)

		iferr(json.NewDecoder(r).Decode(&dummy))
	}
}

type readWrapper struct {
	r io.Reader
}

func (rw *readWrapper) Read(b []byte) (n int, err error) {
	return rw.r.Read(b)
}

func (rw *readWrapper) Reset(r io.Reader) {
	rw.r = r
}

func BenchmarkWrappedReader(b *testing.B) {
	array := makeJSONArray(b.N / 100)

	b.ResetTimer()
	r := bytes.NewReader(nil)
	rw := &readWrapper{r: r}
	dec := json.NewDecoder(rw)

	var dummy interface{}
	for _, raw := range array {
		r.Reset(*raw)
		rw.Reset(r)

		iferr(dec.Decode(&dummy))
	}
}

func BenchmarkCompact(b *testing.B) {
	array := makeJSONArray(b.N / 100)

	b.ResetTimer()

	buf := bytes.NewBuffer(nil)
	dec := json.NewDecoder(buf)

	var dummy interface{}
	for _, raw := range array {
		json.Compact(buf, *raw)

		iferr(dec.Decode(&dummy))
	}
}
*/
