package json

import (
	"testing"
	"encoding/json"
)

func TestWriteAnyJSON(t *testing.T) {
	var input = MultiDimension {
		//"one" : StringDimension {"1","2"},
		"two" : true,
		"thre": 3,
		"5"   : MultiDimension {"1":"6", "4": "l", "5": "u"},
	}
	result := make( map[string] interface{}, 10 )



	strJSON := WriteAnyJSON(input)
	if err := json.Unmarshal([]byte(strJSON), &result); err != nil {
	//!= `{"one":[0:"1",1:"2"],"two":true,"thre":3}` {
		t.Errorf("WriteAnyJSON=%v, stroka=%s, error=%q", input, strJSON, err)
	} else if !equalMaps(input, result) {
		t.Errorf("WriteAnyJSON=%v, result=%v", input, result)
	}
}

func equalMaps(input, result map[string] interface{}) bool {
	for key, val := range input {
		if valResult, ok := result[key]; !ok {
			return false
		} else {
			switch vv := val.(type) {
			case map[string] interface{}:
				switch vvResult := valResult.(type) {
				case map[string] interface{} :
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