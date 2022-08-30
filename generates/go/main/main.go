/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package main

import (
	"flag"
	"go/types"
	"os"
	"path"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/dbEngine/dbEngine/psql"
	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/generates/go/tpl"
	"github.com/ruslanBik4/logs"
)

var (
	fDstPath  = flag.String("dst_path", "./api", "path for generated files")
	fDstGit   = flag.String("dst_pgit", "github.com/ruslanBik4/httpgo", "path for generated files")
	fCfgPath  = flag.String("src_path", "cfg", "path to cfg DB files")
	fOnlyShow = flag.Bool("read_only", true, "only show DB schema")
)

var tables map[string]dbEngine.Table

func init() {
	conn := psql.NewConn(nil, nil, nil)
	dbCfgPath := path.Join(path.Join(*fCfgPath, "DB"), "DB")
	cfgDB := dbEngine.CfgDB{
		Url:       "",
		GetSchema: &struct{}{},
		PathCfg:   &dbCfgPath,
	}
	ctx := context.WithValue(context.Background(), dbEngine.DB_SETTING, cfgDB)
	DB, err := dbEngine.NewDB(ctx, conn)
	if err != nil {
		logs.ErrorLog(err, "dbEngine.NewDB")
		return
	}
	tables = make(map[string]dbEngine.Table)
	for tableName, table := range DB.Tables {
		tables[tableName] = table
	}
}

func main() {
	err := MakeSrcDir(*fDstPath)
	if err != nil {
		logs.ErrorLog(errors.Wrap(err, "NewCreator"))
		return
	}

	routes := make([]string, 0)
	for _, table := range tables {

		r := apis.ApiRoute{
			Desc:           table.Comment(),
			DTO:            nil,
			Fnc:            nil,
			FncAuth:        nil,
			FncIsForbidden: nil,
			TestFncAuth:    nil,
			Method:         apis.POST,
			Multipart:      false,
			NeedAuth:       false,
			OnlyAdmin:      false,
			OnlyLocal:      false,
			WithCors:       false,
			Resp:           nil,
		}
		hasTypeExt, unSafe := false, false
		for _, column := range table.Columns() {
			p := crud.NewDbApiParams(column)

			r.Params = append(r.Params, p.InParam)
			if t, ok := p.Type.(apis.TypeInParam); ok && t.DTO != nil {
				hasTypeExt = hasTypeExt || t.BasicKind < 0

			} else if ok && t.BasicKind == types.UnsafePointer {
				unSafe = true
			}
		}
		d := tpl.NewEndpointTpl(r, table)
		f, err := MakeSrcFile(*fDstPath, table)
		if err != nil {
			logs.ErrorLog(err)
			continue
		}

		tpl.WriteHead(f, hasTypeExt, unSafe)
		defer func() {
			err := f.Close()
			if err != nil {
				logs.ErrorLog(err)
			}
		}()

		d.WriteApisFile(f)
		logs.StatusLog(d.NameRoutes())
		routes = append(routes, d.NameRoutes())
	}
	CreateMainFile(routes)

	err = ImportReact(*fDstPath, "clone", "https://github.com/markovcy/react")
	if err != nil {
		logs.ErrorLog(err, "react")
	}
}

func CreateMainFile(routes []string) {

	dst := path.Join(*fDstPath, "main")
	err := MakeSrcDir(dst)
	if err != nil {
		logs.ErrorLog(errors.Wrap(err, "NewCreator"))
		return
	}

	f, err := os.Create(path.Join(dst, "main.go"))
	if err != nil {
		logs.ErrorLog(err)
		return
	}

	m := tpl.NewApiMain("test", tpl.JWT)
	m.WriteCreateMain(f, path.Join(*fDstGit, *fDstPath), routes)
	defer func() {
		err := f.Close()
		if err != nil {
			logs.ErrorLog(err)
		}
	}()

}
