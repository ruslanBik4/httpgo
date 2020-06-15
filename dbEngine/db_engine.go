package dbEngine

type DB struct {
	Cfg      map[string]interface{}
	Tables   []*Table
	Routines []*Routine
}

type Table interface {
	Columns() []Column
	Insert()
	Select()
	SelectAndScanEach(each func() error, rowValue RowScanner) error
	SelectAndRunEach(each func(values []interface{}, columns []Column) error) error
}

type Routine interface {
	Select()
	Call()
	Params()
}

type Column interface {
	CharacterMaximumLength() int
	Comment() string
	Name() string
	Type() string
	Required() bool
}

type RowScanner interface {
	GetFields([]Column) []interface{}
}
