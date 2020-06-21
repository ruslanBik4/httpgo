// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dbEngine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/logs"
)

type ParserTableDDL struct {
	Table
	*DB
	err          error
	mapParse     []func(string) bool
	isCreateDone bool
}

func NewtableParser(table Table, db *DB) ParserTableDDL {
	t := ParserTableDDL{Table: table, DB: db}
	t.mapParse = []func(string) bool{
		t.updateTable,
		t.addComment,
		t.updateIndex,
		t.skipPartition,
	}

	return t
}

func (p ParserTableDDL) Parse(ddl string) error {
	ddl = strings.ToLower(ddl)
	for _, sql := range strings.Split(ddl, ";") {
		if !p.execSql(strings.Trim(sql, "\n")) {
			logs.DebugLog("unknow sql", sql)
		}

		if p.err != nil {
			p.logError(p.err, ddl, p.Name())

		}

		p.err = nil

	}

	return nil
}

func (p ParserTableDDL) logError(err error, ddlSQL string, fileName string) {
	pgErr, ok := err.(*pgconn.PgError)
	if ok && pgErr.Position > 0 {
		line := strings.Count(ddlSQL[:pgErr.Position-1], "\n") + 1
		fmt.Printf("\033[%d;1m%s\033[0m %v:%d: %s %#v\n", 35, "[[ERROR]]", fileName, line, pgErr.Message, pgErr)
	} else {
		logs.ErrorLog(err, prefix, fileName)
	}
}

func (p ParserTableDDL) execSql(sql string) bool {
	for i, fnc := range p.mapParse {
		if (!p.isCreateDone || (i > 0)) && fnc(sql) {
			p.isCreateDone = p.isCreateDone || (i == 0)
			return true
		}
	}

	return false
}

func (p ParserTableDDL) addComment(ddl string) bool {
	if !strings.HasPrefix(ddl, "comment") {
		return false
	}

	err := p.Conn.ExecDDL(context.TODO(), ddl)
	if err == nil {
		logs.StatusLog(prefix, ddl)
	} else if isErrorAlreadyExists(err) {
		err = nil
	} else if err != nil {
		p.logError(err, ddl, p.Name())
	}

	return true
}

var regPartionTable = regexp.MustCompile(`create\s+table\s+(\w+)\s+partition`)

func (p ParserTableDDL) skipPartition(ddl string) bool {
	fields := regPartionTable.FindStringSubmatch(ddl)
	if len(fields) == 0 {
		return false
	}

	_, ok := p.Tables[fields[1]]
	if !ok {
		err := p.Conn.ExecDDL(context.TODO(), ddl)
		if err == nil {
			logs.StatusLog(prefix, ddl)
		} else if isErrorAlreadyExists(err) {
			err = nil
		} else if err != nil {
			p.logError(err, ddl, p.Name())
		}
	}

	return true
}

var regTable = regexp.MustCompile(`create\s+table\s+(?P<name>\w+)\s+\((?P<fields>(\s*(\w*)\s*(?P<define>[\w\[\]':\s]*(\(\d+\))?[\w\s]*),?)*)\s*(primary\s+key\s*\([^)]+\))?\s*\)`)

var regField = regexp.MustCompile(`(\w+)\s+([\w()\[\]\s]+)`)

func (p ParserTableDDL) updateTable(ddl string) bool {
	var err error
	//:= string(bytes.Replace(ddl, []byte("\n"), []byte(""), -1)))
	fields := regTable.FindStringSubmatch(ddl)
	if len(fields) == 0 {
		return false
	}

	for i, name := range regTable.SubexpNames() {
		if !(i < len(fields)) {
			return false
		}

		switch name {
		case "":
		case "p":
			if fields[i] != p.Name() {
				p.err = errors.New("bad p name! " + fields[i])
				return false
			}
		case "fields":

			nameFields := strings.Split(fields[i], ",")
			for _, name := range nameFields {

				title := regField.FindStringSubmatch(name)
				if len(title) < 3 ||
					strings.HasPrefix(strings.ToUpper(title[1]), "primary") ||
					strings.HasPrefix(strings.ToUpper(title[1]), "constraint") {
					continue
				}

				fieldName := title[1]
				if fs := p.FindField(fieldName); fs == nil {
					sql := " ADD COLUMN " + name
					err = p.addColumn(sql, fieldName)
				} else {
					err = p.checkColumn(title[2], fs)
				}

			}
		}
	}

	p.err = err
	return true
}

