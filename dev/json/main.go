package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/francoispqt/gojay"
)

/*
TODO:
- receivedRequest に、UnmarshalJSONを実装、その内部で単項とバッチを振り分ける、のが可能かどうか確認
	- パフォーマンス比較
- EmbeddedJSONについて、ポインタにしてアロケーションしておくのが良いのか、あるいは値としてゼロ値をあてにしていいのか、パフォーマンス確認
	- ポインタ
	- ポインタ+アロケーション
	- 値
	- 値+アロケーション
- 

*/

func main() {
	basic()
	fmt.Println("*******************************************")
	gojayDecoderBuffer()
}

func basic() {
	params := []byte(`[1,2]`)
	id := []byte(`"noid"`)

	var rawParams json.RawMessage = params
	var rawID json.RawMessage = id

	stdReq := &stdRequest{
		Version: "2.0",
		Method:  "add",
		Params:  &rawParams,
		ID:      &rawID,
	}

	result, err := json.Marshal(stdReq)
	if err != nil {
		panic(err)
	}
	fmt.Println("standard package encoding")
	fmt.Println(string(result))

	bufReq := &bufRequest{
		Version: "2.0",
		Method:  "add",
		Params:  &rawParams,
		ID:      &rawID,
	}

	result, err = json.Marshal(bufReq)
	if err != nil {
		panic(err)
	}
	fmt.Println("manual buffer encode")
	fmt.Println(string(result))

	var embededParams gojay.EmbeddedJSON = params
	var embededID gojay.EmbeddedJSON = id

	gojayReq := &gojayRequest{
		Version: "2.0",
		Method:  "add",
		Params:  &embededParams,
		ID:      &embededID,
	}

	result, err = gojay.Marshal(gojayReq)
	if err != nil {
		panic(err)
	}
	fmt.Println("gojay encode")
	fmt.Println(string(result))
}

func gojayDecoderBuffer() {
	src := strings.NewReader(`{"jsonrpc":"2.0","method":"add","params":[1,2],"id":"noid"}{"jsonrpc":"2.0","method":"add","params":[1,2],"id":"noid"}`)
	dec := gojay.NewDecoder(src)

	req := gojayRequest{
		Params: &gojay.EmbeddedJSON{},
		ID:     &gojay.EmbeddedJSON{},
	}
	err := dec.Decode(&req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("decoded")
	fmt.Println(req)

	buf := bufferpool.GetForBufferSlice().Bytes()
	_, err = src.Read(buf)
	if err != io.EOF {
		panic(err)
	}
}
