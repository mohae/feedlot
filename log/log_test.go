package log

import (
	"log"
	"testing"
)

func TestParseLogFlag(t *testing.T) {
	tests := []struct {
		v        string
		expected int
		err      error
	}{
		{"", 0, UnknownLogFlagErr{}},
		{"date", 0, UnknownLogFlagErr{"date"}},
		{"Ldate", log.Ldate, nil},
		{"LTIME", log.Ltime, nil},
		{"lmicroseconds", log.Lmicroseconds, nil},
		{"llongfile", log.Llongfile, nil},
		{"LShortFile", log.Lshortfile, nil},
		{"lUTC", log.LUTC, nil},
		{"lstdflags", log.LstdFlags, nil},
		{"none", 0, nil},
	}

	for _, test := range tests {
		v, err := ParseLogFlag(test.v)
		if err != nil {
			if err != test.err {
				t.Errorf("%q: got %s; want %s", test.v, err, test.err)
			}
			continue
		}
		if test.err != nil {
			t.Errorf("%q: got no error; wanted %s", test.v, test.err)
			continue
		}
		if v != test.expected {
			t.Errorf("%q: got %d; want %d", test.v, v, test.expected)
		}
	}
}
