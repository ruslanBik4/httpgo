package json

import "log"

func getType(value interface {} ) string {
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


	default:
		return value.(string)
	}
}
