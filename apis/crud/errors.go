// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"regexp"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
)

var (
	regDuplicated = regexp.MustCompile(`duplicate key value violates unique constraint "(\w*)"`)
	regKeyWrong   = regexp.MustCompile(`[Kk]ey\s+(?:[(\w\s]+)?\((\w+)(?:,[^=]+)?\)+=\(([^)]+)\)([^.]+)`)
)

func CreateErrResult(err error) (interface{}, error) {
	msg := err.Error()
	e, ok := errors.Cause(err).(*pgconn.PgError)
	if ok {
		msg = e.Detail
		logs.DebugLog(e)
		logs.StatusLog(e, msg)
	}

	// Key (id)=(3) already exists. duplicate key value violates unique constraint "candidates_name_uindex"
	// duplicate key value violates unique constraint "candidates_mobile_uindex"
	// Key (digest(blob, 'sha1'::text))=(\x34d3fb7ceb19bf448d89ab76e7b1e16260c1d8b0) already exists.
	// key (phone)=(+380) already exists.

	if s := regKeyWrong.FindStringSubmatch(msg); len(s) > 0 {
		return map[string]string{
			s[1]: "`" + s[2] + "`" + s[3],
		}, apis.ErrWrongParamsList
	} else {
		logs.StatusLog(regKeyWrong.String(), s)
	}
	if s := regDuplicated.FindStringSubmatch(msg); len(s) > 0 {
		logs.DebugLog("%#v %[1]T", errors.Cause(err))
		return map[string]string{
			s[1]: "duplicate key value violates unique constraint",
		}, apis.ErrWrongParamsList
	}

	return nil, err
}

func RenderCreatedResult(ctx *fasthttp.RequestCtx, id int64, msg string, colSel []string, url string) (interface{}, error) {
	msg = "Success saving: " + strings.Join(colSel, ", ") + " values:\n" + msg

	ctx.SetStatusCode(fasthttp.StatusCreated)
	g, ok := ctx.UserValue(ParamsGetFormActions.Name).(bool)
	if ok && g {
		url += "/form?html"

		lang := ctx.UserValue("lang")
		if l, ok := lang.(string); ok {
			url += "&lang=" + l
		}

		return insertResult{
			FormActions: []FormActions{
				{
					Typ: "redirect",
					Url: url,
				},
			},
			Id:  id,
			Msg: msg,
		}, nil
	}

	return id, nil
}
