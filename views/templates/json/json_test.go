package json

import (
	"encoding/json"
	"testing"
)

func TestWriteAnyJSON(t *testing.T) {
	var input = map[string]interface{}{
		//"one" : StringDimension {"1","2"},
		"two":  true,
		"thre": 3,
		"5":    map[string]interface{}{"1": "6", "4": "l", "5": "u"},
	}
	result := make(map[string]interface{}, 10)

	strJSON := AnyJSON(input)
	if err := json.Unmarshal([]byte(strJSON), &result); err != nil {
		//!= `{"one":[0:"1",1:"2"],"two":true,"thre":3}` {
		t.Errorf("WriteAnyJSON=%v, stroka=%s, error=%q", input, strJSON, err)
	} else if !equalMaps(input, result) {
		t.Errorf("WriteAnyJSON=%v, result=%v", input, result)
	}

	t.Skipped()
}

func equalMaps(input, result map[string]interface{}) bool {
	for key, val := range input {
		if valResult, ok := result[key]; !ok {
			return false
		} else {
			switch vv := val.(type) {
			case map[string]interface{}:
				switch vvResult := valResult.(type) {
				case map[string]interface{}:
					for key, val := range vv {
						if valResult, ok := vvResult[key]; !ok {
							return false
						} else if val != valResult {
							return false

						}
					}
				default:
					return false
				}
			default:
				if val != valResult {
					return false
				}

			}
		}
	}

	return true
}
