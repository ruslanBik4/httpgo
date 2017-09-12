package qb

import (
	"fmt"
	"regexp"
	"strings"
)

var sqlCommand1 = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name, COUNT(*) count1, COUNT(tab1.*) as count2, COUNT(tab1.min) as count3   FROM AAA a INNER JOIN BBB b ON a.id = b.id_user INNER JOIN CCC c ON b.id = c.id_customer GROUP BY id_user ORDER BY fname, second_name"

//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a"

//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name  FROM AAA a"
//var sql = "SELECT   FROM AAA a INNER JOIN BBB b ON a.id = b.id_user"

const regSelect = `^SELECT\s+(?P<fields>.*)\s+FROM\s+`
const regField = `^\s*(((?P<func_name>[a-zA-Z0-9]*)[(])?[ ]*(((?P<field_table>[a-zA-Z0-9_]*)[.])?(?P<field_name>[a-zA-Z0-9_*]*)){1})[)]?(((\s+AS\s+)|\s+)(?P<alias>[a-zA-Z0-9_]*))?`
const regFrom = `FROM\s*(?P<table>[.a-zA-Z0-9_]*)\s*(?P<table_alias>[a-zA-Z0-9_]*)*\s*`
const regJoin = "(INNER|LEFT|RIGHT)[ ]*JOIN[ ]*(?P<join_table_name>[.a-zA-Z0-9_]*)[ ]*(?P<join_table_alias>[a-zA-Z0-9_]*)*[ ]*ON[ ]*((?P<on_left_table>[a-zA-Z0-9_]*)[.])+(?P<on_left_field>[a-zA-Z0-9_]*)[ ]*=[ ]*((?P<on_right_table>[a-zA-Z0-9_]*)[.])+(?P<on_right_field>[a-zA-Z0-9_]*)"

type tSQLStructure struct {
	Fields []*tSqlField
	From   *tSqlFrom
	Joins  []*tSqlJoin
}
type tSqlField struct {
	fun   string
	table string
	name  string
	alias string
}
type tSqlFrom struct {
	tableName  string
	tableAlias string
}
type tSqlJoin struct {
	tableName  string
	tableAlias string
	onLeft     *tSqlJoinOn
	onRight    *tSqlJoinOn
}
type tSqlJoinOn struct {
	tableName string
	fieldName string
}

func main() {
	sqlStructure, ok := parse(sqlCommand1)
	if ok {
		fmt.Println("***** Fields *****")
		for _, field := range sqlStructure.Fields {
			fmt.Println(fmt.Sprintf("Function Name: '%s', Table Name: '%s', Field Name: '%s', Alias: '%s'", field.fun, field.table, field.name, field.alias))
		}
		fmt.Println("******************")

		fmt.Println("***** FROM *****")
		fmt.Println("Table Name: " + sqlStructure.From.tableName)
		fmt.Println("Table Alias: " + sqlStructure.From.tableAlias)
		fmt.Println("******************")

		fmt.Println("***** Joins *****")
		for _, v := range sqlStructure.Joins {
			fmt.Println("Join Table Name: " + v.tableName)
			fmt.Println("Join Table Alias: " + v.tableAlias)
			fmt.Println("Left Join Table: " + v.onLeft.tableName)
			fmt.Println("Left Join Field: " + v.onLeft.fieldName)
			fmt.Println("Right Join Table: " + v.onRight.tableName)
			fmt.Println("Right Join Field: " + v.onRight.fieldName)
			fmt.Println("******************")
		}
	}
}

func parse(sql string) (*tSQLStructure, bool) {
	fields, ok := getFields(sql)
	if !ok {
		return nil, false
	}
	from, ok := getFrom(sql)
	if !ok {
		return nil, false
	}
	joins, ok := getJoins(sql)
	if !ok {
		return nil, false
	}
	return &tSQLStructure{
		Fields: fields,
		From:   from,
		Joins:  joins,
	}, true
}

func getFrom(sql string) (*tSqlFrom, bool) {
	var compRegEx = regexp.MustCompile(regFrom)
	var match []string = compRegEx.FindStringSubmatch(sql)

	var tableName string = ""
	var tableAlias string = ""
	for i, name := range compRegEx.SubexpNames() {
		if name == "table" && i > 0 && i <= len(match) {
			tableName = match[i]
		}
		if name == "table_alias" && i > 0 && i <= len(match) {
			tableAlias = match[i]
		}
	}

	if tableName == "" && tableAlias == "" {
		return nil, false
	} else {
		return &tSqlFrom{tableName: tableName, tableAlias: tableAlias}, true
	}
}

