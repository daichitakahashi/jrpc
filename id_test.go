package jrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	type TestCase struct {
		src     interface{}
		_string string
		_type   string
		_panic  bool
		desc    string
	}
	testcases := []TestCase{
		{
			src:     0,
			_string: "0",
			_type:   "number",
			_panic:  false,
			desc:    "int",
		}, {
			src:     int8(1),
			_string: "1",
			_type:   "number",
			_panic:  false,
			desc:    "int8",
		}, {
			src:     int16(2),
			_string: "2",
			_type:   "number",
			_panic:  false,
			desc:    "int16",
		}, {
			src:     int32(3),
			_string: "3",
			_type:   "number",
			_panic:  false,
			desc:    "int32",
		}, {
			src:     int64(-4),
			_string: "-4",
			_type:   "number",
			_panic:  false,
			desc:    "int64",
		}, {
			src:     uint(5),
			_string: "5",
			_type:   "number",
			_panic:  false,
			desc:    "uint",
		}, {
			src:     uint8(6),
			_string: "6",
			_type:   "number",
			_panic:  false,
			desc:    "uint8",
		}, {
			src:     uint16(7),
			_string: "7",
			_type:   "number",
			_panic:  false,
			desc:    "uint16",
		}, {
			src:     uint32(8),
			_string: "8",
			_type:   "number",
			_panic:  false,
			desc:    "uint32",
		}, {
			src:     uint64(9),
			_string: "9",
			_type:   "number",
			_panic:  false,
			desc:    "uint64",
		}, {
			src:     float32(1.234),
			_string: "1.234",
			_type:   "number",
			_panic:  false,
			desc:    "float32",
		}, {
			src:     float64(5.678),
			_string: "5.678",
			_type:   "number",
			_panic:  false,
			desc:    "float64",
		}, {
			src:     "test sentence",
			_string: "test sentence",
			_type:   "string",
			_panic:  false,
			desc:    "string",
		}, {
			src:     nil,
			_string: "null",
			_type:   "unknown",
			_panic:  false,
			desc:    "null",
		}, {
			src:     true,
			_string: "",
			_type:   "",
			_panic:  true,
			desc:    "invalid id",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			if testcase._panic {
				require.Panics(t, func() {
					NewID(testcase.src)
				})
				return
			}
			id := NewID(testcase.src)
			require.Equal(t, testcase._string, id.String())
			require.Equal(t, testcase._type, id.Type())
		})
	}
}

func TestID_Equals(t *testing.T) {
	t.Run("number", func(t *testing.T) {

		id := NewID(int32(12345))

		type TestCase struct {
			value  interface{}
			equals bool
			desc   string
		}
		testcases := []TestCase{
			{
				value:  12345,
				equals: true,
				desc:   "int",
			}, {
				value:  int8(123),
				equals: false,
				desc:   "int8",
			}, {
				value:  int16(12345),
				equals: true,
				desc:   "int16",
			}, {
				value:  int32(12345),
				equals: true,
				desc:   "int32",
			}, {
				value:  int64(12345),
				equals: true,
				desc:   "int64",
			}, {
				value:  uint(12345),
				equals: true,
				desc:   "uint",
			}, {
				value:  uint8(123),
				equals: false,
				desc:   "uint8",
			}, {
				value:  uint16(12345),
				equals: true,
				desc:   "uint16",
			}, {
				value:  uint32(12345),
				equals: true,
				desc:   "uint32",
			}, {
				value:  uint64(12345),
				equals: true,
				desc:   "uint64",
			}, {
				value:  float32(12345.0),
				equals: true,
				desc:   "float32",
			}, {
				value:  12345.0,
				equals: true,
				desc:   "float64",
			}, {
				value:  1.2345,
				equals: false,
				desc:   "float64-false",
			},
		}
		for _, testcase := range testcases {
			require.Equal(
				t,
				testcase.equals,
				id.Equals(testcase.value),
				testcase.desc,
			)
		}
	})

	t.Run("string", func(t *testing.T) {
		id := NewID("12345")
		require.True(t, id.Equals("12345"))
		require.False(t, id.Equals(12345))
		id = NewID(12345)
		require.False(t, id.Equals("12345"))
	})

	t.Run("unknown", func(t *testing.T) {
		id := UnknownID
		require.True(t, id.Equals(nil))
	})

	t.Run("invalid type", func(t *testing.T) {
		id := ID{
			n:     json.Number("12345"),
			nType: 16,
		}
		require.False(t, id.Equals(12345))
		require.False(t, id.Equals("12345"))

		id = ID{
			n:     json.Number("abcde"),
			nType: typeNumber,
		}
		require.False(t, id.Equals(1234))
	})
}

func TestID_MarshalJSON(t *testing.T) {
	type (
		example struct {
			ID ID `json:"id,omitempty"`
		}
		TestCase struct {
			ID           ID
			expectedJSON string
			err          bool
			desc         string
		}
	)
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
		}, {
			ID:   ID{nType: 8}, // invalid type
			err:  true,
			desc: "invalid id type",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			ex := example{
				ID: testcase.ID,
			}
			b, err := json.Marshal(ex)
			if !testcase.err {
				require.Nil(t, err)
				require.Equal(t, testcase.expectedJSON, string(b))
			} else {
				require.NotNil(t, err)
				require.Panics(t, func() {
					_ = testcase.ID.Type()
				})
			}
		})
	}
}

func TestID_UnmarshalJSON(t *testing.T) {

	type (
		example struct {
			ID ID `json:"id,omitempty"`
		}

		TestCase struct {
			src        []byte
			expectedID ID
			success    bool
			desc       string
		}
	)
	testcases := []TestCase{
		{
			src:        []byte(`{"id":1}`),
			expectedID: NewID(1),
			success:    true,
			desc:       "integer ID",
		}, {
			src:        []byte(`{"id":"2"}`),
			expectedID: NewID("2"),
			success:    true,
			desc:       "string ID",
		}, {
			src:        []byte(`{"id":12.3456}`),
			expectedID: NewID(12.3456),
			success:    true,
			desc:       "float ID",
		}, {
			src:        []byte(`{"id":null}`),
			expectedID: UnknownID,
			success:    true,
			desc:       "explicit null ID",
		}, {
			src:        []byte(`{}`),
			expectedID: NoID,
			success:    true,
			desc:       "omitted ID",
		}, {
			src:        []byte(`{"id":true}`),
			expectedID: NoID,
			success:    false,
			desc:       "invalid bool",
		}, {
			src:        []byte(`{"id":{"key":"value"}`),
			expectedID: NoID,
			success:    false,
			desc:       "invalid object",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			var ex example
			require.Equal(t, testcase.success, json.Unmarshal(testcase.src, &ex) == nil)
			require.Equal(t, testcase.expectedID, ex.ID)
		})
	}
}
