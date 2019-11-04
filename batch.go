package jrpc

/*
TODO:

*/

const defaultBatchCapacity = 5

// BatchRequest is
type BatchRequest []*Request

// Add appends
func (reqs BatchRequest) Add(req ...*Request) {
	reqs = append(reqs, req...)
}

// MarshalJSON implements json.Marshaler
func (reqs BatchRequest) MarshalJSON() (b []byte, err error) {
	buf, err := reqs.encode()
	if err != nil {
		return
	}
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (reqs BatchRequest) encode() (*Buffer, error) {
	buf := bufferpool.Get()
	err := reqs.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (reqs BatchRequest) encodeTo(buf *Buffer) (err error) {
	lastIdx := len(reqs) - 1
	buf.AppendByte('[')
	for i := range reqs {
		err = reqs[i].encodeTo(buf)
		if err != nil {
			return
		}
		if i != lastIdx {
			buf.AppendByte(',')
		} else {
			buf.AppendByte(']')
		}
	}
	return nil
}

func (reqs BatchRequest) encodeLine() (*Buffer, error) {
	buf := bufferpool.Get()
	err := reqs.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	buf.AppendByte('\n')
	return buf, nil
}

var _ encoderLine = (BatchRequest)(nil)

// BatchResponse is
type BatchResponse []*Response

// GetFromID is
func (resps BatchResponse) GetFromID(id ID) (*Response, bool) {
	for _, r := range resps {
		if r.ID == id {
			return r, true
		}
	}
	return nil, false
}

// Map makes
// without UnknownID and NoID
func (resps BatchResponse) Map() map[ID]*Response {
	mapped := make(map[ID]*Response)
	for _, r := range resps {
		if r.ID == UnknownID || r.ID == NoID {
			continue
		}
		mapped[r.ID] = r
	}
	return mapped
}

/*
// MarshalJSON implements json.Marshaler
func (resps *BatchResponse) MarshalJSON() (b []byte, err error) {
	buf, err := resps.encode()
	if err != nil {
		return
	}
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}
*/

func (resps BatchResponse) encode() (*Buffer, error) {
	buf := bufferpool.Get()
	err := resps.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (resps BatchResponse) encodeTo(buf *Buffer) (err error) {
	lastIdx := len(resps) - 1
	buf.AppendByte('[')
	for i := range resps {
		err = resps[i].encodeTo(buf)
		if err != nil {
			return
		}
		if i != lastIdx {
			buf.AppendByte(',')
		} else {
			buf.AppendByte(']')
		}
	}
	return nil
}

func (resps BatchResponse) encodeLine() (*Buffer, error) {
	buf := bufferpool.Get()
	err := resps.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	buf.AppendByte('\n')
	return buf, nil
}

var _ encoderLine = (BatchResponse)(nil)
