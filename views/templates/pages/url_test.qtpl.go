/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Code generated by qtc from "url_test.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

// All the text outside function templates is treated as comments,
// i.e. it is just ignored by quicktemplate compiler (`qtc`). It is for humans.

//line views/templates/pages/url_test.qtpl:3
package pages

//line views/templates/pages/url_test.qtpl:3
import (
	"go/types"
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"

	"github.com/ruslanBik4/httpgo/views/templates/css"
)

//line views/templates/pages/url_test.qtpl:9

//line views/templates/pages/url_test.qtpl:9
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line views/templates/pages/url_test.qtpl:10
type ParamUrlTestPage struct {
	Basic   *types.Basic
	Name    string
	IsSlice bool
	Value   interface{}
	Req     bool
	Type    string
	Comment string
}

type URLTestPage struct {
	Host       string
	Method     string
	Multipart  bool
	Path       string
	Language   string
	Charset    string
	LinkStyles []string
	MetaTags   []string
	Params     []ParamUrlTestPage
	Resp       string
}

//line views/templates/pages/url_test.qtpl:35
func (p *ParamUrlTestPage) StreamTypeInput(qw422016 *qt422016.Writer) {
//line views/templates/pages/url_test.qtpl:36
	switch p.Basic.Info() {
//line views/templates/pages/url_test.qtpl:37
	case types.IsInteger, types.IsFloat, types.IsComplex:
//line views/templates/pages/url_test.qtpl:37
		qw422016.N().S(`number`)
//line views/templates/pages/url_test.qtpl:39
	case types.IsString:
//line views/templates/pages/url_test.qtpl:39
		qw422016.N().S(`text`)
//line views/templates/pages/url_test.qtpl:41
	case types.IsBoolean:
//line views/templates/pages/url_test.qtpl:41
		qw422016.N().S(`checkbox`)
//line views/templates/pages/url_test.qtpl:43
	default:
//line views/templates/pages/url_test.qtpl:44
		qw422016.E().V(p.Basic.Info())
//line views/templates/pages/url_test.qtpl:45
	}
//line views/templates/pages/url_test.qtpl:46
}

//line views/templates/pages/url_test.qtpl:46
func (p *ParamUrlTestPage) WriteTypeInput(qq422016 qtio422016.Writer) {
//line views/templates/pages/url_test.qtpl:46
	qw422016 := qt422016.AcquireWriter(qq422016)
//line views/templates/pages/url_test.qtpl:46
	p.StreamTypeInput(qw422016)
//line views/templates/pages/url_test.qtpl:46
	qt422016.ReleaseWriter(qw422016)
//line views/templates/pages/url_test.qtpl:46
}

//line views/templates/pages/url_test.qtpl:46
func (p *ParamUrlTestPage) TypeInput() string {
//line views/templates/pages/url_test.qtpl:46
	qb422016 := qt422016.AcquireByteBuffer()
//line views/templates/pages/url_test.qtpl:46
	p.WriteTypeInput(qb422016)
//line views/templates/pages/url_test.qtpl:46
	qs422016 := string(qb422016.B)
//line views/templates/pages/url_test.qtpl:46
	qt422016.ReleaseByteBuffer(qb422016)
//line views/templates/pages/url_test.qtpl:46
	return qs422016
//line views/templates/pages/url_test.qtpl:46
}

//line views/templates/pages/url_test.qtpl:48
func (u *URLTestPage) StreamURL(qw422016 *qt422016.Writer) {
//line views/templates/pages/url_test.qtpl:49
	qw422016.E().S(u.Host)
//line views/templates/pages/url_test.qtpl:49
	qw422016.E().S(u.Path)
//line views/templates/pages/url_test.qtpl:50
}

//line views/templates/pages/url_test.qtpl:50
func (u *URLTestPage) WriteURL(qq422016 qtio422016.Writer) {
//line views/templates/pages/url_test.qtpl:50
	qw422016 := qt422016.AcquireWriter(qq422016)
//line views/templates/pages/url_test.qtpl:50
	u.StreamURL(qw422016)
//line views/templates/pages/url_test.qtpl:50
	qt422016.ReleaseWriter(qw422016)
//line views/templates/pages/url_test.qtpl:50
}

