package dbEngine

type StringColumn struct {
	comment, name string
	req           bool
}

func NewStringColumn(name, comment string, req bool) *StringColumn {
	return &StringColumn{comment, name, req}
}

func (s *StringColumn) Comment() string {
	return s.comment
}

func (s *StringColumn) Type() string {
	return "string"
}

func (s *StringColumn) Required() bool {
	return s.req
}

func (s *StringColumn) Name() string {
	return s.name
}

func (s *StringColumn) CharacterMaximumLength() int {
	return len(s.name)
}

func SimpleColumns(names ...string) []Column {
	s := make([]Column, len(names))
	for i, name := range names {
		s[i] = NewStringColumn(name, name, false)
	}

	return s
}
