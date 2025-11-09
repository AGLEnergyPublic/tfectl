// Go equivalent of the _TsvOutput Class in the Python Knack package
package resources

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type TsvOutput struct{}

func NewTsvOutput() *TsvOutput {
	return &TsvOutput{}
}

func (*TsvOutput) dumpObj(data interface{}, stream *strings.Builder) error {
	var toWrite string
	switch v := data.(type) {
	case []interface{}:
		toWrite = strconv.Itoa(len(v))
	case map[string]interface{}:
		toWrite = ""
	case bool:
		toWrite = strconv.FormatBool(v)
	case string:
		stream.WriteString(v)
	default:
		toWrite = fmt.Sprint(data)
	}
	_, err := stream.WriteString(toWrite)
	return err
}

func (t *TsvOutput) dumpRow(data interface{}, stream *strings.Builder) error {
	separator := ""

	var values []interface{}
	var err error

	switch v := data.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		values = make([]interface{}, len(keys))
		for i, k := range keys {
			values[i] = v[k]
		}

	case []interface{}:
		values = v

	case bool:
		if err = t.dumpObj(v, stream); err != nil {
			return err
		}

		_, err := stream.WriteString("\n")
		return err

	default:
		if err = t.dumpObj(data, stream); err != nil {
			return err
		}
		stream.WriteString("\n")
		return err
	}

	for _, value := range values {
		if _, err = stream.WriteString(separator); err != nil {
			return err
		}
		if err = t.dumpObj(value, stream); err != nil {
			return err
		}
		separator = "\t"
	}
	_, err = stream.WriteString("\n")
	return err
}

func (t *TsvOutput) Dump(data []byte) (string, error) {
	var d []interface{}

	if err := json.Unmarshal(data, &d); err != nil {
		return "", fmt.Errorf("unable to unmarshal bytes to []interface{}: %w", err)
	}

	var io strings.Builder
	for _, item := range d {
		if err := t.dumpRow(item, &io); err != nil {
			return "", fmt.Errorf("failed to dump row: %w", err)
		}
	}

	return io.String(), nil
}