//line views/templates/pages/url_test.qtpl:50
func (u *URLTestPage) URL() string {
//line views/templates/pages/url_test.qtpl:50
	qb422016 := qt422016.AcquireByteBuffer()
//line views/templates/pages/url_test.qtpl:50
	u.WriteURL(qb422016)
//line views/templates/pages/url_test.qtpl:50
	qs422016 := string(qb422016.B)
//line views/templates/pages/url_test.qtpl:50
	qt422016.ReleaseByteBuffer(qb422016)
//line views/templates/pages/url_test.qtpl:50
	return qs422016
//line views/templates/pages/url_test.qtpl:50
}

//line views/templates/pages/url_test.qtpl:52
func (u *URLTestPage) StreamEncType(qw422016 *qt422016.Writer) {
//line views/templates/pages/url_test.qtpl:53
	if u.Multipart {
//line views/templates/pages/url_test.qtpl:53
		qw422016.N().S(`multipart/form-data`)
//line views/templates/pages/url_test.qtpl:55
	} else {
//line views/templates/pages/url_test.qtpl:56
	}
//line views/templates/pages/url_test.qtpl:57
}

//line views/templates/pages/url_test.qtpl:57
func (u *URLTestPage) WriteEncType(qq422016 qtio422016.Writer) {
//line views/templates/pages/url_test.qtpl:57
	qw422016 := qt422016.AcquireWriter(qq422016)
//line views/templates/pages/url_test.qtpl:57
	u.StreamEncType(qw422016)
//line views/templates/pages/url_test.qtpl:57
	qt422016.ReleaseWriter(qw422016)
//line views/templates/pages/url_test.qtpl:57
}

//line views/templates/pages/url_test.qtpl:57
func (u *URLTestPage) EncType() string {
//line views/templates/pages/url_test.qtpl:57
	qb422016 := qt422016.AcquireByteBuffer()
//line views/templates/pages/url_test.qtpl:57
	u.WriteEncType(qb422016)
//line views/templates/pages/url_test.qtpl:57
	qs422016 := string(qb422016.B)
//line views/templates/pages/url_test.qtpl:57
	qt422016.ReleaseByteBuffer(qb422016)
//line views/templates/pages/url_test.qtpl:57
	return qs422016
//line views/templates/pages/url_test.qtpl:57
}

