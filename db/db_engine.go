package db_engine

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

type Field interface{
}


type RowScanner interface {
	GetFields([] Field) []interface{}
}

