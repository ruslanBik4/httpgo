/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

// Code generated by qtc from "enpdointHeader.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line enpdointHeader.qtpl:3
package tpl

//line enpdointHeader.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line enpdointHeader.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line enpdointHeader.qtpl:4
type enpdointHeader struct {
}

func NewenpdointHeader() *enpdointHeader {
	return &enpdointHeader{}
}

//line enpdointHeader.qtpl:13
func StreamHead(qw422016 *qt422016.Writer, hasTypeExt bool, sdkLib ...string) {
//line enpdointHeader.qtpl:13
	qw422016.N().S(`// Code generated by httpgo-gen-go. DO NOT EDIT.
// versions:
// 	httpgo v1.2.*
// source: %s %s
package db

import (
	"fmt"
	"bytes"
	"strings"
	"go/types"
`)
//line enpdointHeader.qtpl:24
	for _, lib := range sdkLib {
//line enpdointHeader.qtpl:24
		qw422016.N().S(`"`)
//line enpdointHeader.qtpl:24
		qw422016.E().S(lib)
//line enpdointHeader.qtpl:24
		qw422016.N().S(`"
`)
//line enpdointHeader.qtpl:25
	}
//line enpdointHeader.qtpl:25
	qw422016.N().S(`	"github.com/jackc/pgx/v4"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"github.com/pkg/errors"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/logs"
	"github.com/ruslanBik4/dbEngine/dbEngine"
`)
//line enpdointHeader.qtpl:36
	if hasTypeExt {
//line enpdointHeader.qtpl:36
		qw422016.N().S(`"github.com/ruslanBik4/httpgo/typesExt"`)
//line enpdointHeader.qtpl:36
	}
//line enpdointHeader.qtpl:36
	qw422016.N().S(`
	"github.com/ruslanBik4/httpgo/views/templates/forms"
)


`)
//line enpdointHeader.qtpl:41
}

//line enpdointHeader.qtpl:41
func WriteHead(qq422016 qtio422016.Writer, hasTypeExt bool, sdkLib ...string) {
//line enpdointHeader.qtpl:41
	qw422016 := qt422016.AcquireWriter(qq422016)
//line enpdointHeader.qtpl:41
	StreamHead(qw422016, hasTypeExt, sdkLib...)
//line enpdointHeader.qtpl:41
	qt422016.ReleaseWriter(qw422016)
//line enpdointHeader.qtpl:41
}

//line enpdointHeader.qtpl:41
func Head(hasTypeExt bool, sdkLib ...string) string {
//line enpdointHeader.qtpl:41
	qb422016 := qt422016.AcquireByteBuffer()
//line enpdointHeader.qtpl:41
	WriteHead(qb422016, hasTypeExt, sdkLib...)
//line enpdointHeader.qtpl:41
	qs422016 := string(qb422016.B)
//line enpdointHeader.qtpl:41
	qt422016.ReleaseByteBuffer(qb422016)
//line enpdointHeader.qtpl:41
	return qs422016
//line enpdointHeader.qtpl:41
}