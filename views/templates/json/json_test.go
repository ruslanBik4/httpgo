package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAnyJSONOld(t *testing.T) {
	var input = map[string]interface{}{
		// "one" : StringDimension {"1","2"},
		"two":  true,
		"thre": 3.00,
		"5":    map[string]interface{}{"1": "6", "4": "l", "5": "u"},
	}
	result := make(map[string]interface{}, 10)

	strJSON := AnyJSON(input)
	err := json.Unmarshal([]byte(strJSON), &result)
	if assert.Nil(t, err, "WriteAnyJSON=%v, stroka=%s", input, strJSON) {
		assert.Equal(t, input, result)
	}

	t.Skipped()
}

func equalMaps(input, result map[string]interface{}) bool {
	for key, val := range input {
		valResult, ok := result[key]
		if !ok {
			return false
		}

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

	return true
}
