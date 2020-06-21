package dbEngine

import (
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/logs"
)

type DB struct {
	Cfg      map[string]interface{}
	Conn     Connection
	Tables   map[string]Table
	Routines map[string]Routine
}

func NewDB(ctx context.Context, conn Connection) (*DB, error) {
	db := &DB{Conn: conn}
	if dbUrl, ok := ctx.Value("dbURL").(string); ok {
		logs.DebugLog("init conn with url - ", dbUrl)
		err := conn.InitConn(ctx, dbUrl)
		if err != nil {
			return nil, errors.Wrap(err, "initConn")
		}

		if doRead, ok := ctx.Value("fillSchema").(bool); ok && doRead {
			db.Tables, db.Routines, err = conn.GetSchema(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "initConn")
			}
		}
		if mPath, ok := ctx.Value("migration").(string); ok {
			err = filepath.Walk(filepath.Join(mPath, "table"), db.ReadTableSQL)
			if err != nil {
				return nil, errors.Wrap(err, "migration")
			}
		}
	}

	return db, nil
}

func (db *DB) ReadTableSQL(path string, info os.FileInfo, err error) error {
	if (err != nil) || ((info != nil) && info.IsDir()) {
		return nil
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".ddl":
		fileName := filepath.Base(path)
		tableName := strings.TrimSuffix(fileName, ext)
		ddl, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		table, ok := db.Tables[tableName]
		if !ok {
			err = db.Conn.ExecDDL(context.TODO(), string(ddl))
			if err == nil {
				table = db.Conn.NewTable(tableName, "table")
				err = table.GetColumns(context.TODO())
				if err == nil {
					db.Tables[tableName] = table
					logs.StatusLog("New table add to DB", tableName)
				}
				return err
			} else {
				logs.ErrorLog(err, "table - "+tableName)
			}
		} else {
			return NewtableParser(table, db).Parse(string(ddl))
		}

	default:
		return nil
	}

	return err
}

type Connection interface {
	InitConn(ctx context.Context, dbURL string) error
	GetSchema(ctx context.Context) (map[string]Table, map[string]Routine, error)
	GetStat() string
	ExecDDL(ctx context.Context, sql string, args ...interface{}) error
	NewTable(name, typ string) Table
}

type Table interface {
	Columns() []Column
	FindField(name string) Column
	FindIndex(name string) Index
	GetColumns(ctx context.Context) error
	Insert(ctx context.Context)
	Name() string
	RecacheField(nameColumn string) Column
	Select(ctx context.Context)
	SelectAndScanEach(ctx context.Context, each func() error, rowValue RowScanner) error
	SelectAndRunEach(ctx context.Context, each func(values []interface{}, columns []Column) error) error
}

type Routine interface {
	Name() string
	Select() error
	Call()
	Params()
}

type Column interface {
	BasicType() types.BasicKind
	BasicTypeInfo() types.BasicInfo
	CheckAttr(fieldDefine string) string
	CharacterMaximumLength() int
	Comment() string
	Name() string
	Type() string
	Required() bool
	SetNullable(bool)
}

type Index interface {
	Name() string
}

type RowScanner interface {
	GetFields([]Column) []interface{}
}
