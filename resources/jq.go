package resources

import (
	"encoding/json"
	"fmt"
	"reflect"

	jq "github.com/itchyny/gojq"
	log "github.com/sirupsen/logrus"
)

// Constructor
var inputConstr func() interface{}

func JqRun(jsonStr []byte, query string) {
	q, err := jq.Parse(query)
	if err != nil {
		log.Fatal(err)
	}

	rt := reflect.TypeOf(jsonStr)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		inputConstr = func() interface{} {
			c := new([]interface{})
			return *c
		}
	case reflect.Map:
		inputConstr = func() interface{} {
			c := new(map[string]interface{})
			return *c
		}
	default:
		log.Fatalf("input stream cannot be parsed!")
	}

	var input = inputConstr()
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
