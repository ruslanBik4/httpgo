// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apis

import (
	"errors"
	"github.com/valyala/fasthttp"
)

// errors declaration
var (
	errNotFoundPage     = errors.New("path not found")
	errMethodNotAllowed = errors.New("method %s not allowed, try %s")

	ErrUnAuthorized       = errors.New("user is UnAuthorized")
	ErrRouteForbidden     = errors.New("not allow permission")
	errRouteOnlyLocal     = errors.New("not allow permission for remote domain")
	ErrPathAlreadyExists  = errors.New("this path already exists")
	ErrWrongParamsList    = errors.New("wrong params list: %+v")
	errIncompatibleParams = errors.New("found incompatible params- %+v")
)

type ErrorResp struct {
	FormErrors map[string]string `json:"formErrors"`
}

func NewErrorResp(formErrors map[string]string) *ErrorResp {
	return &ErrorResp{FormErrors: formErrors}
}

func NewErrorRespBadDTO() *ErrorResp {
	return NewErrorResp(map[string]string{"DTO": "wrong struct"})
}

func WriteCustomErrorResponse(ctx *fasthttp.RequestCtx, code int, err error, args map[string]string) (any, error) {
	if args == nil {
		args = make(map[string]string)
	}
	if err != nil {
		args["error"] = err.Error()
	}

	ret := NewErrorResp(args)
	ctx.SetStatusCode(code)

	return ret, nil
}
