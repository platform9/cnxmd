package cnxmd

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

const DefaultBufferLength = 256
const LineDelim = '\n'
const KvDelim = "="
const HeadLine = "CONNECTION_METADATA/1.1"

// Parse attempts to read a CNXMD header from the supplied byte slice.
// If successful, err == nil, bytesConsumed is the total length of
// the header, and kv is the map of key value pairs.
// If a header cannot be read, err != nil, and the other two return
// parameters are undefined.
func Parse(data []byte) (
	bytesConsumed int,
	kv map[string]string,
	err error,
) {
	buf := 	bytes.NewBuffer(data)
	r := bufio.NewReaderSize(buf, DefaultBufferLength)
	head, err := r.ReadString(LineDelim)	
	if err != nil {
		err = fmt.Errorf("failed to read head line: %s", err)
		return
	}
	l := len(head)
	head = head[:l-1]
	if head != HeadLine {
		err = fmt.Errorf("unexpected head line: %s", head)
		return
	}
	bytesConsumed += l
	kv = make(map[string]string)
	for {
		var line string
		line, err = r.ReadString(LineDelim)
		if err != nil {
			err = fmt.Errorf("failed to read line: %s", err)
			return
		}
		l = len(line)
		bytesConsumed += l
		line = line[:l-1]
		if line == "" {
			return // done
		}
		components := strings.Split(line, KvDelim)
		if len(components) < 2 {
			err = fmt.Errorf("invalid line: %s", line)
			return
		}
		val := strings.Join(components[1:len(components)], KvDelim)
		kv[components[0]] = val
	}
}
