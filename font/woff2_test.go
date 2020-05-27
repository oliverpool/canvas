package font

import (
	"fmt"
	"testing"

	"github.com/tdewolff/test"
)

func TestWOFF2Error(t *testing.T) {
	var tts = []struct {
		data string
		err  string
	}{
		{"wOF200000000\x00\x00000000\xff\xff\xff\xff000000000000000000000000", "length in header must match file size"},
		{"wOF200000000\x00\x00000000\x00\x00\x00\b00000000000000000000000030000000", "length in header must match file size"},
		{"wOF200000000\x00\x01000000\x00\x00\x000000000000000000000000000Y\xbf\x00\x00Z\x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "length in header must match file size"},
	}
	for i, tt := range tts {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := ParseWOFF2([]byte(tt.data))
			test.T(t, err.Error(), tt.err)
		})
	}
}
