package jrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	type example struct {
		ID ID `json:"id"`
	}

	t.Run("NewID", func(t *testing.T) {
		type TestCase struct {
			value   interface{}
			idValue interface{}
		}

		testcases := []TestCase{
			{
				value:   nil,
				idValue: nil,
			}, {
				value:   0,
				idValue: int64(0),
			}, {
				value:   int8(1),
				idValue: int64(1),
			}, {
				value:   int16(2),
				idValue: int64(2),
			}, {
				value:   int32(3),
				idValue: int64(3),
			}, {
				value:   int64(4),
				idValue: int64(4),
			}, {
				value:   uint(5),
				idValue: int64(5),
			}, {
				value:   uint8(6),
				idValue: int64(6),
			}, {
				value:   uint16(7),
				idValue: int64(7),
			}, {
				value:   uint32(8),
				idValue: int64(8),
			}, {
				value:   uint64(9),
				idValue: int64(9),
			}, {
				value:   float32(10.00001),
				idValue: float64(10.00001),
			}, {
				value:   float64(10.00002),
				idValue: float64(10.00002),
			}, {
				value:   "11",
				idValue: "11",
			},
		}
		for _, testcase := range testcases {
			id := NewID(testcase.value)
			assert.Equal(t, testcase.idValue, id.val)
		}

		assert.Panics(t, func() {
			NewID(true)
		}, "panic occurs when creating invalid id")
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		type TestCase struct {
			ID           ID
			expectedJSON string
			desc         string
		}
		testcases := []TestCase{
			{
				ID:           NewID(1),
				expectedJSON: `{"id":1}`,
				desc:         "integer ID",
			}, {
				ID:           NewID("2"),
				expectedJSON: `{"id":"2"}`,
				desc:         "string ID",
			}, {
				ID:           NewID(12.3456),
				expectedJSON: `{"id":12.3456}`,
				desc:         "float ID",
			}, {
				ID:           UnknownID,
				expectedJSON: `{"id":null}`,
				desc:         "explicit null ID",
			}, {
				ID:           NoID,
				expectedJSON: `{"id":null}`,
				desc:         "no ID",
			},
		}
		for _, testcase := range testcases {
			t.Run(testcase.desc, func(t *testing.T) {
				ex := example{
					ID: testcase.ID,
				}
				b, err := json.Marshal(ex)
				assert.Nil(t, err)
				assert.Equal(t, testcase.expectedJSON, string(b))
			})
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		type TestCase struct {
			src           []byte
			expectedID    ID
			expectedValue interface{}
			desc          string
		}
		testcases := []TestCase{
			{
				src:           []byte(`{"id":1}`),
				expectedID:    NewID(1),
				expectedValue: int64(1),
				desc:          "integer ID",
			}, {
				src:           []byte(`{"id":"2"}`),
				expectedID:    NewID("2"),
				expectedValue: "2",
				desc:          "string ID",
			}, {
				src:           []byte(`{"id":12.3456}`),
				expectedID:    NewID(12.3456),
				expectedValue: 12.3456,
				desc:          "float ID",
			}, {
				src:           []byte(`{"id":null}`),
				expectedID:    UnknownID,
				expectedValue: unknown{},
				desc:          "explicit null ID",
			}, {
				src:           []byte(`{}`),
				expectedID:    NoID,
				expectedValue: nil,
				desc:          "omitted ID",
			},
		}
		for _, testcase := range testcases {
			t.Run(testcase.desc, func(t *testing.T) {
				var ex example
				assert.Nil(t, json.Unmarshal(testcase.src, &ex))
				assert.Equal(t, testcase.expectedID, ex.ID)
				assert.True(t, ex.ID.EqualsValue(testcase.expectedValue))
			})
		}
	})
	/*
		t.Run("request id marshal", func(t *testing.T) {
			type testcase struct {
				ID           ID
				expectedJSON string
				desc         string
			}
			cases := []testcase{
				{
					ID:           NewID(1),
					expectedJSON: `{"jsonrpc":"","method":"","id":1}`,
					desc:         "integer ID",
				}, {
					ID:           NewID("2"),
					expectedJSON: `{"jsonrpc":"","method":"","id":"2"}`,
					desc:         "string ID",
				}, {
					ID:           NewID(12.3456),
					expectedJSON: `{"jsonrpc":"","method":"","id":12.3456}`,
					desc:         "float ID",
				}, {
					ID:           UnknownID,
					expectedJSON: `{"jsonrpc":"","method":"","id":null}`,
					desc:         "explicit null ID",
				}, {
					ID:           NoID,
					expectedJSON: `{"jsonrpc":"","method":""}`,
					desc:         "omitted ID",
				},
			}
			for _, c := range cases {
				t.Run(c.desc, func(t *testing.T) {
					r := &Request{
						ID: c.ID,
					}
					b, err := json.Marshal(r)
					assert.Nil(t, err)
					assert.Equal(t, c.expectedJSON, string(b))
				})
			}
		})

		t.Run("request id unmarshal", func(t *testing.T) {
			type testcase struct {
				src             []byte
				expectedRequest Request
				desc            string
			}
			cases := []testcase{
				{
					src:             []byte(`{"jsonrpc":"","method":"","id":1}`),
					expectedRequest: Request{ID: NewID(1)},
					desc:            "integer ID",
				}, {
					src:             []byte(`{"jsonrpc":"","method":"","id":"2"}`),
					expectedRequest: Request{ID: NewID("2")},
					desc:            "string ID",
				}, {
					src:             []byte(`{"jsonrpc":"","method":"","id":12.3456}`),
					expectedRequest: Request{ID: NewID(12.3456)},
					desc:            "float ID",
				}, {
					src:             []byte(`{"jsonrpc":"","method":"","id":null}`),
					expectedRequest: Request{ID: UnknownID},
					desc:            "explicit null ID",
				}, {
					src:             []byte(`{"jsonrpc":"","method":""}`),
					expectedRequest: Request{ID: NoID},
					desc:            "omitted ID",
				},
			}
			for _, c := range cases {
				t.Run(c.desc, func(t *testing.T) {
					var r *Request
					assert.Nil(t, json.Unmarshal(c.src, &r))
					assert.Equal(t, c.expectedRequest, *r)
				})
			}
		})

		t.Run("response id marshal", func(t *testing.T) {
			type testcase struct {
				ID           ID
				expectedJSON string
				desc         string
			}
			cases := []testcase{
				{
					ID:           NewID(1),
					expectedJSON: `{"jsonrpc":"","result":null,"id":1}`,
					desc:         "integer ID",
				}, {
					ID:           NewID("2"),
					expectedJSON: `{"jsonrpc":"","result":null,"id":"2"}`,
					desc:         "string ID",
				}, {
					ID:           NewID(12.3456),
					expectedJSON: `{"jsonrpc":"","result":null,"id":12.3456}`,
					desc:         "float ID",
				}, {
					ID:           UnknownID,
					expectedJSON: `{"jsonrpc":"","result":null,"id":null}`,
					desc:         "explicit null ID",
				}, {
					ID:           NoID,
					expectedJSON: `{"jsonrpc":"","result":null,"id":null}`,
					desc:         "omitted ID",
				},
			}
			for _, c := range cases {
				t.Run(c.desc, func(t *testing.T) {
					r := &Response{
						ID: c.ID,
					}
					b, err := json.Marshal(r)
					assert.Nil(t, err)
					assert.Equal(t, c.expectedJSON, string(b))
				})
			}
		})

		t.Run("response id unmarshal", func(t *testing.T) {
			type testcase struct {
				src              []byte
				expectedResponse Response
				desc             string
			}
			cases := []testcase{
				{
					src:              []byte(`{"jsonrpc":"","result":null,"id":1}`),
					expectedResponse: Response{ID: NewID(1)},
					desc:             "integer ID",
				}, {
					src:              []byte(`{"jsonrpc":"","result":null,"id":"2"}`),
					expectedResponse: Response{ID: NewID("2")},
					desc:             "string ID",
				}, {
					src:              []byte(`{"jsonrpc":"","method":"","id":12.3456}`),
					expectedResponse: Response{ID: NewID(12.3456)},
					desc:             "float ID",
				}, {
					src:              []byte(`{"jsonrpc":"","method":"","id":null}`),
					expectedResponse: Response{ID: UnknownID},
					desc:             "explicit null ID",
				}, {
					src:              []byte(`{"jsonrpc":"","method":""}`),
					expectedResponse: Response{ID: NoID},
					desc:             "omitted ID",
				},
			}
			for _, c := range cases {
				t.Run(c.desc, func(t *testing.T) {
					var r *Response
					assert.Nil(t, json.Unmarshal(c.src, &r))
					assert.Equal(t, c.expectedResponse, *r)
				})
			}
		})
	*/
	/*
		t.Run("invalid ids", func(t *tesing.T) {

		})
	*/
}
