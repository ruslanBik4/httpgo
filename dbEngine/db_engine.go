package dbEngine

type DB struct {
    Cfg         map[string] interface{}
    Tables      []*Table
    Routines    []*Routine
}

type Table interface {
    Select()
    Insert()
}

type Routine interface {
    Select()
    Call()
}

type Column interface{
    CharacterMaximumLength() int
    Name() string
}


type RowScanner interface {
	GetFields([] Column) []interface{}
}