//line views/templates/pages/url_test.qtpl:63
func (u *URLTestPage) StreamShowURlTestPage(qw422016 *qt422016.Writer) {
//line views/templates/pages/url_test.qtpl:63
	qw422016.N().S(`
<body>
`)
//line views/templates/pages/url_test.qtpl:65
	css.StreamPutStyles(qw422016)
//line views/templates/pages/url_test.qtpl:65
	qw422016.N().S(`
<div class="content-wrap">
<div id="container-fluid">
    <div class="row-fluid">
        <div class="sidebar-section">
            <div id="catalog_pane"  class="well sidebar-nav">
`)
//line views/templates/pages/url_test.qtpl:72
	qw422016.N().S(`<form action="`)
//line views/templates/pages/url_test.qtpl:73
	u.StreamURL(qw422016)
//line views/templates/pages/url_test.qtpl:73
	qw422016.N().S(`" method="`)
//line views/templates/pages/url_test.qtpl:73
	qw422016.E().S(u.Method)
//line views/templates/pages/url_test.qtpl:73
	qw422016.N().S(`" enctype="`)
//line views/templates/pages/url_test.qtpl:73
	u.StreamEncType(qw422016)
//line views/templates/pages/url_test.qtpl:73
	qw422016.N().S(`" target="_blank">`)
//line views/templates/pages/url_test.qtpl:74
	for _, param := range u.Params {
//line views/templates/pages/url_test.qtpl:74
		qw422016.N().S(`<p>`)
//line views/templates/pages/url_test.qtpl:75
		qw422016.E().S(param.Comment)
//line views/templates/pages/url_test.qtpl:75
		qw422016.N().S(`</p><input name='`)
//line views/templates/pages/url_test.qtpl:76
		qw422016.E().S(param.Name)
//line views/templates/pages/url_test.qtpl:76
		if param.IsSlice {
//line views/templates/pages/url_test.qtpl:76
			qw422016.N().S(`[]`)
//line views/templates/pages/url_test.qtpl:76
		}
//line views/templates/pages/url_test.qtpl:76
		qw422016.N().S(`'`)
//line views/templates/pages/url_test.qtpl:77
		if param.Value != nil {
//line views/templates/pages/url_test.qtpl:77
			qw422016.N().S(`value='`)
//line views/templates/pages/url_test.qtpl:78
			qw422016.E().V(param.Value)
//line views/templates/pages/url_test.qtpl:78
			qw422016.N().S(`'`)
//line views/templates/pages/url_test.qtpl:79
		}
//line views/templates/pages/url_test.qtpl:79
		qw422016.N().S(`type='`)
//line views/templates/pages/url_test.qtpl:80
		param.StreamTypeInput(qw422016)
//line views/templates/pages/url_test.qtpl:80
		qw422016.N().S(`'`)
//line views/templates/pages/url_test.qtpl:81
		if param.Req {
//line views/templates/pages/url_test.qtpl:81
			qw422016.N().S(`required`)
//line views/templates/pages/url_test.qtpl:83
		}
//line views/templates/pages/url_test.qtpl:83
		qw422016.N().S(`/>`)
//line views/templates/pages/url_test.qtpl:85
	}
//line views/templates/pages/url_test.qtpl:85
	qw422016.N().S(`<button> Send real URL "`)
//line views/templates/pages/url_test.qtpl:86
	u.StreamURL(qw422016)
//line views/templates/pages/url_test.qtpl:86
	qw422016.N().S(`"</button></form>`)
//line views/templates/pages/url_test.qtpl:88
	qw422016.N().S(` `)
//line views/templates/pages/url_test.qtpl:89
	qw422016.N().S(`

            </div>
        </div>
        <div class="content-sep"> </div>
         <div class="content-section">
         <div id="content" rel="/admin/">
           <span>  Phyton code </span>
           <pre >
            import requests

            WEBKIT = '------WebKitFormBoundary7MA4YWxkTrZu0gW'
            HEADERS = {
                    'Authorization': 'Bearer 2571020569',
                    'content-type': "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
                }


            def post_multipart(url: str, data: dict):
                return requests.post(
                    url, files={
                        k: (None, v) for k, v in data.items()
                    }, headers=headers
                )

            def post_json(url: str, data: dict):
                return requests.post(
                    url, data=data,
                    headers={'Content-Type': 'application/json', **headers}
                )

            def ta_request(path, params):
                payload = ''
                for p in params:
                    payload += '%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n%s\r\n' % (WEBKIT, p[0], p[1])

                payload += '%s--' % WEBKIT

                response = requests.request("POST", path, data=payload, headers=HEADERS).json()


            if __name__ == '__main__':
                ta_request("`)
//line views/templates/pages/url_test.qtpl:131
	u.StreamURL(qw422016)
//line views/templates/pages/url_test.qtpl:131
	qw422016.N().S(`", [
                 `)
//line views/templates/pages/url_test.qtpl:132
	for _, param := range u.Params {
//line views/templates/pages/url_test.qtpl:132
		qw422016.N().S(`
                    `)
//line views/templates/pages/url_test.qtpl:133
		if param.Req {
//line views/templates/pages/url_test.qtpl:133
			qw422016.N().S(`
                        required
                    `)
//line views/templates/pages/url_test.qtpl:135
		}
//line views/templates/pages/url_test.qtpl:135
		qw422016.N().S(`
                    ('`)
//line views/templates/pages/url_test.qtpl:136
		qw422016.E().S(param.Name)
//line views/templates/pages/url_test.qtpl:136
		qw422016.N().S(`','`)
//line views/templates/pages/url_test.qtpl:136
		qw422016.E().V(param.Value)
//line views/templates/pages/url_test.qtpl:136
		qw422016.N().S(`'), // type: `)
//line views/templates/pages/url_test.qtpl:136
		qw422016.E().S(param.Type)
//line views/templates/pages/url_test.qtpl:136
		qw422016.N().S(`,  `)
//line views/templates/pages/url_test.qtpl:136
		qw422016.E().S(param.Comment)
//line views/templates/pages/url_test.qtpl:136
		qw422016.N().S(`
                 `)
//line views/templates/pages/url_test.qtpl:137
	}
//line views/templates/pages/url_test.qtpl:137
	qw422016.N().S(`
                ])
                </pre>

           <span>  JS code </span>
           <pre>
           request({
             'multipart/form-data',
             'POST',
             processData = true,
             "`)
//line views/templates/pages/url_test.qtpl:147
	u.StreamURL(qw422016)
//line views/templates/pages/url_test.qtpl:147
	qw422016.N().S(`",
             {
                   `)
//line views/templates/pages/url_test.qtpl:149
	for _, param := range u.Params {
//line views/templates/pages/url_test.qtpl:149
		qw422016.N().S(`
                      "`)
//line views/templates/pages/url_test.qtpl:150
		qw422016.E().S(param.Name)
//line views/templates/pages/url_test.qtpl:150
		qw422016.N().S(`": `)
//line views/templates/pages/url_test.qtpl:150
		qw422016.E().V(param.Value)
//line views/templates/pages/url_test.qtpl:150
		qw422016.N().S(`,// type: `)
//line views/templates/pages/url_test.qtpl:150
		qw422016.E().S(param.Type)
//line views/templates/pages/url_test.qtpl:150
		qw422016.N().S(`,  `)
//line views/templates/pages/url_test.qtpl:150
		qw422016.E().S(param.Comment)
//line views/templates/pages/url_test.qtpl:150
		qw422016.N().S(`
                   `)
//line views/templates/pages/url_test.qtpl:151
	}
//line views/templates/pages/url_test.qtpl:151
	qw422016.N().S(`
              },
             beforeSend = null,
             complete = null,
             error = null,
             success = null,
             withoutAPI = false,
             onprogress = null
           })
           </pre>
           `)
//line views/templates/pages/url_test.qtpl:161
	if u.Resp > "" {
//line views/templates/pages/url_test.qtpl:161
		qw422016.N().S(`
            <span>  Response example </span>
          <div>
             `)
//line views/templates/pages/url_test.qtpl:164
		qw422016.E().S(u.Resp)
//line views/templates/pages/url_test.qtpl:164
		qw422016.N().S(`
           </div>
           `)
//line views/templates/pages/url_test.qtpl:166
	}
//line views/templates/pages/url_test.qtpl:166
	qw422016.N().S(`
            </div>
        </div>
        </div>
        </div>

</div>

</body>
`)
//line views/templates/pages/url_test.qtpl:175
}

//line views/templates/pages/url_test.qtpl:175
func (u *URLTestPage) WriteShowURlTestPage(qq422016 qtio422016.Writer) {
//line views/templates/pages/url_test.qtpl:175
	qw422016 := qt422016.AcquireWriter(qq422016)
//line views/templates/pages/url_test.qtpl:175
	u.StreamShowURlTestPage(qw422016)
//line views/templates/pages/url_test.qtpl:175
	qt422016.ReleaseWriter(qw422016)
//line views/templates/pages/url_test.qtpl:175
}

//line views/templates/pages/url_test.qtpl:175
func (u *URLTestPage) ShowURlTestPage() string {
//line views/templates/pages/url_test.qtpl:175
	qb422016 := qt422016.AcquireByteBuffer()
//line views/templates/pages/url_test.qtpl:175
	u.WriteShowURlTestPage(qb422016)
//line views/templates/pages/url_test.qtpl:175
	qs422016 := string(qb422016.B)
//line views/templates/pages/url_test.qtpl:175
	qt422016.ReleaseByteBuffer(qb422016)
//line views/templates/pages/url_test.qtpl:175
	return qs422016
//line views/templates/pages/url_test.qtpl:175
}
