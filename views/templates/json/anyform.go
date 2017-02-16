package json


func isArray(value interface {} ) bool {
	switch value.(type) {
	case map[string]interface{}:
		return true
	default:
		return false
	}
}
