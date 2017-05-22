package qb

import (
"regexp"
"fmt"
"strings"
)

//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a INNER JOIN BBB b ON a.id = b.id_user"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a LEFT JOIN BBB b ON a.id = b.id_user"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a RIGHT JOIN BBB b ON a.id = b.id_user"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a RIGHT JOIN BBB b ON a.id = b.id_user INNER JOIN CCC c ON a.id = c.id_customer"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a GROUP BY id_user"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a ORDER BY id_user"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a GROUP BY id_user ORDER BY id_user"

var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name, COUNT(*) count1, COUNT(tab1.*) as count2, COUNT(tab1.min) as count3   FROM AAA a INNER JOIN BBB b ON a.id = b.id_user INNER JOIN CCC c ON b.id = c.id_customer GROUP BY id_user ORDER BY fname, second_name"
//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name FROM AAA a"

//var sql = "SELECT  a.id as id_user, a.group group_id, a.name, a.first_name fname, last_name lname, second_name  FROM AAA a"
//var sql = "SELECT   FROM AAA a INNER JOIN BBB b ON a.id = b.id_user"

const regSelect = "^SELECT[ ]+(?P<fields>.*)[ ]+FROM[ ]+"
const regField = "^[ ]*(((?P<func_name>[a-zA-Z0-9]*)[(])?[ ]*(((?P<field_table>[a-zA-Z0-9_]*)[.])?(?P<field_name>[a-zA-Z0-9_*]*)){1})[)]?((([ ]+as[ ]+)|[ ]+)(?P<alias>[a-zA-Z0-9_]*))?"
const regFrom = "FROM[ ]*(?P<table>[.a-zA-Z0-9_]*)[ ]*(?P<table_alias>[a-zA-Z0-9_]*)*[ ]*"
const regJoin = "(INNER|LEFT|RIGHT)[ ]*JOIN[ ]*(?P<join_table_name>[.a-zA-Z0-9_]*)[ ]*(?P<join_table_alias>[a-zA-Z0-9_]*)*[ ]*ON[ ]*((?P<on_left_table>[a-zA-Z0-9_]*)[.])+(?P<on_left_field>[a-zA-Z0-9_]*)[ ]*=[ ]*((?P<on_right_table>[a-zA-Z0-9_]*)[.])+(?P<on_right_field>[a-zA-Z0-9_]*)"

type SQLStructure struct {
	Fields []*SqlField
	From   *SqlFrom
	Joins  []*SqlJoin
}
type SqlField struct {
	fun   string
	table string
	name  string
	alias string
}
type SqlFrom struct {
	tableName  string
	tableAlias string
}
type SqlJoin struct {
	tableName  string
	tableAlias string
	onLeft     *SqlJoinOn
	onRight    *SqlJoinOn
}
type SqlJoinOn struct {
	tableName string
	fieldName string
}

func main() {
	sqlStructure, ok := Parse(sql)
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

func Parse(sql string) (*SQLStructure, bool) {
	fields, ok := GetFields(sql)
	if !ok {
		return nil, false
	}
	from, ok := GetFrom(sql)
	if !ok {
		return nil, false
	}
	joins, ok := GetJoins(sql)
	if !ok {
		return nil, false
	}
	return &SQLStructure{
		Fields: fields,
		From:   from,
		Joins:  joins,
	}, true
}

func GetFrom(sql string) (*SqlFrom, bool) {
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
		return &SqlFrom{tableName: tableName, tableAlias: tableAlias}, true
	}
}

func GetJoins(sql string) ([]*SqlJoin, bool) {
	var compRegEx = regexp.MustCompile(regJoin)
	var match [][]string = compRegEx.FindAllStringSubmatch(sql, -1)
	var result []*SqlJoin = make([]*SqlJoin, 0)
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

func getJoin(join []string, groupNames []string) (*SqlJoin, bool) {
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
		return &SqlJoin{
			tableName:  joinTableName,
			tableAlias: joinTableAlias,
			onLeft: &SqlJoinOn{
				tableName: onLeftTable,
				fieldName: onLeftField,
			},
			onRight: &SqlJoinOn{
				tableName: onRightTable,
				fieldName: onRightField,
			},
		}, true
	}
}

func GetFields(sql string) ([]*SqlField, bool) {
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

func getFields(textFields string) ([]*SqlField, bool) {
	var fieldItems []string = strings.Split(textFields, ",")
	var reg *regexp.Regexp = regexp.MustCompile(regField)
	var groupNames [] string = reg.SubexpNames()

	var result []*SqlField = make([]*SqlField, 0)
	for _, text := range fieldItems {
		if field, ok := getField(text, groupNames, reg); ok {
			result = append(result, field)
		} else {
			return nil, false
		}
	}

	return result, true
}

func getField(text string, groupNames []string, reg *regexp.Regexp) (*SqlField, bool) {
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
		return &SqlField{fun: funcName, table: fieldTableName, name: fieldName, alias: fieldAlias}, true
	}
}