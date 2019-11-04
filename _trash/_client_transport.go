package jrpc

/*
TODO:

*/

/*
type (
	// ClientTransport is
	ClientTransport interface {
		Transport(ctx context.Context, ger GetterEncodedRequest, responseExpected bool) error
	}

	// ClientTransport2 is
	ClientTransport2 interface {
		SendRequest(ctx context.Context, r io.Reader) error // , responseExpected bool) error
		ReceivedResponse(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error)
		Close() error
	}

	// GetterEncodedRequest is
	GetterEncodedRequest interface {
		GetEncodedRequest(context.Context, io.Writer) (SetterReceivedResponse, error)
	}

	// SetterReceivedResponse is
	SetterReceivedResponse interface {
		SetReceivedResponse(context.Context, io.Reader) error
		NoResponseReceived() error
	}

	transportPipe struct {
		encodedReader  *io.PipeReader
		decodingWriter *io.PipeWriter
		responded      chan bool
		once           sync.Once
	}
)

func newTransportPipe() (tp *transportPipe, encodedWriter *io.PipeWriter, decodingReader *io.PipeReader) {
	tp = &transportPipe{
		responded: make(chan bool, 2),
	}
	tp.encodedReader, encodedWriter = io.Pipe()
	decodingReader, tp.decodingWriter = io.Pipe()
	return
}

func (tp *transportPipe) GetEncodedRequest(ctx context.Context, w io.Writer) (SetterReceivedResponse, error) {
	if w == nil {
		return nil, errors.New("jrpc: io.Writer must not be nil")
	}

	buf := bufferpool.GetForBufferSlice()
	defer buf.Free()

	err := ctxCopy(ctx, w, tp.encodedReader, buf.Bytes())
	return tp, err
}

func (tp *transportPipe) SetReceivedResponse(ctx context.Context, r io.Reader) (err error) {
	if r == nil {
		err = errors.New("jrpc: io.Reader must not be nil")
		return
	}
	tp.once.Do(func() {
		buf := bufferpool.GetForBufferSlice()
		defer buf.Free()

		tp.responded <- true
		tp.responded <- true

		err = ctxCopy(ctx, tp.decodingWriter, r, buf.Bytes())
		tp.decodingWriter.Close()
		if err == errSuccessful { // ensure decoding finished
			err = nil
		}
	})
	return
}

func (tp *transportPipe) NoResponseReceived() error {
	tp.once.Do(func() {
		close(tp.responded)
		tp.decodingWriter.Close()
	})
	return nil
}

var _ GetterEncodedRequest = (*transportPipe)(nil)
var _ SetterReceivedResponse = (*transportPipe)(nil)

*/
