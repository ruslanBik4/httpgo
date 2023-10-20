/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package forms

import (
	"bytes"
	"testing"
)

func TestFormField_WriteCreate(t *testing.T) {
	tests := []struct {
		name         string
		fields       FormField
		wantQq422016 string
	}{
		{
			"simple",
			FormField{
				Title:       "title",
				Action:      "/test",
				Description: "test",
				Method:      "get",
				HideBlock: map[string]any{
					"data": map[string]any{
						"content_filler": map[string][]string{
							"1": {
								"13",
								"15",
								"16",
								"19",
								"21",
								"-1",
							},
							"2": {
								"13",
								"15",
								"16",
								"20",
								"21",
								"-1",
							},
						},
						"id_type_material": map[string][]string{
							"1": {
								"13",
								"14",
								"21",
								"-1",
							},
							"2": {
								"13",
								"15",
								"21",
								"-1",
							},
							"3": {
								"13",
								"12",
								"18",
								"21",
								"-1",
							},
							"4": {
								"13",
								"15",
								"16",
								"18",
								"21",
								"-1",
							},
							"6": {
								"13",
								"17",
								"21",
								"-1",
							},
							"7": {
								"11",
								"4",
								"-1",
							},
							"8": {
								"11",
								"22",
								"-1",
							},
						},
					},
					"defaultBlocks": []string{
						"11",
					},
				},
				Blocks: []BlockColumns{
					{
						Id: 1,
						Buttons: []Button{
							{
								Title:    "default",
								Position: false,
								Type:     "submit",
								OnClick:  "return false",
							},
						},
						Columns: []*ColumnDecor{
							{
								Column:            nil,
								IsHidden:          false,
								IsDisabled:        false,
								IsReadOnly:        false,
								IsSlice:           false,
								IsNewPrimary:      false,
								SelectWithNew:     false,
								ExtProperties:     nil,
								InputType:         "select",
								SpecialInputName:  "input_name",
								DefaultInputValue: "",
								Attachments:       nil,
								SelectOptions:     nil,
								PatternList:       nil,
								PatternName:       "",
								PlaceHolder:       "",
								LinkNew:           "",
								Label:             "",
								multiple:          false,
								pattern:           "",
								patternDesc:       "",
								Value:             nil,
								Suggestions:       "",
								SuggestionsParams: nil,
							},
						},
						Multiple:    true,
						Title:       "first figure",
						Description: "test description",
					},
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
				Blocks:      tt.fields.Blocks,
			}
			buf := bytes.NewBuffer(nil)
			f.WriteCreate(buf, "httpgo", "forms", "test")
		})
	}
}
