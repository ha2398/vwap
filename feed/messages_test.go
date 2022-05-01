// +build unit

package feed

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func Test_GetValueForKey(t *testing.T) {
	testCases := []struct {
		desc           string
		message        *Message
		key            string
		expectedOutput string
	}{
		{
			desc:           "empty message, empty key",
			message:        &Message{},
			key:            "",
			expectedOutput: "",
		},
		{
			desc:           "nil message",
			message:        nil,
			key:            "",
			expectedOutput: "",
		},
		{
			desc: "absent key",
			message: &Message{
				"someKey": "someValue",
			},
			key:            "myKey",
			expectedOutput: "",
		},
		{
			desc: "key present, string value",
			message: &Message{
				"someKey": "someValue",
			},
			key:            "someKey",
			expectedOutput: "someValue",
		},
		{
			desc: "key present, non-string value",
			message: &Message{
				"someKey": 12.34,
			},
			key:            "someKey",
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		output := tc.message.GetValueForKey(tc.key)
		assert.Equal(t, tc.expectedOutput, output,
			"For test %q, got wrong output", tc.desc)
	}
}

func Test_ParseMatch(t *testing.T) {
	testCases := []struct {
		desc             string
		message          Message
		expectedMatch    Match
		expectedHasMatch bool
		expectedError    error
	}{
		{
			desc:             "empty message",
			message:          Message{},
			expectedMatch:    Match{},
			expectedHasMatch: false,
			expectedError:    nil,
		},
		{
			desc: "message with no-match type",
			message: Message{
				TypeKey: "someType",
			},
			expectedMatch:    Match{},
			expectedHasMatch: false,
			expectedError:    nil,
		},
		{
			desc: "Match message invalid price value",
			message: Message{
				TypeKey:  MatchType,
				PriceKey: "hello world",
			},
			expectedMatch:    Match{},
			expectedHasMatch: true,
			expectedError:    errors.New("error parsing \"price\" field: strconv.ParseFloat: parsing \"hello world\": invalid syntax"),
		},
		{
			desc: "Match message invalid size value",
			message: Message{
				TypeKey:  MatchType,
				PriceKey: "1.23",
				SizeKey:  "hello world",
			},
			expectedMatch:    Match{},
			expectedHasMatch: true,
			expectedError:    errors.New("error parsing \"size\" field: strconv.ParseFloat: parsing \"hello world\": invalid syntax"),
		},
		{
			desc: "Match message with valid fields",
			message: Message{
				TypeKey:      MatchType,
				PriceKey:     "1.23",
				SizeKey:      "4.56",
				ProductIDKey: "myProduct",
			},
			expectedMatch: Match{
				IsLast:    false,
				Price:     1.23,
				Size:      4.56,
				ProductID: "myProduct",
			},
			expectedHasMatch: true,
			expectedError:    nil,
		},
	}

	for _, tc := range testCases {
		match, hasMatch, err := ParseMatch(tc.message)
		assert.True(t, cmp.Equal(match, tc.expectedMatch),
			"For test %q, got wrong Match", tc.desc)
		assert.Equal(t, tc.expectedHasMatch, hasMatch,
			"For test %q, got wrong hasMatch", tc.desc)
		assert.Equal(t, tc.expectedError, err,
			"For test %q, got unexpected error value", tc.desc)
	}
}
