package igc

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestAtoi(t *testing.T) {
	for _, tc := range []struct {
		dataStr     string
		expected    int
		expectedErr string
	}{
		{
			dataStr:     "",
			expectedErr: `"": syntax error`,
		},
		{
			dataStr:  "1",
			expected: 1,
		},
		{
			dataStr:  "12",
			expected: 12,
		},
		{
			dataStr:     "-",
			expectedErr: `"-": syntax error`,
		},
		{
			dataStr:  "-1",
			expected: -1,
		},
		{
			dataStr:  "12",
			expected: 12,
		},
	} {
		t.Run(tc.dataStr, func(t *testing.T) {
			actual, actualErr := atoi([]byte(tc.dataStr))
			assert.Equal(t, tc.expected, actual)
			assert.EqualError(t, actualErr, tc.expectedErr)
		})
	}
}

func TestIntPow(t *testing.T) {
	for _, tc := range []struct {
		x        int
		y        int
		expected int
	}{
		{x: 1, y: 0, expected: 1},
		{x: 1, y: 1, expected: 1},
		{x: 1, y: 2, expected: 1},
		{x: 2, y: 0, expected: 1},
		{x: 2, y: 1, expected: 2},
		{x: 2, y: 2, expected: 4},
		{x: 2, y: 3, expected: 8},
		{x: 3, y: 0, expected: 1},
		{x: 3, y: 1, expected: 3},
		{x: 3, y: 2, expected: 9},
		{x: 3, y: 3, expected: 27},
	} {
		t.Run(fmt.Sprintf("intPow(%d, %d)", tc.x, tc.y), func(t *testing.T) {
			assert.Equal(t, tc.expected, intPow(tc.x, tc.y))
		})
	}
}

func TestReservedCharacters(t *testing.T) {
	for c, expected := range map[byte]bool{
		'\r':   true,
		'\n':   true,
		' ':    false,
		'!':    true,
		'"':    false,
		'#':    false,
		'$':    true,
		'%':    false,
		'&':    false,
		'\'':   false,
		'(':    false,
		')':    false,
		'0':    false,
		'9':    false,
		':':    false,
		';':    false,
		'<':    false,
		'=':    false,
		'>':    false,
		'?':    false,
		'@':    false,
		'A':    false,
		'Z':    false,
		'[':    false,
		'\\':   true,
		']':    false,
		'^':    true,
		'_':    false,
		'`':    false,
		'a':    false,
		'z':    false,
		'{':    false,
		'|':    false,
		'}':    false,
		'~':    true,
		'*':    true,
		'\x80': true,
		'\xff': true,
	} {
		t.Run(string(c), func(t *testing.T) {
			assert.Equal(t, expected, invalidCharsRx.MatchString(string(c)))
		})
	}
}
