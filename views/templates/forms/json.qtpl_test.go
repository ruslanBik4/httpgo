// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package forms

import (
	"bytes"
	"testing"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/stretchr/testify/assert"
	qt422016 "github.com/valyala/quicktemplate"
)

func TestColumnDecor_DataForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			"id_photos",
			fields{
				Attachments: []AttachmentList{
					{
						1,
						"https://site.com/photots/",
					},
				},
			},
			`{"id":1, "url":"https://site.com/photots/"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			got := col.DataForJSON()
			t.Log(got)
			assert.Equal(t, tt.want, got, "DataForJSON() = %v, want %v")

		})
	}
}

func TestColumnDecor_InputTypeForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			if got := col.InputTypeForJSON(); got != tt.want {
				t.Errorf("InputTypeForJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnDecor_RenderAttr(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		i int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			if got := col.RenderAttr(tt.args.i); got != tt.want {
				t.Errorf("RenderAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnDecor_RenderInputs(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			if got := col.RenderInputs(tt.args.data); got != tt.want {
				t.Errorf("RenderInputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnDecor_RenderValue(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			if got := col.RenderValue(tt.args.value); got != tt.want {
				t.Errorf("RenderValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumnDecor_StreamDataForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		qw422016 *qt422016.Writer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			w := bytes.NewBufferString("")
			col.WriteDataForJSON(w)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestColumnDecor_StreamInputTypeForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		qw422016 *qt422016.Writer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			w := bytes.NewBufferString("")
			col.WriteInputTypeForJSON(w)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestColumnDecor_StreamRenderAttr(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		qw422016 *qt422016.Writer
		i        int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			w := bytes.NewBufferString("")
			col.WriteRenderAttr(w, 0)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestColumnDecor_StreamRenderInputs(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		qw422016 *qt422016.Writer
		data     map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			w := bytes.NewBufferString("")
			col.WriteRenderInputs(w, nil)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestColumnDecor_StreamRenderValue(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		qw422016 *qt422016.Writer
		value    interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			w := bytes.NewBufferString("")
			col.WriteRenderValue(w, nil)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestColumnDecor_WriteDataForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	tests := []struct {
		name         string
		fields       fields
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			qq422016 := &bytes.Buffer{}
			col.WriteDataForJSON(qq422016)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteDataForJSON() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestColumnDecor_WriteInputTypeForJSON(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	tests := []struct {
		name         string
		fields       fields
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			qq422016 := &bytes.Buffer{}
			col.WriteInputTypeForJSON(qq422016)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteInputTypeForJSON() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestColumnDecor_WriteRenderAttr(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		i int
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			qq422016 := &bytes.Buffer{}
			col.WriteRenderAttr(qq422016, tt.args.i)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteRenderAttr() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestColumnDecor_WriteRenderInputs(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			qq422016 := &bytes.Buffer{}
			col.WriteRenderInputs(qq422016, tt.args.data)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteRenderInputs() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestColumnDecor_WriteRenderValue(t *testing.T) {
	type fields struct {
		Column            dbEngine.Column
		IsHidden          bool
		IsDisabled        bool
		IsReadOnly        bool
		IsSlice           bool
		IsNewPrimary      bool
		SelectWithNew     bool
		InputType         string
		DefaultInputValue string
		Attachments       []AttachmentList
		SelectOptions     map[string]string
		PatternList       dbEngine.Table
		PatternName       string
		PlaceHolder       string
		LinkNew           string
		Label             string
		pattern           string
		patternDesc       string
		Value             interface{}
		Suggestions       string
	}
	type args struct {
		value interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &ColumnDecor{
				Column:            tt.fields.Column,
				IsHidden:          tt.fields.IsHidden,
				IsDisabled:        tt.fields.IsDisabled,
				IsReadOnly:        tt.fields.IsReadOnly,
				IsSlice:           tt.fields.IsSlice,
				IsNewPrimary:      tt.fields.IsNewPrimary,
				SelectWithNew:     tt.fields.SelectWithNew,
				InputType:         tt.fields.InputType,
				DefaultInputValue: tt.fields.DefaultInputValue,
				Attachments:       tt.fields.Attachments,
				SelectOptions:     tt.fields.SelectOptions,
				PatternList:       tt.fields.PatternList,
				PatternName:       tt.fields.PatternName,
				PlaceHolder:       tt.fields.PlaceHolder,
				LinkNew:           tt.fields.LinkNew,
				Label:             tt.fields.Label,
				pattern:           tt.fields.pattern,
				patternDesc:       tt.fields.patternDesc,
				Value:             tt.fields.Value,
				Suggestions:       tt.fields.Suggestions,
			}
			qq422016 := &bytes.Buffer{}
			col.WriteRenderValue(qq422016, tt.args.value)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteRenderValue() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestFormField_FormHTML(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		blocks []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			if got := f.FormHTML(tt.args.blocks...); got != tt.want {
				t.Errorf("FormHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormField_FormJSON(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
		blocks      []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			"id_photos",
			fields{
				blocks: []BlockColumns{
					{
						Columns: []*ColumnDecor{
							{
								Column: dbEngine.NewStringColumn("id_photos", "test column", false),
								Attachments: []AttachmentList{
									{
										1,
										"https://site.com/photots/",
									},
								},
							},
						},
					},
				},
			},
			`{"title" : "","action": "","description": "","method": "","blocks": [{"id": "0","title": "","description": "","fields": [{"name": "id_photos","required":false,"type": "", "value":"", "title": "", "list":[{"id":1,"url":"https://site.com/photots/"}]
}],"actions": [{"groups": []}]}]}`,
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
			}
			got := f.FormJSON(tt.fields.blocks...)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestFormField_RenderForm(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		isHTML bool
		blocks []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			if got := f.RenderForm(tt.args.isHTML, tt.args.blocks...); got != tt.want {
				t.Errorf("RenderForm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormField_StreamFormHTML(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		qw422016 *qt422016.Writer
		blocks   []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			w := bytes.NewBufferString("")
			f.WriteFormHTML(w)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestFormField_StreamFormJSON(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		qw422016 *qt422016.Writer
		blocks   []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			w := bytes.NewBufferString("")
			f.WriteFormJSON(w)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestFormField_StreamRenderForm(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		qw422016 *qt422016.Writer
		isHTML   bool
		blocks   []BlockColumns
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			w := bytes.NewBufferString("")
			f.WriteRenderForm(w, false)
			assert.Equal(t, tt.name, w.String())
		})
	}
}

func TestFormField_WriteFormHTML(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		blocks []BlockColumns
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			qq422016 := &bytes.Buffer{}
			f.WriteFormHTML(qq422016, tt.args.blocks...)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteFormHTML() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestFormField_WriteFormJSON(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		blocks []BlockColumns
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			qq422016 := &bytes.Buffer{}
			f.WriteFormJSON(qq422016, tt.args.blocks...)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteFormJSON() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}

func TestFormField_WriteRenderForm(t *testing.T) {
	type fields struct {
		Title       string
		Action      string
		Method      string
		Description string
		HideBlock   interface{}
	}
	type args struct {
		isHTML bool
		blocks []BlockColumns
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantQq422016 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FormField{
				Title:       tt.fields.Title,
				Action:      tt.fields.Action,
				Method:      tt.fields.Method,
				Description: tt.fields.Description,
				HideBlock:   tt.fields.HideBlock,
			}
			qq422016 := &bytes.Buffer{}
			f.WriteRenderForm(qq422016, tt.args.isHTML, tt.args.blocks...)
			if gotQq422016 := qq422016.String(); gotQq422016 != tt.wantQq422016 {
				t.Errorf("WriteRenderForm() = %v, want %v", gotQq422016, tt.wantQq422016)
			}
		})
	}
}
