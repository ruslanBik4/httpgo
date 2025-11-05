/*
 * Copyright (c) 2022-2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

var (
	regDuplicated    = regexp.MustCompile(`duplicate key value violates unique constraint "(\w*)"`)
	regSyntaxWrong   = regexp.MustCompile(`invalid\sinput\ssyntax\sfor\stype\s+([\w\s)(]+)\:\s+"([^"]*)"`)
	regCheckViolates = regexp.MustCompile(`new row for relation "(\w+)(?:,[^=]+)?" violates check constraint "(\w*)"`)
	regKeyWrong      = regexp.MustCompile(`[Kk]ey\s+(?:[(\w\s]+)?\((\w+)(?:,[^=]+)?\)+=\(([^)]+)\)([^.]+)`)
	regKeyTypeWrong  = regexp.MustCompile(`column\s+"(\w+)(?:,[^=]+)?"\s+is\s+of\s+type\s+([(\w\s)]+) but expression is of type ([(\w\s)]+)$`)
)

func CreateErrResult(err error) (any, error) {
	if err == nil || errors.Unwrap(err) == sql.ErrNoRows || errors.Cause(err) == sql.ErrNoRows || errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	msg := err.Error()
	e, ok := errors.Cause(err).(*pgconn.PgError)
	if ok {
		if e.Detail != "" {
			msg = e.Detail
		} else {
			msg = e.Message
		}
		logs.DebugLog("%#v", e)
	}

	logs.StatusLog(msg)
	// Key (id)=(3) already exists. duplicate key value violates unique constraint "candidates_name_uindex"
	// duplicate key value violates unique constraint "candidates_mobile_uindex"
	// Key (digest(blob, 'sha1'::text))=(\x34d3fb7ceb19bf448d89ab76e7b1e16260c1d8b0) already exists.
	// key (phone)=(+380) already exists.

	if s := regKeyWrong.FindStringSubmatch(msg); len(s) > 0 {
		return apis.NewErrorResp(map[string]string{
			s[1]: "`" + s[2] + "`" + s[3],
		}), apis.ErrWrongParamsList
	}

	// column "risk" is of type numeric but expression is of type text
	if s := regKeyTypeWrong.FindStringSubmatch(msg); len(s) > 0 {
		return apis.NewErrorResp(map[string]string{
			s[1]: s[3] + " instead of" + s[2],
		}), apis.ErrWrongParamsList
	}

	// invalid input syntax for type numeric: ""
	if s := regSyntaxWrong.FindStringSubmatch(msg); len(s) > 0 {
		return apis.NewErrorResp(map[string]string{
			s[1]: s[0],
		}), apis.ErrWrongParamsList
	}

	if s := regDuplicated.FindStringSubmatch(msg); len(s) > 0 {
		logs.DebugLog("%#v %[1]T", errors.Cause(err))
		return apis.NewErrorResp(map[string]string{
			s[1]: "duplicate key value violates unique constraint",
		}), apis.ErrWrongParamsList
	}

	//new row for relation "trading_plans" violates check constraint "risk_no_zero"
	if e.ConstraintName != "" {
		return apis.NewErrorResp(map[string]string{
			e.TableName: fmt.Sprintf(`violates check constraint "%s"`, e.ConstraintName),
		}), apis.ErrWrongParamsList
	}

	logs.ErrorLog(err)
	return nil, err
}
