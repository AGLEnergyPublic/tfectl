package resources

import (
	"bytes"
)

func removeEmptyLines(buf *bytes.Buffer) *bytes.Buffer {
	if buf == nil {
		return bytes.NewBuffer(nil)
	}

	data := buf.Bytes()
	lines := bytes.Split(data, []byte("\n"))
	var nonEmptyLines [][]byte

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) > 0 {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	cleaned := bytes.Join(nonEmptyLines, []byte("\n"))
	return bytes.NewBuffer(cleaned)
}
