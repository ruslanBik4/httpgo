// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

import (
	"go/types"
	"sync"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/dbEngine"
)

func TestColumn_BasicType(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   types.BasicKind
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.BasicType(); got != tt.want {
				t.Errorf("BasicType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_BasicTypeInfo(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   types.BasicInfo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.BasicTypeInfo(); got != tt.want {
				t.Errorf("BasicTypeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_CharacterMaximumLength(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.CharacterMaximumLength(); got != tt.want {
				t.Errorf("CharacterMaximumLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_Comment(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.Comment(); got != tt.want {
				t.Errorf("Comment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_Name(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_Required(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.Required(); got != tt.want {
				t.Errorf("Required() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_Type(t *testing.T) {
	type fields struct {
		Table                  dbEngine.Table
		name                   string
		DataType               string
		ColumnDefault          string
		IsNullable             bool
		CharacterSetName       string
		comment                string
		UdtName                string
		characterMaximumLength int
		PrimaryKey             bool
		IsHidden               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Table:                  tt.fields.Table,
				name:                   tt.fields.name,
				DataType:               tt.fields.DataType,
				ColumnDefault:          tt.fields.ColumnDefault,
				IsNullable:             tt.fields.IsNullable,
				CharacterSetName:       tt.fields.CharacterSetName,
				comment:                tt.fields.comment,
				UdtName:                tt.fields.UdtName,
				characterMaximumLength: tt.fields.characterMaximumLength,
				PrimaryKey:             tt.fields.PrimaryKey,
				IsHidden:               tt.fields.IsHidden,
			}
			if got := c.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConn_GetNotice(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			if got := c.GetNotice(); got != tt.want {
				t.Errorf("GetNotice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConn_GetRoutinesProp(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	tests := []struct {
		name              string
		fields            fields
		wantRoutinesCache map[string]*Routine
		wantErr           bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			gotRoutinesCache, err := c.GetRoutinesProp(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoutinesProp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, gotRoutinesCache, tt.wantRoutinesCache) {
				t.Errorf("GetRoutinesProp() gotRoutinesCache = %v, want %v", gotRoutinesCache, tt.wantRoutinesCache)
			}
		})
	}
}

func TestConn_GetSchema(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]Table
		want1   map[string]*Routine
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			got, got1, err := c.GetSchema(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("GetSchema() got = %v, want %v", got, tt.want)
			}
			if !assert.Equal(t, got1, tt.want1) {
				t.Errorf("GetSchema() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConn_GetStat(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			if got := c.GetStat(); got != tt.want {
				t.Errorf("GetStat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConn_GetTablesProp(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	tests := []struct {
		name            string
		fields          fields
		wantSchemaCache map[string]Table
		wantErr         bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			gotSchemaCache, err := c.GetTablesProp(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTablesProp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, gotSchemaCache, tt.wantSchemaCache) {
				t.Errorf("GetTablesProp() gotSchemaCache = %v, want %v", gotSchemaCache, tt.wantSchemaCache)
			}
		})
	}
}

func TestConn_InitConn(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		ctx   context.Context
		dbURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}

			assert.Implements(t, (*dbEngine.Connection)(nil), c)
			if err := c.InitConn(tt.args.ctx, tt.args.dbURL); (err != nil) != tt.wantErr {
				t.Errorf("InitConn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConn_SelectAndRunEach(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		each func(values []interface{}, columns []pgproto3.FieldDescription) error
		sql  string
		args []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			if err := c.SelectAndRunEach(context.TODO(), tt.args.each, tt.args.sql, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("SelectAndRunEach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConn_SelectAndScanEach(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		each     func() error
		rowValue dbEngine.RowScanner
		sql      string
		args     []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			if err := c.SelectAndScanEach(context.TODO(), tt.args.each, tt.args.rowValue, tt.args.sql, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("SelectAndScanEach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConn_StartChannels(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			assert.Implements(t, (*dbEngine.Connection)(nil), c)
		})
	}
}

func TestConn_addNoticeToErrLog(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			if got := c.addNoticeToErrLog(tt.args.args...); !assert.Equal(t, got, tt.want) {
				t.Errorf("addNoticeToErrLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConn_listen(t *testing.T) {
	type fields struct {
		Pool          *pgxpool.Pool
		Config        *pgxpool.Config
		Notice        *pgconn.Notice
		AfterConnect  fncConn
		BeforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
		ctxPool       context.Context
		Cancel        context.CancelFunc
	}
	type args struct {
		ch string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conn{
				Pool:          tt.fields.Pool,
				Config:        tt.fields.Config,
				Notice:        tt.fields.Notice,
				AfterConnect:  tt.fields.AfterConnect,
				BeforeAcquire: tt.fields.BeforeAcquire,
				channels:      tt.fields.channels,
				ctxPool:       tt.fields.ctxPool,
				Cancel:        tt.fields.Cancel,
			}
			assert.Implements(t, (*dbEngine.Connection)(nil), c)

		})
	}
}

func TestNewColumn(t *testing.T) {
	type args struct {
		table                  dbEngine.Table
		name                   string
		dataType               string
		columnDefault          string
		isNullable             bool
		characterSetName       string
		comment                string
		udtName                string
		characterMaximumLength int
		primaryKey             bool
		isHidden               bool
	}
	tests := []struct {
		name string
		args args
		want *Column
	}{
		// TODO: Add test cases.
		{
			"int",
			args{name: "intField", dataType: "int4"},
			&Column{name: "intField", DataType: "int4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewColumn(tt.args.table, tt.args.name, tt.args.dataType, tt.args.columnDefault, tt.args.isNullable, tt.args.characterSetName, tt.args.comment, tt.args.udtName, tt.args.characterMaximumLength, tt.args.primaryKey, tt.args.isHidden)
			assert.Implements(t, (*dbEngine.Column)(nil), got)
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("NewColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewColumnPone(t *testing.T) {
	type args struct {
		name                   string
		comment                string
		characterMaximumLength int
	}
	tests := []struct {
		name string
		args args
		want *Column
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewColumnPone(tt.args.name, tt.args.comment, tt.args.characterMaximumLength); !assert.Equal(t, got, tt.want) {
				t.Errorf("NewColumnPone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConn(t *testing.T) {
	type args struct {
		afterConnect  fncConn
		beforeAcquire func(context.Context, *pgx.Conn) bool
		channels      []string
	}
	tests := []struct {
		name string
		args args
		want *Conn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConn(tt.args.afterConnect, tt.args.beforeAcquire, tt.args.channels...); !assert.Equal(t, got, tt.want) {
				t.Errorf("NewConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoutine_Call(t *testing.T) {
	type fields struct {
		conn    *Conn
		Name    string
		ID      int
		Comment string
		Fields  []*PgxRoutineParams
		params  []*PgxRoutineParams
		Overlay *Routine
		Type    string
		lock    sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Routine{
				conn:    tt.fields.conn,
				name:    tt.fields.Name,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				Fields:  tt.fields.Fields,
				params:  tt.fields.params,
				Overlay: tt.fields.Overlay,
				Type:    tt.fields.Type,
				lock:    tt.fields.lock,
			}
			assert.Implements(t, (*dbEngine.Routine)(nil), r)

		})
	}
}

func TestRoutine_GetParams(t *testing.T) {
	type fields struct {
		conn    *Conn
		Name    string
		ID      int
		Comment string
		Fields  []*PgxRoutineParams
		params  []*PgxRoutineParams
		Overlay *Routine
		Type    string
		lock    sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routine{
				conn:    tt.fields.conn,
				name:    tt.fields.Name,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				Fields:  tt.fields.Fields,
				params:  tt.fields.params,
				Overlay: tt.fields.Overlay,
				Type:    tt.fields.Type,
				lock:    tt.fields.lock,
			}
			if err := r.GetParams(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("GetParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoutine_Params(t *testing.T) {
	type fields struct {
		conn    *Conn
		Name    string
		ID      int
		Comment string
		Fields  []*PgxRoutineParams
		params  []*PgxRoutineParams
		Overlay *Routine
		Type    string
		lock    sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Routine{
				conn:    tt.fields.conn,
				name:    tt.fields.Name,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				Fields:  tt.fields.Fields,
				params:  tt.fields.params,
				Overlay: tt.fields.Overlay,
				Type:    tt.fields.Type,
				lock:    tt.fields.lock,
			}
			assert.Implements(t, (*dbEngine.Routine)(nil), r)
		})
	}
}

func TestRoutine_Select(t *testing.T) {
	type fields struct {
		conn    *Conn
		Name    string
		ID      int
		Comment string
		Fields  []*PgxRoutineParams
		params  []*PgxRoutineParams
		Overlay *Routine
		Type    string
		lock    sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Routine{
				conn:    tt.fields.conn,
				name:    tt.fields.Name,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				Fields:  tt.fields.Fields,
				params:  tt.fields.params,
				Overlay: tt.fields.Overlay,
				Type:    tt.fields.Type,
				lock:    tt.fields.lock,
			}
			assert.Implements(t, (*dbEngine.Routine)(nil), r)
		})
	}
}

func TestTable_Columns(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	tests := []struct {
		name   string
		fields fields
		want   []dbEngine.Column
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if got := t.Columns(); !assert.Equal(t1, got, tt.want) {
				t1.Errorf("Columns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_GetColumns(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if err := t.GetColumns(context.TODO()); (err != nil) != tt.wantErr {
				t1.Errorf("GetColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTable_GetFields(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		columns []dbEngine.Column
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if got := t.GetFields(tt.args.columns); !assert.Equal(t1, got, tt.want) {
				t1.Errorf("GetFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTable_Insert(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			assert.Implements(t1, (*dbEngine.Table)(nil), t)
		})
	}
}

func TestTable_Name(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if got := t.Name(); got != tt.want {
				t1.Errorf("Name() = %v, want %v", got, tt.want)
			}
			assert.Implements(t1, (*dbEngine.Table)(nil), t)
		})
	}
}

func TestTable_Select(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			assert.Implements(t1, (*dbEngine.Table)(nil), t)
		})
	}
}

func TestTable_SelectAndRunEach(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		ctx  context.Context
		each func(values []interface{}, columns []dbEngine.Column) error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if err := t.SelectAndRunEach(tt.args.ctx, tt.args.each); (err != nil) != tt.wantErr {
				t1.Errorf("SelectAndRunEach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTable_SelectAndScanEach(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		ctx      context.Context
		each     func() error
		rowValue dbEngine.RowScanner
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if err := t.SelectAndScanEach(tt.args.ctx, tt.args.each, tt.args.rowValue); (err != nil) != tt.wantErr {
				t1.Errorf("SelectAndScanEach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTable_readColumnRow(t1 *testing.T) {
	type fields struct {
		conn    *Conn
		name    string
		Type    string
		ID      int
		Comment string
		Fields  []*Column
		PK      string
	}
	type args struct {
		values  []interface{}
		columns []pgproto3.FieldDescription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := Table{
				conn:    tt.fields.conn,
				name:    tt.fields.name,
				Type:    tt.fields.Type,
				ID:      tt.fields.ID,
				Comment: tt.fields.Comment,
				columns: tt.fields.Fields,
				PK:      tt.fields.PK,
			}
			if err := t.readColumnRow(tt.args.values, tt.args.columns); (err != nil) != tt.wantErr {
				t1.Errorf("readColumnRow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
