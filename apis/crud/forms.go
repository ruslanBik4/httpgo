// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	. "github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/json"
)

func RoutesFromDB(ctx context.Context, tables ...string) apis.ApiRoutes {
	DB, ok := ctx.Value("DB").(*DB)
	if !ok {
		logs.ErrorLog(ErrDBNotFound, "not in context")
		return nil
	}

	patternList, ok := DB.Tables["patterns_list"]
	if !ok {
		logs.ErrorLog(ErrNotFoundTable{Table: "patterns_list"}, "it wiil be problem on validations fields")
	}

	pathVersion, ok := ctx.Value(PathVersion).(string)
	if !ok {
		pathVersion = PathVersion
	}

	preRoute := pathVersion + "/table/"
	routes := make(apis.ApiRoutes, 0)
	inParams := []apis.InParam{ParamsLang, ParamsGetFormActions}

	for tableName, table := range DB.Tables {
		if len(tables) > 0 {
			for _, name := range tables {
				if name == tableName {
					goto createRoutes
				}
			}
		}

	createRoutes:
		rUpd := &apis.ApiRoute{
			Desc:      "update table '" + tableName + "' data",
			Method:    apis.POST,
			Multipart: true,
			//todo: resolve on future
			//DTO:         dtoField{},
			FncAuth:     nil,
			TestFncAuth: nil,
			NeedAuth:    true,
			OnlyAdmin:   false,
			OnlyLocal:   false,
			Params:      inParams,
			Resp:        nil,
		}

		rIns := &apis.ApiRoute{
			Desc:      "insert into table '" + tableName + "' data",
			Method:    apis.POST,
			Multipart: true,
			NeedAuth:  true,
			//DTO:       dtoField{},
			Params: inParams,
		}

		rGet := &apis.ApiRoute{
			Desc:     "get data from table '" + tableName + "'",
			Method:   apis.GET,
			NeedAuth: true,
			Params:   inParams,
		}

		params := make([]string, 0)
		autoIncCols := make([]string, 0)
		priColumns := make([]string, 0)
		basicParams := []apis.InParam{
			ParamsHTML,
			ParamsLang,
		}

		for _, col := range table.Columns() {

			p := newDbApiParams(col)

			rUpd.Params = append(rUpd.Params, p.InParam)
			i := p.InParam
			i.Req = false
			rGet.Params = append(rGet.Params, i)

			if !col.AutoIncrement() {
				p.Req = col.Required()
				p.DefValue = col.Default()
				rIns.Params = append(rIns.Params, p.InParam)
				params = append(params, p.Name)
			} else {
				autoIncCols = append(autoIncCols, p.Name)
			}

			if col.Primary() || (col.Name() == "id") {
				priColumns = append(priColumns, p.Name)
				pForm := p
				pForm.Req = false
				basicParams = append(basicParams, pForm.InParam)
			}
		}

		// if !strings.HasPrefix(table.Name(), "form") {

		pathForm := preRoute + tableName + "/form"

		routes[pathForm] = &apis.ApiRoute{
			Desc:     "get form for insert/update data into " + tableName,
			Fnc:      TableForm(DB, preRoute, table, patternList, priColumns),
			Method:   apis.POST,
			NeedAuth: true,
			Params:   basicParams,
		}

		rUpd.Fnc = TableUpdate(preRoute, table, params, priColumns)
		routes[preRoute+tableName+"/update"] = rUpd

		rIns.Fnc = TableInsert(preRoute, DB, table, params)
		routes[preRoute+tableName+"/put"] = rIns

		params = append(params, autoIncCols...)
		for i := range rGet.Params {
			rGet.Params[i].PartReq = params
		}
		rGet.Fnc = TableSelect(preRoute, table, params)
		routes[preRoute+tableName+"/get"] = rGet
		// report := NewReportJSON(table)
		// routes[preRoute+tableName+"/report"] = report.getRoute()
		// routes[preRoute+tableName+"/data"] = &apis.ApiRoute{
		// 	Desc:   "from  table '" + tableName + "' data",
		// 	Params: basicParams,
		// 	Fnc:    TableData(DB, table, priColumns),
		// }
		//
		// // }
		//
		// routes[preRoute+tableName+"/view"] = &apis.ApiRoute{
		// 	Desc:   "view data of table " + tableName,
		// 	Fnc:    TableView(preRoute, DB, table, patternList, priColumns),
		// 	Params: append(basicParams, ParamsCounter),
		// }

		// routes[preRoute+tableName+"/"] = &apis.ApiRoute{
		// 	Desc: "show row of table according to ID" + tableName,
		// 	// NeedAuth: true,
		// 	Fnc: TableRow(table),
		// 	Params: []apis.InParam{
		// 		ParamsLang,
		// 		{
		// 			Name:     "id",
		// 			DefValue: apis.ApisValues(apis.ChildRoutePath),
		// 			Desc:     "id of photos record for download",
		// 			Req:      false,
		// 			Type:     apis.NewTypeInParam(types.Int32),
		// 		},
		// 	},
		// }
	}

	return routes
}

