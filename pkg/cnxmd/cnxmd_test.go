package cnxmd

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {

	testData := []struct{
		data string
		err string
		consumed int
		kv map[string]string
	}{
		{
			`fdljkafjda`,
			"failed to read head line: EOF",
			0,
			map[string]string {},
		},
		{
			`fdljkafjda
`,
			"unexpected head line: fdljkafjda",
			0,
			map[string]string {},
		},
		{
			`CONNECTION_METADATA/1.1
`,
			"failed to read line: EOF",
			0,
			map[string]string {},
		},
		{
			`CONNECTION_METADATA/1.1

`,
			"",
			25,
			map[string]string {},
		},
		{
			`CONNECTION_METADATA/1.1
Yxhs=fdasj
invalid-kv-line

`,
			"invalid line: invalid-kv-line",
			0,
			map[string]string {},
		},
		{
			`CONNECTION_METADATA/1.1
foo=bar
joe=jane=jack

`,
			"",
			47,
			map[string]string {"foo":"bar", "joe":"jane=jack"},
		},
		{
			"CONNECTION_METADATA/1.1\n" + "x=" + strings.Repeat("y",
				250) + "\n\n",
			"",
			278,
			map[string]string {"x":strings.Repeat("y", 250)},
		},
		{
			// unprintable characters ok
			"CONNECTION_METADATA/1.1\n" + "\x01\x02=\x03\x04\n\n",
			"",
			31,
			map[string]string {"\x01\x02":"\x03\x04"},
		},
	}

	for i, entry := range testData {
		consumed, kv, err := Parse([]byte(entry.data))

		if entry.err != "" {
			if err == nil {
				t.Fatalf("%d: unexpected non-error", i)
			}
			if err.Error() != entry.err {
				t.Fatalf("%d: unexpected error message: %s", i, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%d: unexpected error: %s", i, err)
		}
		if entry.consumed != consumed {
			t.Fatalf("%d: actual consumed (%d) != expected (%d)", i,
				consumed, entry.consumed)
		}
		for k,v := range entry.kv {
			val := kv[k]
			if val != v {
				t.Fatalf("%d: value '%v' for key '%s' != expected '%v'",
					i, val, k, v)
			}
		}
	}
}