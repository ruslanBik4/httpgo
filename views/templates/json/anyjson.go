package json

import (

	"strings"
	"github.com/ruslanBik4/httpgo/models/logs"
)

type MultiDimension map[string] interface{}
type MapMultiDimension [] map[string] interface{}
type SimpleDimension [] interface{}
type StringDimension [] string

func isSimpleDimension(value interface{}) ([] interface{}, bool) {
	switch vv := value.(type) {
	case [] interface{}:
		return vv, true
	}

	return nil, false
}
func isMultiDimension(value interface{}) (map[string] interface{}, bool) {
	switch vv := value.(type) {
	case map[string] interface{}:
		return vv, true
	}

	return nil, false
}
func isMapMultiDimension(value interface{}) ([] map[string] interface{}, bool) {
	switch vv := value.(type) {
	case [] map[string] interface{}:
		return vv, true
	}

	return nil, false
}

func writeElement(value interface {} ) string {
	logs.DebugLog("value=",value)
	switch value.(type) {
	case map[string]interface{}:
		return "map"
	case string:
		return "string"
	case int:
		return "int"
	case bool:
		return "bool"
	case float64:
		return "float64"
	case nil:
		return "nil"
	case []interface{}:
		return "array"
	case []string:
		return "array"
	default:
		return value.(string)
	}
}

func PrepareString(str string) string{
	replacer := strings.NewReplacer(str)
	return replacer.Replace(str)
}
