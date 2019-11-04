package jrpc

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeValue(t *testing.T) {
	type TestCase struct {
		value          interface{}
		expectedResult string
		desc           string
	}
	testcases := []TestCase{
		{
			value:          nil,
			expectedResult: "null",
			desc:           "nil",
		}, {
			value:          -1,
			expectedResult: "-1",
			desc:           "int",
		}, {
			value:          int8(-2),
			expectedResult: "-2",
			desc:           "int8",
		}, {
			value:          int16(-3),
			expectedResult: "-3",
			desc:           "int16",
		}, {
			value:          int32(-4),
			expectedResult: "-4",
			desc:           "int32",
		}, {
			value:          int64(-5),
			expectedResult: "-5",
			desc:           "int64",
		}, {
			value:          uint(1),
			expectedResult: "1",
			desc:           "uint",
		}, {
			value:          uint8(2),
			expectedResult: "2",
			desc:           "uint8",
		}, {
			value:          uint16(3),
			expectedResult: "3",
			desc:           "uint16",
		}, {
			value:          uint32(4),
			expectedResult: "4",
			desc:           "uint32",
		}, {
			value:          uint64(5),
			expectedResult: "5",
			desc:           "uint64",
		}, {
			value:          float32(1.2345),
			expectedResult: "1.2345",
			desc:           "float32",
		}, {
			value:          1.23456789,
			expectedResult: "1.23456789",
			desc:           "float64",
		}, {
			value:          true,
			expectedResult: "true",
			desc:           "bool",
		}, {
			value:          "\"example\"",
			expectedResult: `"\"example\""`,
			desc:           "string",
		}, {
			value:          sampleMarshaler{},
			expectedResult: "false",
			desc:           "json.Marshaler",
		}, {
			value:          []byte("hogehoge"),
			expectedResult: `"` + base64.StdEncoding.EncodeToString([]byte("hogehoge")) + `"`,
			desc:           "[]byte",
		}, {
			value:          []int{0, 1, 2, 3, 4},
			expectedResult: "[0,1,2,3,4]",
			desc:           "[]int",
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			data, err := encodeValue(testcase.value)
			require.NoError(t, err)
			require.Equal(t, testcase.expectedResult, string(data))
		})
	}
}

func BenchmarkEncodeValue(b *testing.B) {
	type TestCase struct {
		value          interface{}
		expectedResult string
		desc           string
	}
	testcases := []TestCase{
		{
			value:          nil,
			expectedResult: "null",
			desc:           "nil",
		}, {
			value:          -1,
			expectedResult: "-1",
			desc:           "int",
		}, {
			value:          int8(-2),
			expectedResult: "-2",
			desc:           "int8",
		}, {
			value:          int16(-3),
			expectedResult: "-3",
			desc:           "int16",
		}, {
			value:          int32(-4),
			expectedResult: "-4",
			desc:           "int32",
		}, {
			value:          int64(-5),
			expectedResult: "-5",
			desc:           "int64",
		}, {
			value:          uint(1),
			expectedResult: "1",
			desc:           "uint",
		}, {
			value:          uint8(2),
			expectedResult: "2",
			desc:           "uint8",
		}, {
			value:          uint16(3),
			expectedResult: "3",
			desc:           "uint16",
		}, {
			value:          uint32(4),
			expectedResult: "4",
			desc:           "uint32",
		}, {
			value:          uint64(5),
			expectedResult: "5",
			desc:           "uint64",
		}, {
			value:          float32(1.2345),
			expectedResult: "1.2345",
			desc:           "float32",
		}, {
			value:          1.23456789,
			expectedResult: "1.23456789",
			desc:           "float64",
		}, {
			value:          true,
			expectedResult: "true",
			desc:           "bool",
		}, {
			value:          "\"example000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"",
			expectedResult: `"\"example000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\""`,
			desc:           "string",
		}, {
			value:          sampleMarshaler{},
			expectedResult: "false",
			desc:           "json.Marshaler",
		}, {
			value:          []byte("hogehoge"),
			expectedResult: `"` + base64.StdEncoding.EncodeToString([]byte("hogehoge")) + `"`,
			desc:           "[]byte",
		}, {
			value:          []int{0, 1, 2, 3, 4},
			expectedResult: "[0,1,2,3,4]",
			desc:           "[]int",
		},
	}
	b.Run("encodeValue", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, testcase := range testcases {
				raw, err := encodeValue(testcase.value)
				if err != nil {
					b.Fatal("encode failed:", testcase.desc)
				} else if string(raw) != testcase.expectedResult {
					b.Fatal("invalid value:", testcase.desc)
				}
			}
		}
	})

	b.Run("json.Marshal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, testcase := range testcases {
				raw, err := json.Marshal(testcase.value)
				if err != nil {
					b.Fatal("encode failed:", testcase.desc)
				} else if string(raw) != testcase.expectedResult {
					b.Fatal("invalid value:", testcase.desc)
				}
			}
		}
	})
}

type sampleMarshaler struct{}

func (sm sampleMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("false"), nil
}