func getJoins(sql string) ([]*tSqlJoin, bool) {
	var compRegEx = regexp.MustCompile(regJoin)
	var match [][]string = compRegEx.FindAllStringSubmatch(sql, -1)
	var result []*tSqlJoin = make([]*tSqlJoin, 0)
	var groupNames []string = compRegEx.SubexpNames()
	for _, v := range match {
		if join, ok := getJoin(v, groupNames); ok {
			result = append(result, join)
		} else {
			return nil, false
		}
	}
	return result, true
}

func getJoin(join []string, groupNames []string) (*tSqlJoin, bool) {
	var joinTableName string = ""
	var joinTableAlias string = ""
	var onLeftTable string = ""
	var onLeftField string = ""
	var onRightTable string = ""
	var onRightField string = ""

	for i, name := range groupNames {
		if name == "join_table_name" && i > 0 && i <= len(join) {
			joinTableName = join[i]
		}
		if name == "join_table_alias" && i > 0 && i <= len(join) {
			joinTableAlias = join[i]
		}
		if name == "on_left_table" && i > 0 && i <= len(join) {
			onLeftTable = join[i]
		}
		if name == "on_left_field" && i > 0 && i <= len(join) {
			onLeftField = join[i]
		}
		if name == "on_right_table" && i > 0 && i <= len(join) {
			onRightTable = join[i]
		}
		if name == "on_right_field" && i > 0 && i <= len(join) {
			onRightField = join[i]
		}
	}

	if (joinTableName == "" && joinTableAlias == "") ||
		onLeftTable == "" || onLeftField == "" ||
		onRightTable == "" || onRightField == "" {
		return nil, false
	} else {
		return &tSqlJoin{
			tableName:  joinTableName,
			tableAlias: joinTableAlias,
			onLeft: &tSqlJoinOn{
				tableName: onLeftTable,
				fieldName: onLeftField,
			},
			onRight: &tSqlJoinOn{
				tableName: onRightTable,
				fieldName: onRightField,
			},
		}, true
	}
}

func getFields(sql string) ([]*tSqlField, bool) {
	if fieldsText, ok := getTextSelectFields(sql); ok {
		if fields, ok := getFields(fieldsText); ok {
			return fields, ok
		}
	}
	return nil, false
}

func getTextSelectFields(sql string) (fields string, ok bool) {
	var compRegEx = regexp.MustCompile(regSelect)
	var match []string = compRegEx.FindStringSubmatch(sql)

	var groupNames []string = compRegEx.SubexpNames()
	for i, name := range groupNames {
		if name == "fields" && i > 0 && i <= len(match) {
			fields = match[i]
		}
	}

	if fields == "" {
		return "", false
	} else {
		return fields, true
	}
}

//func getFields(textFields string) ([]*tSqlField, bool) {
//	var fieldItems []string = strings.Split(textFields, ",")
//	var reg *regexp.Regexp = regexp.MustCompile(regField)
//	var groupNames []string = reg.SubexpNames()
//
//	var result []*tSqlField = make([]*tSqlField, 0)
//	for _, text := range fieldItems {
//		if field, ok := getField(text, groupNames, reg); ok {
//			result = append(result, field)
//		} else {
//			return nil, false
//		}
//	}
//
//	return result, true
//}

func getField(text string, groupNames []string, reg *regexp.Regexp) (*tSqlField, bool) {
	var fieldNote string = strings.TrimSpace(text)
	var elements []string = reg.FindStringSubmatch(fieldNote)

	var funcName = ""
	var fieldTableName = ""
	var fieldName = ""
	var fieldAlias = ""

	for i, name := range groupNames {
		if name == "func_name" {
			if i > 0 && i <= len(elements) {
				funcName = strings.TrimRight(elements[i], ".")
			}
		}
		if name == "field_table" {
			if i > 0 && i <= len(elements) {
				fieldTableName = strings.TrimRight(elements[i], ".")
			}
		}
		if name == "field_name" {
			if i > 0 && i <= len(elements) {
				fieldName = elements[i]
			}
		}
		if name == "alias" {
			if i > 0 && i <= len(elements) {
				fieldAlias = elements[i]
			}
		}
	}

	if fieldTableName == "" && fieldName == "" {
		return nil, false
	} else {
		return &tSqlField{fun: funcName, table: fieldTableName, name: fieldName, alias: fieldAlias}, true
	}
}
