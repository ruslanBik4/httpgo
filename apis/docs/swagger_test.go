package docs

import "testing"

func TestLoadSpec(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{"simple"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoadSpec()
		})
	}
}
