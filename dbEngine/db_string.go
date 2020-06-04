package dbEngine


type StringColumn struct {
    name string
}

func NewStringColumn(name string) *StringColumn{
    return &StringColumn{name}
}

func (s *StringColumn) Name() string {
    return s.name
}

func (s *StringColumn) CharacterMaximumLength() int {
    return len(s.name)
}