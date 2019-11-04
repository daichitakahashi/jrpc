package jrpc

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
)

/*
TODO:

*/

// Decoder is
type Decoder struct {
	m     sync.Mutex
	err   error
	dirty bool
	//buf   []byte

	r   *bufio.Reader
	dec *json.Decoder
}

// NewDecoder is
func NewDecoder(r io.Reader) *Decoder {
	br := bufio.NewReader(r)
	return &Decoder{
		//buf: make([]byte, peekLength),
		r:   br,
		dec: json.NewDecoder(br),
	}
}

// Decode is
func (d *Decoder) Decode(dst []*Request) (requests []*Request, batch bool, err error) {
	return d.DecodeContext(nil, dst)
}

// DecodeContext is
func (d *Decoder) DecodeContext(ctx context.Context, dst []*Request) ([]*Request, bool, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if d.err != nil {
		return dst[:0], false, d.err
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return dst[:0], false, ctx.Err()
		default:
			var batch bool
			var err error
			done := make(chan struct{}, 1)
			go func() {
				dst, batch, err = d.decode(dst, done)
			}()

			select {
			case <-ctx.Done():
				return dst[:0], false, ctx.Err() // do not store context error
			case <-done:
				return dst, batch, err
			}
		}
	}
	return d.decode(dst, nil)

}

func (d *Decoder) decode(dst []*Request, done chan struct{}) ([]*Request, bool, error) {
	if done != nil {
		defer func() { close(done) }()
	}

	var char byte
	char, err := d.firstByte()
	if err != nil {
		d.err = err
		return dst[:0], false, err
	}
	batch := char == '['

	var req *Request
	if batch {
		d.dec.Token()
		for d.dec.More() {
			req = &Request{}
			err = d.dec.Decode(req)
			if err != nil {
				dst, batch, err = d.handleError(err, dst, batch)
				if err != nil || d.err != nil {
					return dst, batch, err
				}
				// when error is json.UnmarshalTypeError, decode is continuable
				continue
			}
			dst = append(dst, req)
		}

		_, err = d.dec.Token()
		if err != nil {
			return d.handleError(err, dst, batch)
		}
	} else {
		req = &Request{}
		err = d.dec.Decode(req)
		if err != nil {
			return d.handleError(err, dst, batch)
		}
		dst = append(dst, req)
	}
	return dst, batch, err
}

func (d *Decoder) handleError(err error, dst []*Request, batch bool) ([]*Request, bool, error) {
	switch err.(type) {
	case *json.SyntaxError:
		dst = append(dst[:0], &Request{
			Version: "2.0",
			Method:  rpcParseError,
			err:     err,
		})
		d.err = err
		return dst, false, nil
	case *json.UnmarshalTypeError:
		dst = append(dst, &Request{
			Version: "2.0",
			Method:  rpcInvalidRequest,
			err:     err,
		})
		return dst, batch, nil
	default:
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			// probably, connection is closed(in http, body reaches EOF)
			dst = append(dst[:0], &Request{
				Version: "2.0",
				Method:  rpcParseError,
				err:     err,
			})
			d.err = err
			return dst, false, nil
		}
		// or other stream error
		d.err = err
		return dst[:0], false, err
	}
}

// Calibrate is utility
func Calibrate(requests []*Request, capacity int) []*Request {
	for i := range requests {
		requests[i] = nil
	}
	if cap(requests) > capacity {
		return requests[:0:capacity]
	}
	return requests[:0]
}

//const peekLength = 32

