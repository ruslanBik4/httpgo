package apis

import (
	//"bufio"
	//"go/types"
	//"net"
	//"sync"
	"encoding/json"
	"go/types"
	"testing"

	//"github.com/json-iterator/go"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/dbEngine/dbEngine/psql"
	"github.com/ruslanBik4/dbEngine/typesExt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type commCase string

type PRCommandParams struct {
	Command   commCase `json:"command"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Account   int32    `json:"account"`
	LastQuery commCase `json:"last_query"`
}

// Implementing RouteDTO interface
func (prParams *PRCommandParams) GetValue() interface{} {
	return prParams
}

func (prParams *PRCommandParams) NewValue() interface{} {

	newVal := PRCommandParams{}

	return newVal

}

const jsonText = `{"account":7060246,"command":"adjustments","end_date":"2020-01-25","start_date":"2020-01-01"}`

var (
	route = &ApiRoute{
		Desc:   "test route",
		Method: POST,
		DTO:    &PRCommandParams{},
	}
)

func TestCheckAndRun(t *testing.T) {

	dto := route.DTO.NewValue()
	val := &dto
	//err := jsoniter.UnmarshalFromString(json, &val)

	//assert.Nil(t, err)

	//t.Logf("%+v", DTO)

	err := json.Unmarshal([]byte(jsonText), &val)

	assert.Nil(t, err)

	t.Logf("%+v", dto)
}

var tests = []struct {
	name string
	src  string
	col  types.BasicKind
	want string
}{
	{
		"string",
		"simple string",
		types.String,
		`"simple string"`,
	},
	{
		"string",
		`<html> \d\s`,
		types.String,
		`"<html> \d\s"`,
	},
	{
		"string",
		`{"src": <html> \d\s, "error": false, "code": 123}`,
		types.String,
		`"{"src": <html> \d\s, "error": false, "code": 123}"`,
	},
	{
		"json",
		`{"src": <html> \d\s, "error": false, "code": 123}`,
		typesExt.TMap,
		`{"src": <html> \d\s, "error": false, "code": 123}`,
	},
}

func TestWriteElemValue(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := writeTestValue(tt)
			assert.Equal(t, tt.want, string(ctx.Response.Body()))
		})
	}
}

func writeTestValue(tt struct {
	name string
	src  string
	col  types.BasicKind
	want string
}) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	var col dbEngine.Column
	switch tt.col {
	case types.String:
		col = dbEngine.NewStringColumn(tt.name, "comment", false, 0)
	case typesExt.TMap:
		col = psql.NewColumn(nil, tt.name, "json", nil, true,
			"", "comment", "jsonb", 0,
			false, false)
	}
	WriteElemValue(ctx, []byte(tt.src), col)
	return ctx
}

func BenchmarkWriteElemValue(b *testing.B) {
	b.ReportAllocs()
	for _, tt := range tests {
		b.ReportAllocs()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			b.Run(tt.name, func(b *testing.B) {

				ctx := writeTestValue(tt)
				b.Logf("%s", ctx.Response.Body())
			})
		}
		b.ResetTimer()
		b.ReportAllocs()
	}
	b.ReportAllocs()

}
