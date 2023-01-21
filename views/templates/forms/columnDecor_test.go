// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package forms

import (
	"fmt"
	"testing"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/stretchr/testify/assert"
)

const (
	testURL     = "url"
	testID      = 123
	testName    = "options"
	testPattern = `\d\S\s ame`
	comment     = `label {"pattern": "%s","multiple": true, "suggestions":"%s","suggestions_params":{"name":"%s"}, "id":%d } "read_only"`
)

func TestNewColumnDecor(t *testing.T) {
	column := dbEngine.NewStringColumn("test", fmt.Sprintf(comment, testPattern, testURL, testName, testID), true)
	colDev := NewColumnDecor(column, nil)
	assert.Implements(t, (*dbEngine.Column)(nil), colDev)
	assert.Equal(t, testPattern, colDev.Pattern())
	assert.Equal(t, testURL, colDev.Suggestions)
	assert.Equal(t, testName, colDev.SuggestionsParams["name"])
	assert.True(t, colDev.Required())
	assert.True(t, colDev.IsReadOnly)
	assert.True(t, colDev.multiple)
}

func TestColumnDecor_GetValues(t *testing.T) {
	type fields struct {
		Column        dbEngine.Column
		IsHidden      bool
		IsReadOnly    bool
		IsSlice       bool
		InputType     string
		SelectOptions map[string]SelectOption
		PatternList   dbEngine.Table
		PatternName   string
		PlaceHolder   string
		Label         string
		pattern       string
		patternDesc   string
		Value         interface{}
	}
	tests := []struct {
		name       string
		fields     fields
		wantValues []interface{}
	}{
		// TODO: Add test cases.
		{
			"1",
			fields{
				Column: dbEngine.NewStringColumn("name", "name", true),
				Value:  []string{"1", "2", "3"},
			},
			[]interface{}{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:        tt.fields.Column,
				IsHidden:      tt.fields.IsHidden,
				IsReadOnly:    tt.fields.IsReadOnly,
				IsSlice:       tt.fields.IsSlice,
				InputType:     tt.fields.InputType,
				SelectOptions: tt.fields.SelectOptions,
				PatternList:   tt.fields.PatternList,
				PatternName:   tt.fields.PatternName,
				PlaceHolder:   tt.fields.PlaceHolder,
				Label:         tt.fields.Label,
				pattern:       tt.fields.pattern,
				patternDesc:   tt.fields.patternDesc,
				Value:         tt.fields.Value,
			}

			gotValues := col.GetValues()
			assert.Equal(t, tt.wantValues, gotValues)
		})
	}
}
