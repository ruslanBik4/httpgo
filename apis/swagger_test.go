package apis

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type testStruct struct {
	string
	int
}

func TestNewReflectType(t *testing.T) {
	tests := []struct {
		name string
		args interface{}
		want *ReflectType
	}{
		// TODO: Add test cases.
		{
			"simple",
			1,
			&ReflectType{Type: reflect.TypeOf(1), Props: spec.SchemaProps{Type: []string{"int"}, Default: 1}},
		},
		{
			"struct",
			testStruct{"1", 2},
			&ReflectType{Type: reflect.TypeOf(testStruct{}), Props: spec.SchemaProps{Type: []string{"struct"}}},
		},
		{
			"array",
			[]int{1, 2, 3},
			&ReflectType{Type: reflect.TypeOf([]int{}), Props: spec.SchemaProps{Type: []string{"array"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := NewReflectType(tt.args)
			require.NotNil(t, rt.Props)
			assert.Equalf(t, tt.want, rt, "NewReflectType(%v)", tt.args)
		})
	}
}

func TestNewSwaggerArray(t *testing.T) {
	type args struct {
		desc  string
		props []interface{}
	}
	tests := []struct {
		name string
		args args
		want SwaggerParam
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerArray(tt.args.desc, tt.args.props), "NewSwaggerArray(%v, %v)", tt.args.desc, tt.args.props)
		})
	}
}

func TestNewSwaggerArray1(t *testing.T) {
	type args struct {
		props interface{}
		name  string
	}
	tests := []struct {
		name string
		args args
		want SwaggerParam
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerArray1(tt.args.props, tt.args.name), "NewSwaggerArray1(%v, %v)", tt.args.props, tt.args.name)
		})
	}
}

func TestNewSwaggerContent(t *testing.T) {
	type args struct {
		schema map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerContent(tt.args.schema), "NewSwaggerContent(%v)", tt.args.schema)
		})
	}
}

func TestNewSwaggerObject(t *testing.T) {
	type args struct {
		props interface{}
		name  string
	}
	tests := []struct {
		name string
		args args
		want SwaggerParam
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerObject(tt.args.props, tt.args.name), "NewSwaggerObject(%v, %v)", tt.args.props, tt.args.name)
		})
	}
}

func TestNewSwaggerObjectRoot(t *testing.T) {
	type args struct {
		props interface{}
	}
	tests := []struct {
		name string
		args args
		want SwaggerParam
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerObjectRoot(tt.args.props), "NewSwaggerObjectRoot(%v)", tt.args.props)
		})
	}
}

func TestNewSwaggerParam(t *testing.T) {
	type args struct {
		props interface{}
		name  string
		typ   string
	}
	tests := []struct {
		name string
		args args
		want SwaggerParam
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewSwaggerParam(tt.args.name, "body", tt.args.typ, "g"), "NewSwaggerParam(%v, %v, %v)", tt.args.props, tt.args.name, tt.args.typ)
		})
	}
}

func TestReflectType_CheckType(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		ctx   *fasthttp.RequestCtx
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.CheckType(tt.args.ctx, tt.args.value), "CheckType(%v, %v)", tt.args.ctx, tt.args.value)
		})
	}
}

func TestReflectType_ConvertSlice(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		ctx    *fasthttp.RequestCtx
		values []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			got, err := r.ConvertSlice(tt.args.ctx, tt.args.values)
			if !tt.wantErr(t, err, fmt.Sprintf("ConvertSlice(%v, %v)", tt.args.ctx, tt.args.values)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConvertSlice(%v, %v)", tt.args.ctx, tt.args.values)
		})
	}
}

func TestReflectType_ConvertValue(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		ctx   *fasthttp.RequestCtx
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			got, err := r.ConvertValue(tt.args.ctx, tt.args.value)
			if !tt.wantErr(t, err, fmt.Sprintf("ConvertValue(%v, %v)", tt.args.ctx, tt.args.value)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ConvertValue(%v, %v)", tt.args.ctx, tt.args.value)
		})
	}
}

func TestReflectType_IsSlice(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.IsSlice(), "IsSlice()")
		})
	}
}

func TestReflectType_WriteMap(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		value reflect.Value
		title string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.WriteMap(tt.args.value, tt.args.title), "WriteMap(%v, %v)", tt.args.value, tt.args.title)
		})
	}
}

func TestReflectType_WriteReflectKind(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		kind  reflect.Kind
		value reflect.Value
		sType string
		title string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.WriteReflectKind(tt.args.kind, tt.args.value, tt.args.sType, tt.args.title), "WriteReflectKind(%v, %v, %v, %v)", tt.args.kind, tt.args.value, tt.args.sType, tt.args.title)
		})
	}
}

func TestReflectType_WriteSlice(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		value reflect.Value
		title string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.WriteSlice(tt.args.value, tt.args.title), "WriteSlice(%v, %v)", tt.args.value, tt.args.title)
		})
	}
}

func TestReflectType_WriteStruct(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		value reflect.Value
		title string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.WriteStruct(tt.args.value, tt.args.title), "WriteStruct(%v, %v)", tt.args.value, tt.args.title)
		})
	}
}

func TestReflectType_convertValue(t *testing.T) {
	type fields struct {
		Type  reflect.Type
		value reflect.Value
		Props interface{}
	}
	type args struct {
		title string
		value reflect.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReflectType{
				Type:  tt.fields.Type,
				value: tt.fields.value,
				Props: tt.fields.Props,
			}
			assert.Equalf(t, tt.want, r.convertValue(tt.args.title, tt.args.value), "convertValue(%v, %v)", tt.args.title, tt.args.value)
		})
	}
}
