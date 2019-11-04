package jrpc

import (
	"context"
	"io"
)

// ServeStream is
func ServeStream(ctx context.Context, stream io.ReadWriter, repository *Core) error {
	dec := NewDecoder(stream)
	enc := NewEncoder(stream)
	requests := make([]*Request, 0, 10)

	var batch bool
	var err error
	var resps []*Response

	for {
		requests, batch, err = dec.Decode(Calibrate(requests, 10))
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		resps, err = repository.Execute(ctx, requests, batch)
		if err != nil {
			return err
		}

		if len(resps) == 0 {
			_, err = stream.Write([]byte{'\n'})
		} else {
			err = enc.Encode(resps, batch)
		}
		if err != nil {
			return err
		}
	}
}