func TableForm(DB *DB, preRoute string, table, patternList Table, priColumns []string) apis.ApiRouteHandler {
	return func(ctx *fasthttp.RequestCtx) (interface{}, error) {

		// we must copy colsTable into local array
		f := forms.FormField{
			Title:       table.Comment(),
			Action:      preRoute + table.Name(),
			Method:      "POST",
			Description: "",
		}

		colDecors := make([]*forms.ColumnDecor, 0)

		id, ok := int32(0), false
		args := make([]interface{}, len(priColumns))
		columnsTable := table.Columns()
		for i, name := range priColumns {
			if name == "id" {
				id, ok = ctx.UserValue(name).(int32)
				if !ok {
					var s string
					s, ok = ctx.UserValue(apis.ChildRoutePath).(string)
					if ok {
						i, err := strconv.Atoi(s)
						if err != nil {
							return apis.ChildRoutePath, apis.ErrWrongParamsList
						}
						id = int32(i)
					}
				}
				args[i] = id
			} else {
				args[i] = ctx.UserValue(name)
				ok = args[i] != nil
			}

			if !ok {
				break
			}

			for _, col := range columnsTable {
				if col.Name() == name {
					colDec := forms.NewColumnDecor(col, patternList)
					colDec.Value = args[i]
					// if colDec.AutoIncrement() || col.Name() == "id" {
					colDec.IsHidden = true
					colDec.InputType = "hidden"
					// } else {
					// 	colDec.IsReadOnly = true
					// 	colDec.IsDisabled = true
					label, isStr := GetForeignName(ctx, DB, colDec, args[i]).(string)
					if isStr {
						f.Description += " " + label
					}
					// }

					colDecors = append(colDecors, colDec)
					break
				}
			}
		}

		btnList := []forms.Button{
			{ButtonType: "submit", Title: "Insert", Position: true},
			{ButtonType: "reset", Title: "Clear", Position: false},
		}

		if ok {
			f.Action += "/update"
			btnList[0].Title = "Update"
			colSelect := make([]string, 0, len(columnsTable)-len(priColumns))
		loop_columns:
			for _, colDec := range columnsTable {
				for _, name := range priColumns {
					if name == colDec.Name() {
						continue loop_columns
					}
				}
				colSelect = append(colSelect, colDec.Name())
			}

			err := table.SelectAndRunEach(ctx,
				func(values []interface{}, columns []Column) error {
					ok = false
					for i, col := range columns {
						name := col.Name()
						values[i] = ToStandartColumnValueType(table.Name(), name, id, values[i])
						colDecors = append(colDecors, ToColDev(ctx, DB, patternList, col, values[i]))
					}

					return nil
				},
				ColumnsForSelect(colSelect...),
				WhereForSelect(priColumns...),
				ArgsForSelect(args...),
			)
			if err != nil {
				logs.ErrorLog(err, "")
			}

			// not found record
			if ok {
				ctx.SetStatusCode(fasthttp.StatusNoContent)
				return nil, nil
			}
		} else {
			f.Action += "/put"
		loop_colTables:
			for _, col := range columnsTable {
				if !(col.AutoIncrement() || col.Name() == "id" ||
					strings.Contains(col.Comment(), " (read_only)")) {

					for _, colDec := range colDecors {
						if colDec.Name() == col.Name() {
							continue loop_colTables
						}
					}
					colDecors = append(colDecors, ToColDev(ctx, DB, patternList, col, nil))
				}

			}
		}

		lang, ok := ctx.UserValue(ParamsLang.Name).(string)
		if ok {
			colDecors = append(colDecors, &forms.ColumnDecor{
				Column:      NewStringColumn("lang", "lang", true),
				IsHidden:    true,
				InputType:   "hidden",
				PatternList: nil,
				Value:       lang,
			})
		}

		blocks := []forms.BlockColumns{
			{
				Buttons:     btnList,
				Columns:     colDecors,
				Id:          1,
				Title:       "",
				Description: "",
			},
		}

		_, ok = ctx.UserValue("html").(bool)
		if !ok {
			views.WriteJSONHeaders(ctx)
		}

		if f.Description == "" {
			f.Description = "Input data for " + table.Comment()
		}
		f.WriteRenderForm(
			ctx.Response.BodyWriter(),
			ok, // && isHtml,
			blocks...)

		return nil, nil
	}
}

