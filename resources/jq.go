package resources

import (
	"encoding/json"
	"fmt"
	jq "github.com/itchyny/gojq"
	log "github.com/sirupsen/logrus"
)

func JqRun(jsonStr []byte, query string) {
	q, err := jq.Parse(query)
	if err != nil {
		log.Fatal(err)
	}

	var input []interface{}
	var output []interface{}
	var outputJsonStr []byte

	err = json.Unmarshal(jsonStr, &input)
	if err != nil {
		log.Fatal(err)
	}

	jqIterator := q.Run(input)
	for {
		v, ok := jqIterator.Next()
		if !ok {
			break
			// next element
		}
		if err, ok := v.(error); ok {
			log.Fatal(err)
		}
		output = append(output, v)
	}
	outputJsonStr, _ = json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJsonStr))
}
