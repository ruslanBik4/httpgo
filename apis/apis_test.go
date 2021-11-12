package apis

import (
	"bufio"
	"go/types"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/json-iterator/go"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var (
	ctx  = &fasthttp.RequestCtx{}
	apis = &Apis{
		routes: NewMapRoutes(),
	}
)

func TestApis_AddRoute(t *testing.T) {

	err := apis.addRoute("test", &ApiRoute{})
	assert.Nil(t, err)
	err = apis.addRoute("test", &ApiRoute{})
	assert.NotNil(t, err)

}

func TestRenderApis(t *testing.T) {

	TestCheckAndRun(t)
	const testPath = "moreParams"
	err := apis.addRoute(
		testPath,
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

	ctx.Request.SetRequestURI(testPath)
	ctx.SetUserValue("json", true)
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

type apiDTO struct {
	i interface{}
}

func (a *apiDTO) GetValue() interface{} {
	return a.i
}

func (a *apiDTO) NewValue() interface{} {
	switch v := (a.i).(type) {
	default:
		var r interface{}
		r = v
		return r
	}
}

func TestNewStructInParam(t *testing.T) {
	st := struct {
		i int
		s string
	}{1, "test"}

	a := apiDTO{st}
	var newSt interface{}
	newSt = a.NewValue()
	tt := newSt.(struct {
		i int
		s string
	})
	tt.i = 2
	assert.Equal(t, st, newSt)
}

func TestOnboarding(t *testing.T) {

	fPort := ":8989"
	listener, err := net.Listen("tcp", fPort)
	if err != nil {
		t.Fatal(err)
	}

	time.AfterFunc(time.Minute*3,
		func() {
			err := listener.Close()
			if err != nil {
				t.Fatal(err)
			}
		})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	t.Log("start")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Fatal(err)
			}

			reader := bufio.NewReader(conn)

			str, _ := reader.ReadString('\n')
			t.Log(str)

			const head = `HTTP/1.1 200 Success 
Content-Type: text/html; \n Retry-After: 60
<meta http-equiv="Refresh" content="15" />


`
			w := bufio.NewWriter(conn)
			_, err = w.WriteString(head + "<html>hello</html>")
			if err != nil {
				t.Fatal(err)
			}

			_ = w.Flush()
			_ = conn.Close()
			break
			// wg.Done()
		}
	}()
	go func() {
		c, err := net.Dial("tcp", "127.0.0.1"+fPort)
		if err != nil {
			t.Fatal(err)
		}

		b := []byte("hello\n")
		c.Write(b)
		n, err := c.Read(b)
		if err != nil {
			t.Error(err)
		} else {
			t.Logf("%s (%d)", b, n)
		}
		c.Close()
		wg.Done()
	}()

	wg.Wait()
}
