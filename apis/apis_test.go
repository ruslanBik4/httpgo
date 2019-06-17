package apis

import (
	"github.com/json-iterator/go"
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

	apis.addRoute("test", &ApiRoute{Desc: "test route"})
	resp, err := apis.renderApis(NewCtxApis(0))
	assert.Nil(t, err)
	t.Log(resp)

	b, err := jsoniter.Marshal(resp)
	assert.Nil(t, err)
	t.Log(string(b))

	var v interface{}
	err = jsoniter.Unmarshal(b, &v)
	assert.Nil(t, err)

	res, ok := v.(map[string]interface{})
	assert.True(t, ok, "%T", v)

	for key, val := range res {
		t.Logf(`"%s"=%#v`, key, val)

	}
}
