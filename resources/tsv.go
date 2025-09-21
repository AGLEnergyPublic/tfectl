package resources

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jeremywohl/flatten"
)

func ToTsv(jsonStr []byte) ([]byte, error) {
	var data []interface{}
	var buf bytes.Buffer

	err := json.Unmarshal(jsonStr, &data)
	if err != nil {
		return nil, err
	}

	super := map[string]interface{}{
		"data": data,
	}

	flattened, err := flatten.Flatten(super, "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(&buf)
	writer.Comma = '\t'

	keys := make([]string, 0, len(flattened))
	for k := range flattened {
		keys = append(keys, strings.ReplaceAll(k, "data.0", ""))
	}

	writer.Write(keys)

	values := make([]string, len(keys))
	for i, key := range keys {
		values[i] = fmt.Sprintf("%v", flattened[fmt.Sprintf("data.0%s", key)])
	}

	writer.Write(values)

	writer.Flush()
	cleanBuf := removeEmptyLines(&buf)
	return cleanBuf.Bytes(), nil
}