func GetForeignName(ctx *fasthttp.RequestCtx, DB *DB, col Column, val interface{}) interface{} {
	if val != nil && col.Foreign() != nil {
		table, ok := DB.Tables[col.Foreign().Parent]
		if ok {

			name := GetNameOfTitleColumn(table, ctx.UserValue(ParamsLang.Name))
			if strings.HasPrefix(col.Type(), "_") {
				res := make([]string, 0)
				err := DB.Conn.SelectOneAndScan(ctx,
					&res,
					fmt.Sprintf("select array_agg(%s) from %s where id =ANY($1)", name, table.Name()),
					val,
				)
				if err != nil {
					logs.ErrorLog(err, "%s=%v", name, val)
					return nil
				}
				return res
			}

			res := ""
			err := table.SelectOneAndScan(ctx,
				&res,
				ColumnsForSelect(name),
				WhereForSelect("id"),
				ArgsForSelect(val),
			)
			if err != nil {
				logs.ErrorLog(err, "%s=%v", name, val)
				return nil
			}
			return res
		}
	}

	return nil
}

func GetNameOfTitleColumn(table Table, lang interface{}) string {
	var names = []string{
		"name",
		"title",
		"desc",
		"description",
	}

	for _, name := range names {
		col := table.FindColumn(name)
		if col != nil {
			return GetNameAccordingLang(table, name, lang)
		}
	}
	for _, col := range table.Columns() {
		if col.Name() != "id" {
			return GetNameAccordingLang(table, col.Name(), lang)
		}
	}

	return ""
}

func GetNameAccordingLang(table Table, name string, lang interface{}) string {

	if l, ok := lang.(string); ok && (table.FindColumn(name+"_"+l) != nil) {
		return name + "_" + l
	}

	return name
}

func ToStandartColumnValueType(tableName, colName string, id int32, values interface{}) interface{} {
	// todo- move to dbEngine
	switch v := values.(type) {
	case pgtype.VarcharArray:
		return VarcharArrayToStrings(v.Elements)

	case *pgtype.VarcharArray:
		return VarcharArrayToStrings(v.Elements)

	case pgtype.TextArray:
		return TextArrayToStrings(v.Elements)

	case *pgtype.TextArray:
		return TextArrayToStrings(v.Elements)

	case pgtype.BPCharArray:
		return BPCharArrayToStrings(v.Elements)

	case *pgtype.BPCharArray:
		return BPCharArrayToStrings(v.Elements)

	case pgtype.Int4Array:
		return Int4ArrToStrings(v.Elements)

	case pgtype.Int8Array:
		return Int8ArrToStrings(v.Elements)

	case pgtype.ArrayType:
		str, done := ArrayToStrings(&v)
		if done {
			return str
		}

		return v

	case *pgtype.ArrayType:
		str, done := ArrayToStrings(v)
		if done {
			return str
		}

		return v

	case *pgtype.GenericText:
		logs.DebugLog("%T", v)
		return "genericText: " + v.String

	case pgtype.UntypedTextArray:
		return v.Elements

	case *pgtype.UntypedTextArray:
		return v.Elements

	case []interface{}:
		return UnknownArrayToStrings(v)

	case *pgtype.Bytea, pgtype.Bytea, []uint8:
		return BlobToURL(tableName, colName, id)

	case time.Time:
		return TimeToString(v)

	case *time.Time:
		return TimeToString(*v)

	case nil, string, bool, float32, float64, int32, int64, map[string]string, map[string]interface{}:
		return values

	// case *pgtype.Daterange, pgtype.Daterange:
	//
	// 	d := &DateMarshal{}
	// 	err := d.Set(v)
	// 	if err != nil {
	// 		return fmt.Sprintf("wrong DataMershal %v", err)
	// 	}
	//
	// 	return *d

	default:
		logs.DebugLog("%T", values)
		return values
	}
}

func TimeToString(v time.Time) string {
	return v.Format("2006-01-02")
}

func BlobToURL(tableName string, colName string, id int32) string {
	return fmt.Sprintf("/api/v1/blob/%s?id=%d&name=%s", tableName, id, colName)
}

func ArrayToStrings(v *pgtype.ArrayType) ([]string, bool) {
	src, ok := v.Get().([]interface{})
	if !ok {
		return nil, false
	}

	return UnknownArrayToStrings(src), true
}

func Int4ArrToStrings(src []pgtype.Int4) []int32 {
	str := make([]int32, len(src))
	for i, val := range src {
		str[i] = val.Int
	}

	return str
}

func Int8ArrToStrings(src []pgtype.Int8) []int64 {
	str := make([]int64, len(src))
	for i, val := range src {
		str[i] = val.Int
	}

	return str
}

func UnknownArrayToStrings(src []interface{}) []string {
	str := make([]string, len(src))
	for i, val := range src {
		str[i] = json.Element(val)
	}

	return str
}

func VarcharArrayToStrings(src []pgtype.Varchar) []string {
	str := make([]string, len(src))
	for i, val := range src {
		str[i] = val.String
	}

	return str
}

func TextArrayToStrings(src []pgtype.Text) []string {
	str := make([]string, len(src))
	for i, val := range src {
		str[i] = val.String
	}

	return str
}

func BPCharArrayToStrings(src []pgtype.BPChar) []string {
	str := make([]string, len(src))
	for i, val := range src {
		str[i] = val.String
	}

	return str
}
