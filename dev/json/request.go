package main

import (
	"encoding/json"

	"github.com/daichitakahashi/jrpc"
	"github.com/francoispqt/gojay"
)

var bufferpool = jrpc.NewPool()

type stdRequest struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      *json.RawMessage `json:"id"`
}

type bufRequest struct {
	Version string
	Method  string
	Params  *json.RawMessage
	ID      *json.RawMessage
}

func (br *bufRequest) MarshalJSON() (b []byte, err error) {
	buf := bufferpool.Get()
	buf.AppendString(`{"jsonrpc":"`)
	buf.AppendString(br.Version)
	buf.AppendString(`","method":"`)
	buf.AppendString(br.Method)
	buf.AppendByte('"')
	if br.Params != nil {
		buf.AppendString(`,"params":`)
		buf.Write(*br.Params)
	}
	if br.ID != nil {
		buf.AppendString(`,"id":`)
		buf.Write(*br.ID)
	}
	buf.AppendByte('}')
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (br *bufRequest) encode() *jrpc.Buffer {
	buf := bufferpool.Get()
	buf.AppendString(`{"jsonrpc":"`)
	buf.AppendString(br.Version)
	buf.AppendString(`","method":"`)
	buf.AppendString(br.Method)
	buf.AppendByte('"')
	if br.Params != nil {
		buf.AppendString(`,"params":`)
		buf.Write(*br.Params)
	}
	if br.ID != nil {
		buf.AppendString(`,"id":`)
		buf.Write(*br.ID)
	}
	buf.AppendByte('}')
	return buf
}

type gojayRequest struct {
	Version string
	Method  string
	Params  *gojay.EmbeddedJSON
	ID      *gojay.EmbeddedJSON
}

func (gr *gojayRequest) MarshalJSONObject(enc *gojay.Encoder) {
	enc.StringKey("jsonrpc", gr.Version)
	enc.StringKey("method", gr.Method)
	enc.AddEmbeddedJSONKey("params", gr.Params)
	enc.AddEmbeddedJSONKeyOmitEmpty("id", gr.ID)
}
func (gr *gojayRequest) IsNil() bool {
	return gr == nil
}

var _ gojay.MarshalerJSONObject = (*gojayRequest)(nil)

func (gr *gojayRequest) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "jsonrpc":
		return dec.String(&gr.Version)
	case "method":
		return dec.String(&gr.Method)
	case "params":
		return dec.EmbeddedJSON(gr.Params)
	case "id":
		return dec.EmbeddedJSON(gr.ID)
	}
	return nil
}

func (gr *gojayRequest) NKeys() int {
	return 0
}

var _ gojay.UnmarshalerJSONObject = (*gojayRequest)(nil)

// 型をキャストするかどうかで、下記メソッドの使用・不使用を入れ替えることができる
type gojayBatchRequest []*gojayRequest

func (gbr gojayBatchRequest) UnmarshalJSONArray(dec *gojay.Decoder) error {
	gojayReq := &gojayRequest{
		Params: &gojay.EmbeddedJSON{},
		ID:     &gojay.EmbeddedJSON{},
	}
	err := dec.Object(gojayReq)
	if err != nil {
		return err
	}
	gbr = append(gbr, gojayReq)
	return nil
}

type embeddedArray []*gojay.EmbeddedJSON

func (ea embeddedArray) UnmarshalJSONArray(dec *gojay.Decoder) error {
	var emb gojay.EmbeddedJSON
	err := dec.EmbeddedJSON(&emb)
	if err != nil {
		return err
	}
	ea = append(ea, &emb)
	return nil
}