func (d *Decoder) firstByte() (byte, error) {
	var b byte
	var err error
	if d.dirty {
		buffered := d.dec.Buffered()
		br := buffered.(*bytes.Reader)

		for {
			b, err = br.ReadByte()
			if err == io.EOF {
				d.dirty = false
				break
			}
			if !isSpace(b) {
				return b, nil
			}
		}
		/*
			for {
				n, err := br.Read(d.buf)
				if err == io.EOF { // the only possible error
					d.dirty = false
					break
				}
				for _, b := range d.buf[:n] {
					if !isSpace(b) {
						return b, nil
					}
				}
			}
		*/
	}
	if !d.dirty {
		d.dirty = true
		for {
			b, err = d.r.ReadByte()
			if err != nil {
				return 0, err
			}
			if !isSpace(b) {
				d.r.UnreadByte()
				return b, nil
			}
		}
		/*
			for {
				peeked, err := d.r.Peek(peekLength)
				if err != nil {
					if len(peeked) == 0 {
						return 0, err
					}
				}
				for _, b := range peeked {
					if !isSpace(b) {
						return b, nil
					}
				}
				d.r.Read(d.buf) // discards
			}
		*/
	}
	return 0, nil // should be unreachable
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// Reset not implemented yet
func (d *Decoder) Reset(r io.Reader) {
	d.err = nil
	d.dirty = false
	//d.r.Reset(r)
	d.r = bufio.NewReader(r)
	d.dec = json.NewDecoder(d.r)
}

// Err is
func (d *Decoder) Err() error {
	return d.err
}

/*
for {
	if dec.Err() != nil {
		dec.Reset(r)
	}

	batch, err := dec.Decode(&requests)
	if err != nil {
		switch et := err.(type) {
		case net.OpError:
			if et.Timeout() {
			} else if et.Temporary() {
				// ?????
			}
		default:
			return errors.Wrap(et, "jrpc: error:")
		}
	}

	repository.Execute(ctx,)

	enc := jrpc.NewEncoder(w)
	n, err := enc.Encode(requests, batch)

	// できるだけ、JSON-RPCの仕様で対応できない範囲のエラーのみを返すようにしたい。
}

json.Decoder.Decode内でコールされる、readValue関数
- for文の中で以下の処理
	1. 読み込み済みのデータの解析
	2. 解析にエラーがあれば返る
	3. 前回のデータ読み込みにエラーがあれば返る
		3-1. io.EOFの場合→オブジェクト一つを読み込み済みであればそのまま返る。そうでなく途中であれば、io.ErrUnexpectedEOFを返す
		3−2. その他のエラーの場合→そのままエラーを返す
	4. 次のためにデータを読み込み＆エラーを保存しておく
- であるから、コネクションがクローズした場合には、io.UnexpectedEOFが返る
	- この先でエラーレスポンスを用意しておいても意味がない可能性が高い→はたして用意しておくことに価値があるのかどうか→ない
- タイムアウトならタイムアウトエラーが返る（ex: net.Conn.SetReadDeadline）
	- タイムアウトエラーとコネクションのクローズは基本的には関係がないので、サーバ側で必要なら再度待機させることができる
	- おそらくサーバのプログラマは、エラーがio.ErrUnexpectedEOFであるかどうかで、コネクションのクローズを知るだろう
	- net.Errorインターフェースの問題は、コンテクストのタイムアウトも一括で扱えてしまうこと。
		- contextを取り扱うのをやめる（contextを持ち運ぶのは、それを伝播させる義務があるときor自身が末端であるとき、あるいは何らかのセッションを開始するとき）
		- なぜならコネクション(net.Conn)自体、すでにcontextを持ち運んでいる
		- →net.DialContext を参照
- ということは、尻切れとんぼのJSONについては、シンタックスエラーは出てこず、タイムアウトするのを待つだけ
- SyntaxErrorは、JSONオブジェクト中の誤字に応じて現れると考えれば良い
	- そうなったときに、どのようにエラーの起こった箇所をスキップするか
		- バウンダリ文字列を用いるように構築し直す
		- コネクションを終了させる（他のエラーと同じ扱いにする）
	- そもそも、JSON-RPCは処理が完了したものから順にレスポンスする（送信の順番は保証しない）ため、レスポンスを待たずに次を送ることは推奨されない
		- やろうと思えば、レスポンスのIDを見て振り分けることは可能
		- あくまでバッチ処理ではダメで、個別に送る必要があること、というのはないのでは？

*/
