package json

import (
	"log"
)

type MultiDimension map[string] interface{}
type MapMultiDimension [] map[string] interface{}
type SimpleDimension [] interface{}
type StringDimension [] string

func isSimpleDimension(value interface{}) bool {
	switch value.(type) {
	case []string:
		return true
	case []int:
		return true
	case [] interface{}:
		return true
	}

	return false
}
func writeElement(value interface {} ) string {
	log.Println(value)
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
