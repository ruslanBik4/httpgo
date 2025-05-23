/*
 * Copyright (c) 2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Code generated by qtc from "index.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

// All the text outside function templates is treated as comments,
// i.e. it is just ignored by quicktemplate compiler (`qtc`). It is for humans.

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:3
package pages

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:3
import (
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/valyala/quicktemplate"
	"io"
)

// content of Index page

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:10
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:10
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:11
type IndexPageBody struct {
	Name         []byte
	Pass         []byte
	Content      string
	ContentWrite func(w io.Writer)
	Catalog      layouts.Menu
	TopMenu      layouts.Menu
	FooterMenu   layouts.Menu
	HeadHTML     *layouts.HeadHTMLPage
	OwnerMenu    *layouts.MenuOwnerBody
	Title        string
	Route        string
	Attr         string
	AfterAuthURL string
	ChangeTheme  string
	SearchPanel  *layouts.SearchPanel
}

func (body *IndexPageBody) StreamContentWrite(w *quicktemplate.Writer) {
	body.ContentWrite(w.W())
}

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:34
func (body *IndexPageBody) StreamIndexHTML(qw422016 *qt422016.Writer) {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:34
	qw422016.N().S(`
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:35
	body.HeadHTML.StreamHeadHTML(qw422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:35
	qw422016.N().S(`
<body `)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:36
	qw422016.E().S(body.Attr)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:36
	qw422016.N().S(`>
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:37
	layouts.StreamHeaderHTML(qw422016, body.TopMenu)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:37
	qw422016.N().S(`
<breadcrumbs separator="›" aria-label="breadcrumb" aria-label="breadcrumb">
  <ol class="breadcrumb">
      <li class="breadcrumb-item"><a href="/">Home</a></li>
  </ol>
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:42
	if body.SearchPanel != nil {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:42
		body.SearchPanel.StreamRender(qw422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:42
	}
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:42
	qw422016.N().S(`</breadcrumbs>
<main class="content-wrap">
	<div id="container-fluid">
	        <aside class="sidebar-section">
	            <div id="catalog_pane"  class="well sidebar-nav">
	                <div class="sidebar"></div>
	                `)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:49
	body.Catalog.StreamRenderMenu(qw422016, "left-mnu-list", "left-mnu-item")
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:49
	qw422016.N().S(`
	            </div>
	        </aside>
	        <div role="slider" aria-label="Draggable pane splitter" aria-valuemin="256" aria-valuemax="372" aria-valuenow="256" aria-valuetext="Pane width 256 pixels" tabindex="0">
	        </div>
	        <div class="content-section">
		        <div id="content" `)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
	if body.Route > "" {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
		qw422016.N().S(`hx-get="`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
		qw422016.N().S(body.Route)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
		qw422016.N().S(`" hx-trigger="load once" rel='htmx'`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
	}
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:55
	qw422016.N().S(`>
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:56
	if body.ContentWrite != nil {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:56
		body.StreamContentWrite(qw422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:56
		qw422016.N().S(`
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:57
	} else {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:57
		qw422016.E().S(body.Content)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:57
		qw422016.N().S(`
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:58
	}
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:58
	qw422016.N().S(`</div>
	        </div>
	</div>
</main>
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:62
	layouts.StreamFooterHTML(qw422016, body.FooterMenu)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:62
	qw422016.N().S(`
</body>
`)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
}

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
func (body *IndexPageBody) WriteIndexHTML(qq422016 qtio422016.Writer) {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	qw422016 := qt422016.AcquireWriter(qq422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	body.StreamIndexHTML(qw422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	qt422016.ReleaseWriter(qw422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
}

//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
func (body *IndexPageBody) IndexHTML() string {
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	qb422016 := qt422016.AcquireByteBuffer()
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	body.WriteIndexHTML(qb422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	qs422016 := string(qb422016.B)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	qt422016.ReleaseByteBuffer(qb422016)
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
	return qs422016
//line /Users/ruslan_bik/GolandProjects/httpgo/views/templates/pages/index.qtpl:64
}
