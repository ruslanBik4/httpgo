package json

import "log"

func getType(value interface {} ) string {
	log.Println(value)
	switch value.(type) {
	case map[string]interface{}:
		return "array"
	case string:
		return "string"
	case int:
		return "int"
	default:
		return value.(string)
	}
}
