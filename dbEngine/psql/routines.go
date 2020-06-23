// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"sync"

	"github.com/jackc/pgproto3/v2"
	"golang.org/x/net/context"
)

type PgxRoutineParams struct {
	Fnc                      *Routine `json:"-"`
	Name                     string
	DataType                 string
	DataName                 string
	CHARACTER_SET_NAME       string
	COLUMN_COMMENT           string
	CHARACTER_MAXIMUM_LENGTH int32
	ParameterDefault         string
	Position                 int32
}

type Routine struct {
	conn    *Conn
	name    string
	ID      int
	Comment string
	Fields  []*PgxRoutineParams
	params  []*PgxRoutineParams
	Overlay *Routine
	Type    string
	lock    sync.RWMutex
}

func (r *Routine) Name() string {
	return r.name
}

func (r *Routine) Select(ctx context.Context, args ...interface{}) error {
	panic("implement me")
}

func (r *Routine) Call(context.Context) {
	panic("implement me")
}

func (r *Routine) Params() {
	panic("implement me")
}

// GetParams получение значений полей для форматирования данных
// получение значений полей для таблицы
func (r *Routine) GetParams(ctx context.Context) error {

	return r.conn.SelectAndRunEach(ctx, func(values []interface{}, columns []pgproto3.FieldDescription) error {

		if values[0] == nil {
			return nil
		}

		row := &PgxRoutineParams{
			Fnc:                      r,
			Name:                     values[0].(string),
			DataType:                 values[1].(string),
			DataName:                 values[2].(string),
			CHARACTER_SET_NAME:       values[3].(string),
			CHARACTER_MAXIMUM_LENGTH: values[4].(int32),
			ParameterDefault:         values[5].(string),
			Position:                 values[6].(int32),
		}

		if values[7].(string) == "IN" {
			r.params = append(r.params, row)
		} else {
			r.Fields = append(r.Fields, row)
		}

		return nil
	}, sqlGetFuncParams+" ORDER BY ordinal_position", r.name)
}
