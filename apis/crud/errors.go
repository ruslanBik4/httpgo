/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"database/sql"
	"regexp"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

var (
	regDuplicated = regexp.MustCompile(`duplicate key value violates unique constraint "(\w*)"`)
	regKeyWrong   = regexp.MustCompile(`[Kk]ey\s+(?:[(\w\s]+)?\((\w+)(?:,[^=]+)?\)+=\(([^)]+)\)([^.]+)`)
)

func CreateErrResult(err error) (any, error) {
	if err == nil || errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	msg := err.Error()
	e, ok := errors.Cause(err).(*pgconn.PgError)
	if ok {
		msg = e.Detail
		logs.DebugLog(e, msg)
	}

	// Key (id)=(3) already exists. duplicate key value violates unique constraint "candidates_name_uindex"
	// duplicate key value violates unique constraint "candidates_mobile_uindex"
	// Key (digest(blob, 'sha1'::text))=(\x34d3fb7ceb19bf448d89ab76e7b1e16260c1d8b0) already exists.
	// key (phone)=(+380) already exists.

	if s := regKeyWrong.FindStringSubmatch(msg); len(s) > 0 {
		return apis.NewErrorResp(map[string]string{
			s[1]: "`" + s[2] + "`" + s[3],
		}), apis.ErrWrongParamsList
	}
	if s := regDuplicated.FindStringSubmatch(msg); len(s) > 0 {
		logs.DebugLog("%#v %[1]T", errors.Cause(err))
		return apis.NewErrorResp(map[string]string{
			s[1]: "duplicate key value violates unique constraint",
		}), apis.ErrWrongParamsList
	}

	logs.ErrorLog(err)
	return nil, err
}
