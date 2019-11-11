package apis

import (
	"github.com/json-iterator/go"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var (
	ctx  = &fasthttp.RequestCtx{}
	apis = &Apis{
		routes: NewAPIRoutes(),
	}
)

func TestApis_AddRoute(t *testing.T) {

	err := apis.addRoute("test", &ApiRoute{})
	assert.Nil(t, err)
	err = apis.addRoute("test", &ApiRoute{})
	assert.NotNil(t, err)
}

func TestRenderApis(t *testing.T) {

	err := apis.addRoute(
		"moreParams",
		&ApiRoute{
			Desc:      "test route",
			Method:    POST,
			Multipart: true,
			NeedAuth:  true,
			Params: []InParam{
				{
					Name: "globalTags",
					Desc: "data of dashboard -> filter 'Global Tags'",
					Req:  false,
					Type: NewSliceTypeInParam(types.Int32),
				},
				{
					Name:     "group",
					Desc:     "type grouping data of ohlc (month, week, day)",
					Req:      true,
					Type:     NewTypeInParam(types.String),
					DefValue: "day",
				},
				{
					Name: "account",
					Desc: "account numbers to filter data",
					Req:  true,
					Type: TypeInParam{
						BasicKind: types.Bool,
						// types.Error{Msg:"test err", Soft:true},
						isSlice: false,
					},
					DefValue: testValue,
				},
			},
			Resp: struct {
				Hours map[string]float64
			}{
				map[string]float64{"after 16:00": 15.2,
					"13:30 - 15:30": 1570.86,
					"9:30 - 9:50":   1672.54,
				},
			},
		},
	)
	assert.Nil(t, err)

	resp, err := apis.renderApis(ctx)
	assert.Nil(t, err)
	t.Logf(`%#v`, resp)

	b, err := jsoniter.Marshal(resp)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	t.Log(string(b))

	var v interface{}
	err = jsoniter.Unmarshal(b, &v)
	assert.Nil(t, err)

	res, ok := v.(map[string]interface{})
	assert.True(t, ok, "%T", v)

	for key, val := range res {
		t.Logf(`'%s': %#v`, key, val)
	}
}

func testValue(ctx *fasthttp.RequestCtx) interface{} {
	return ctx.Method()
}
