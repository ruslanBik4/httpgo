package dbEngine

import (
	"go/types"
	"golang.org/x/net/context"
)

type DB struct {
	Cfg      map[string]interface{}
	Conn	 Connection
	Tables   []*Table
	Routines []*Routine
}

type Connection interface {
	InitConn(ctx context.Context, dbURL string) error 
	GetStat() string
}

type Table interface {
	Columns() []Column
	Insert(ctx context.Context)
	Name() string
	Select(ctx context.Context)
	SelectAndScanEach(ctx context.Context, each func() error, rowValue RowScanner) error
	SelectAndRunEach(ctx context.Context, each func(values []interface{}, columns []Column) error) error
}

type Routine interface {
	Select()
	Call()
	Params()
}

type Column interface {
	BasicType() types.BasicKind	
	BasicTypeInfo() types.BasicInfo	
	CharacterMaximumLength() int
	Comment() string
	Name() string
	Type() string
	Required() bool
}

type RowScanner interface {
	GetFields([]Column) []interface{}
}
