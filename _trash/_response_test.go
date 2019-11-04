package jrpc

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeResponse(t *testing.T) {
	resp := makeResponse(nil)
	assert.Equal(t, &Response{
		Version: "2.0",
	}, resp)

	id := json.RawMessage([]byte("id1"))
	req := &Request{
		Version: "2.0",
		ID:      &id,
	}
	resp = makeResponse(req)
	assert.Equal(t, req.Version, resp.Version)
	assert.Equal(t, req.RawID, resp.RawID)
}

func TestresponseError(t *testing.T) {
	type testcase struct {
		original         error
		expectedResponse *Response
		expectedError    error
		desc             string
	}

	testCases := []testcase{
		{
			original:         nil,
			expectedResponse: nil,
			expectedError:    nil,
			desc:             "no error",
		}, {
			original: bufio.ErrInvalidUnreadRune,
			expectedResponse: &Response{
				Version: "2.0",
				Error:   ErrInternal(),
			},
			expectedError: bufio.ErrInvalidUnreadRune,
			desc:          "Unread error means internal operation error, response Internal Error, return original error(matter of server side)",
		}, {
			original:         io.ErrClosedPipe,
			expectedResponse: nil,
			expectedError:    io.ErrClosedPipe,
			desc:             "when reuse closed pipe, response nothing, return original error(matter of server side)",
		}, {
			original:         io.ErrNoProgress,
			expectedResponse: nil,
			expectedError:    io.ErrNoProgress,
			desc:             "when read error occurred, response nothing, return original error(server recognize connection error)",
		}, {
			original:         io.EOF,
			expectedResponse: nil,
			expectedError:    io.EOF,
			desc:             "when reach EOF(connection closed), response nothing, return original error(server recognize connection close)",
		}, {
			original: io.ErrUnexpectedEOF,
			expectedResponse: &Response{
				Version: "2.0",
				Error:   ErrParse(),
			},
			expectedError: nil,
			desc:          "when received json remains incompleted, try to response Parse Error, return nil(matter of client side)",
		}, {
			original: &json.SyntaxError{},
			expectedResponse: &Response{
				Version: "2.0",
				Error:   ErrParse(),
			},
			expectedError: nil,
			desc:          "when received json is invalid, response Parse Error, return nil(matter of client side)",
		}, {
			original: &json.UnmarshalTypeError{},
			expectedResponse: &Response{
				Version: "2.0",
				Error:   ErrInvalidRequest(),
			},
			expectedError: nil,
			desc:          "when type of received json's member is invalid (ex. \"jsonrpc\":2.0), response Parse Error, return nil(matter of client side)",
		}, {
			original:         errors.New("unknown error"),
			expectedResponse: nil, /*&Response{
				Version: "2.0",
				Error:   ErrInternal(),
			},*/
			expectedError: errors.New("unknown error"),
			desc:          "when unknown error occurred, response nothing, return original error(maybe connection error, for example net.OpError)",
		},
	}

	for _, c := range testCases {
		respErr, err := responseError(c.original)
		assert.Equal(t, c.expectedResponse, respErr, c.desc)
		assert.Equal(t, c.expectedError, err, c.desc)
	}
}