func (p ParserTableDDL) checkColumn(title string, fs Column) (err error) {
	res := fs.CheckAttr(title)
	fieldName := fs.Name()
	if res > "" {
		err = ErrNotFoundField{
			Table:     p.Name(),
			FieldName: fieldName,
		}
		// change length
		if strings.Contains(res, "has length") {
			logs.DebugLog(res)
			attr := strings.Split(title, " ")
			if attr[0] == "character" {
				attr[0] += " " + attr[1]
			}

			sql := fmt.Sprintf(" type %s using %s::%[1]s", attr[0], fieldName)
			err = p.alterColumn(sql, fieldName, title, fs)
		}

		// change type
		if strings.Contains(res, "type") {
			attr := strings.Split(title, " ")
			if attr[0] == "double" {
				attr[0] += " " + attr[1]
			}
			sql := fmt.Sprintf(" type %s using %s::%[1]s", attr[0], fieldName)
			if attr[0] == "money" && fs.Type() == "double precision" {
				sql = fmt.Sprintf(
					" type %s using %s::numeric::%[1]s",
					attr[0], fieldName)
			}

			err = p.alterColumn(sql, fieldName, title, fs)
		}

		// set not nullable
		if strings.Contains(res, "is nullable") {
			err = p.alterColumn(" set not null", fieldName, title, fs)
			if err != nil {
				logs.ErrorLog(err)
			} else {
				fs.SetNullable(true)
			}
		}

		// set nullable
		if strings.Contains(res, "is not nullable") {
			err = p.alterColumn(" drop not null", fieldName, title, fs)
			if err != nil {
				logs.ErrorLog(err)
			} else {
				fs.SetNullable(false)
			}
		}

	}

	return err
}

func (p ParserTableDDL) updateIndex(ddl string) bool {
	fields := ddlIndex.FindStringSubmatch(ddl)
	if len(fields) == 0 {
		return false
	}

	ind, err := p.createIndex(fields)
	if err != nil {
		logs.ErrorLog(err, ddl)
		return true
	}

	if p.FindIndex(ind.Name()) != nil {
		logs.StatusLog("index '%s' exists! ", ind.Name())
		//todo: check columns of index
		return true
	}

	err = p.Conn.ExecDDL(context.TODO(), ddl)
	if err == nil {
		logs.StatusLog(prefix, ddl, ind)
	} else if isErrorAlreadyExists(err) {
		err = nil
	} else if err != nil {
		p.logError(err, ddl, p.Name())
	}

	return true
}

var ddlIndex = regexp.MustCompile(`create(?:\s+unique)?\s+index(?:\s+if\s+not\s+exists)?\s+(?P<index>\w+)\s+on\s+(?P<table>\w+)(?:\s+using\s+\w+)?\s*\((?P<columns>[^;]+?)\)\s*(where\s+[^)]\))?`)

func (p ParserTableDDL) createIndex(fields []string) (Index, error) {

	var ind Index
	for i, name := range ddlIndex.SubexpNames() {
		if !(i < len(fields)) {
			return nil, errors.New("out if fields!" + name)
		}

		switch name {
		case "":
		case "p":
			if fields[i] != p.Name() {
				return nil, errors.New("bad p name! " + fields[i])
			}
		case "index":
			// todo implement
			// ind.Name = fields[i]
		case "columns":
			//nameFields := strings.Split(fields[i], ",")
			//for _, name := range nameFields {
			logs.StatusLog("new index column: ", fields[i])
			//}
		default:
			logs.StatusLog("%s %s", name, fields[i])
		}

	}

	return ind, nil
}

func (p ParserTableDDL) addColumn(sAlter string, fieldName string) error {
	err := p.Conn.ExecDDL(context.TODO(), "ALTER TABLE "+p.Name()+sAlter)
	if err != nil {
		logs.ErrorLog(err, `. Field %s.%s`, p.Name, fieldName)
	} else {
		logs.StatusLog("[DB CONFIG] ", p.Name, sAlter)
		p.RecacheField(fieldName)
	}

	return err
}

func (p ParserTableDDL) alterColumn(sAlter string, fieldName, title string, fs Column) error {
	sql := "ALTER TABLE " + p.Name() + " alter column " + fieldName + sAlter
	err := p.Conn.ExecDDL(context.TODO(), sql)
	if err != nil {
		logs.ErrorLog(err,
			`. Field %s.%s, different with define: '%s' %s, sql: %s`,
			p.Name, fieldName, title, fs, sql)
	} else {
		logs.StatusLog("[DB CONFIG] %s ", sql)
		p.RecacheField(fieldName)
	}

	return err
}
