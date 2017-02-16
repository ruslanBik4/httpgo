package json


func getType(value interface {} ) string {
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
