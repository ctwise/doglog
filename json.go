package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"os"
	"strings"
)

// Retrieve a single value from the json buffer.
func getJSONValue(data []byte, keys ...string) (slice []byte, dataType jsonparser.ValueType, err error) {
	slice, dataType, _, err = jsonparser.Get(data, keys...)
	return slice, dataType, err
}

// Retrieve a single string value from the json buffer.
func getJSONString(data []byte, keys ...string) string {
	value, err := jsonparser.GetString(data, keys...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to retrieve string for keys: %v - %s\n", keys, err.Error())
		return ""
	}
	return Expand(value)
}

// Retrieve an array structure from the json buffer.
func getJSONArray(data []byte, keys ...string) []byte {
	slice, dataType, err := getJSONValue(data, keys...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to retrieve value for keys: %v - %s\n", keys, err.Error())
	} else if dataType != jsonparser.Array {
		_, _ = fmt.Fprintf(os.Stderr, "Key did not reference an array: %v\n", keys)
	} else {
		return slice
	}
	return []byte{}
}

// Retrieve a parsed array of strings from the json buffer.
func getJSONArrayOfStrings(data []byte, keys ...string) []string {
	arraySlice := getJSONArray(data, keys...)
	var stringList []string
	_, _ = jsonparser.ArrayEach(arraySlice, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if dataType == jsonparser.String || dataType == jsonparser.Number || dataType == jsonparser.Boolean {
			stringList = append(stringList, Expand(string(value)))
		}
	})
	return stringList
}

// Retrieve a parsed map of values from the json buffer. Numbers and booleans are converted to strings.
func getJSONSimpleMap(data []byte, keys ...string) map[string]string {
	result := make(map[string]string)
	_ = levelPass(data, "", result, keys)
	return result
}

func levelPass(data []byte, path string, result map[string]string, keys []string) error {
	return jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		skey := string(key)
		if len(path) > 0 {
			skey = path + "_" + skey
		}
		if dataType == jsonparser.String || dataType == jsonparser.Number || dataType == jsonparser.Boolean {
			result[skey] = Expand(string(value))
		} else if dataType == jsonparser.Object {
			// Don't count 'attributes' as part of the path
			if skey == "attributes" {
				skey = ""
			}
			err := levelPass(value, skey, result, []string{})
			if err != nil {
				return err
			}
		}
		return nil
	}, keys...)
}

// Expand escape strings. JSON strings from Datadog have embedded escape sequences that aren't getting expanded. We
// have to do it manually.
func Expand(value string) string {
	var result string
	result = strings.Replace(value, "\\n", "\n", -1)
	result = strings.Replace(result, "\\r", "\r", -1)
	result = strings.Replace(result, "\\t", "\t", -1)
	return result
}
