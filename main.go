// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// инициализация и запуск веб-сервера, подключение основных хандлеров
package main

import (
	"context"
	"flag"
	"fmt"
	"go/types"
	"io/ioutil"
	"net"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"

	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/dbEngine/dbEngine/psql"

	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/httpGo"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/telegrambot"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/forms"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/views/templates/tables"
	"github.com/ruslanBik4/logs"
)

//go:generate qtc -dir=views/templates

const fpmSocket = "/var/run/php5-fpm.sock"
const ShowVersion = "/api/version()"

var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
	// ErrIndexMissingSplit describes an index configuration error.
	//ErrIndexMissingSplit = errors.New("configured index file(s) must include split value")

	cacheMu sync.RWMutex
	cache   = map[string][]byte{}
	routes  = MapRoutes{
		GET: {
			"/": {
				Fnc: func(ctx *RequestCtx) (interface{}, error) {
					body := &pages.IndexPageBody{
						TopMenu: layouts.Menu{
							{Link: "Search", Label: "/form/search/"},
							{Link: "View", Label: "/test/view/"},
						},
						Title: "Index page of test server",
					}
					views.RenderHTMLPage(ctx, body.WriteIndexHTML)

					return nil, nil
				},
			},
			ShowVersion: {
				Fnc:  HandleVersion,
				Desc: "view version server",
			},
			"/test/forms/": {
				Fnc: func(ctx *RequestCtx) (interface{}, error) {
					s := make([]*forms.ColumnDecor, 0)
					s = append(s, &forms.ColumnDecor{Column: dbEngine.NewStringColumn("test 1", "test 1", false)})
					s = append(s, &forms.ColumnDecor{Column: dbEngine.NewStringColumn("phone", "test 2", false)})
					s = append(s, &forms.ColumnDecor{Column: dbEngine.NewStringColumn("req", "required", true)})
					s = append(s, &forms.ColumnDecor{Column: dbEngine.NewNumberColumn("number", "number required", true)})
					p := psql.NewColumnPone("psql", "psql column", 0)
					p.UdtName = "_int4"
					s = append(s, &forms.ColumnDecor{Column: p})

					p1 := psql.NewColumnPone("psql bool", "psql bool column", 0)
					p1.UdtName = "bool"
					s = append(s, &forms.ColumnDecor{Column: p1})

					p2 := psql.NewColumnPone("email array", "psql [] string column", 0)
					p2.UdtName = "_varchar"

					decor := &forms.ColumnDecor{
						Column:      p2,
						PatternName: `\d\s\w{3}\d`,
						Value:       []string{"decor1", "decor2"},
					}
					s = append(s, decor)
					f := forms.FormField{
						Title:       "test form",
						Action:      "/test/forms/post",
						Method:      "POST",
						Description: "",
					}

					blocks := forms.BlockColumns{
						Columns:     s,
						Id:          0,
						Title:       "first",
						Description: "test block",
					}

					if ctx.UserValue(ChildRoutePath) == "html" {
						views.WriteHeadersHTML(ctx)
						f.WriteFormHTML(
							ctx.Response.BodyWriter(),
							blocks)

					} else {
						f.WriteFormJSON(
							ctx.Response.BodyWriter(),
							blocks)
					}

					return nil, nil
				},
			},
		},
		POST: {
			"/api/tb/group_count/": {
				Fnc: func(ctx *RequestCtx) (i interface{}, err error) {
					b, err := telegrambot.NewTelegramBotFromEnv()
					if err != nil {
						return nil, errors.Wrap(err, "NewTelegramBot")
					}

					err = b.GetChatMemberCount(b.ChatID)
					if err != nil {
						return nil, errors.Wrap(err, "GetChatMemberCount")
					}

					return b.GetResult(), nil
				},
				Desc:      "test route",
				Method:    POST,
				Multipart: true,
			},
			"/moreParams/": {
				Desc:      "test route",
				Method:    POST,
				Multipart: true,
				NeedAuth:  true,
				Params: []InParam{
					{
						Name: "globalTags",
						Desc: "data of dashboard -> filter 'Global Tags'",
						Req:  false,
						Type: NewSliceTypeInParam(types.Int32),
					},
					{
						Name:     "group",
						Desc:     "type grouping data of ohlc (month, week, day)",
						Req:      true,
						Type:     NewTypeInParam(types.String),
						DefValue: "day",
					},
					{
						Name:     "account",
						Desc:     "account numbers to filter data",
						Req:      true,
						Type:     NewTypeInParam(types.Bool),
						DefValue: testValue,
					},
				},
				Resp: struct {
					Hours map[string]float64
				}{
					map[string]float64{"after 16:00": 15.2,
						"13:30 - 15:30": 1570.86,
						"9:30 - 9:50":   1672.54,
					},
				},
			},
			"/test/forms/post": {
				Fnc: func(ctx *RequestCtx) (interface{}, error) {

					return ctx.UserValue(MultiPartParams), nil
				},
				Method:    POST,
				Multipart: true,
			},
		},
	}
)

func testValue(ctx *RequestCtx) interface{} {
	return ctx.Method()
}

func handlerComponents(ctx *RequestCtx) {

	filename := strings.TrimLeft(ctx.URI().String(), "/")

	ServeFile(ctx, filepath.Join(*fSystem+"/views", filename))

}

// считываем файлы типа css/js etc в память и потом отдаем из нее
func setCache(path string, data []byte) {
	cacheMu.Lock()
	cache[path] = data
	cacheMu.Unlock()
}
func getCache(path string) ([]byte, bool) {
	cacheMu.RLock()
	data, ok := cache[path]
	cacheMu.RUnlock()
	return data, ok
}
func emptyCache() {
	cacheMu.RLock()
	cache = make(map[string][]byte, 0)
	cacheMu.RUnlock()

}
func serveAndCache(filename string, ctx *RequestCtx) {
	keyName := path.Base(filename)

	data, ok := getCache(keyName)
	if ok {
		// if found header no-cache - reread resource
		cache := ctx.Request.Header.Peek("Cache-Control")
		if len(cache) > 0 {
			// for _, val := range cache {
			// 	if val == "no-cache" {
			// 		ok = false
			// 		break
			// 	}
			// }
		}
	}
	if !ok {
		data, err := ioutil.ReadFile(filepath.Join(*fSystem, filename))
		if os.IsNotExist(err) {
			data, err = ioutil.ReadFile(filepath.Join(*fWeb, filename))
		}
		// if system.WriteError(w, err) {
		// 	return
		// }
		setCache(keyName, data)
		logs.DebugLog("recache file", filename)
	}
	logs.DebugLog(" %+v", data)
	// ServeContent(ctx, filename, time.Time{}, bytes.NewReader(data))
}

func sockCatch() {
	err := recover()
	logs.ErrorLog(err.(error))
}

// func isAJAXRequest(r *Request) bool {
// 	return len(r.Header["X-Requested-With"]) > 0
// }
func handlerMenu(ctx *RequestCtx) (interface{}, error) {

	// userID := users.IsLogin(r)
	// resultID, err := strconv.Atoi(userID)
	// if err != nil || !admin.GetUserPermissionForPageByUserId(resultID, ctx.URI().String(), "View") {
	// 	views.RenderNoPermissionPage(w)
	// 	return
	// }
	var menu db.MenuItems

	idx := strings.LastIndex(ctx.URI().String(), "menu/") + 5
	idMenu := ctx.URI().String()[idx : len(ctx.URI().String())-1]

	//отдаем полный список меню для фронтового фреймворка
	if idMenu == "all" {
		qBuilder := qb.CreateEmpty()
		qBuilder.AddTable("", "menu_items")

		arrJSON, err := qBuilder.SelectToMultidimension()
		if err != nil {
			return nil, err
		}

		return arrJSON, nil
	}

	var catalog, content string
	// отрисовка меню страницы
	if menu.GetMenu(idMenu) > 0 {

		p := &layouts.MenuOwnerBody{Title: idMenu}

		for _, item := range menu.Items {
			p.TopMenu = append(p.TopMenu, layouts.ItemMenu{Link: "/menu/" + item.Name + "/", Content: item.Title})

		}

		// return into parent menu if he occurent
		if menu.Self.ParentID > 0 {
			p.TopMenu = append(p.TopMenu, layouts.ItemMenu{
				Link:    fmt.Sprintf("/menu/%d/", menu.Self.ParentID),
				Content: "< на уровень выше",
			})
		}
		catalog = p.MenuOwner()
	}
	//для отрисовки контента страницы
	if menu.Self.Link > "" {
		content = fmt.Sprintf("<div class='autoload' data-href='%s'></div>", menu.Self.Link)
	}
	return catalog + content, nil
}

// считываю части из папки
func cacheWalk(path string, info os.FileInfo, err error) error {
	if (err != nil) || ((info != nil) && info.IsDir()) {
		//log.Println(err, info)
		return nil
	}
	ext := filepath.Ext(path)
	switch ext {
	case ".php":
		return nil
	}

	keyName := filepath.Base(path)
	if _, ok := getCache(keyName); !ok {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			logs.ErrorLog(err)
			return err
		}
		setCache(keyName, data)
		//log.Println(keyName)
	}
	return nil
}
func cacheFiles() {
	filepath.Walk(filepath.Join(*fSystem, "js"), cacheWalk)
	filepath.Walk(filepath.Join(*fSystem, "css"), cacheWalk)

	cachePath := *fChePath
	p := strings.Index(cachePath, ";")
	for p > 0 {

		filepath.Walk(filepath.Join(*fWeb, cachePath[:p]), cacheWalk)
		cachePath = cachePath[p+1:]
		p = strings.Index(cachePath, ";")
	}
	filepath.Walk(filepath.Join(*fWeb, cachePath), cacheWalk)
}

// show doc
// @/godoc/
func handlerGoDoc(ctx *RequestCtx) (interface{}, error) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("godoc", "http=:6060", "index")
	cmd.Dir = ServerConfig.SystemPath()

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
	// }
	// ctx.Redirect("http://localhost:6060", StatusPermanentRedirect)
	//
	// return nil, nil
}

// rereads files to cache directive
func handlerRecache(ctx *RequestCtx) (interface{}, error) {

	emptyCache()
	cacheFiles()
	return "recache succesfull!", nil
}

var (
	fPort    = flag.String("port", ":80", "host address to listen on")
	fSystem  = flag.String("path", "./", "path to static files")
	fCfgPath = flag.String("config path", "config", "path to cfg files")
	fWeb     = flag.String("web", "/home/web/", "path to web files")
	fSession = flag.String("sessionPath", "/var/lib/php/session", "path to store sessions data")
	fCache   = flag.String("cacheFileExt", `.eot;.ttf;.woff;.woff2;.otf;`, "file extensions for caching HTTPGO")
	fChePath = flag.String("cachePath", "css;js;fonts;images", "path to cached files")
)

func init() {
	flag.Parse()
	// ServerConfig := server.GetServerConfig()
	// if err := ServerConfig.Init(fSystem, fWeb, fSession); err != nil {
	// 	logs.ErrorLog(err)
	// }

	// MongoConfig := server.GetMongodConfig()
	// if err := MongoConfig.Init(fSystem, fWeb, fSession); err != nil {
	// 	logs.ErrorLog(err)
	// }
	// logs.StatusLog("Server starting", ServerConfig.StartTime)
	// services.InitServices()
}

func main() {
	// users.SetSessionPath(*fSession)
	// go cacheFiles()
	//
	// fonts.GetPath(fWeb)

	conn := psql.NewConn(nil, nil, nil)
	ctx := context.WithValue(context.Background(), "dbURL", "")
	ctx = context.WithValue(ctx, "fillSchema", true)
	ctx = context.WithValue(ctx, "migration", "table")
	db, err := dbEngine.NewDB(ctx, conn)
	if err != nil {
		logs.ErrorLog(err, "")
		return
	}

	for key, table := range db.Tables {
		if key != table.Name() {
			logs.StatusLog(key, table.Name())
		}

		t := ApiRoutes{"/test/view/": tables.ViewRoute("", table, db)}

		bd := routes.AddRoutes(t)
		if len(bd) == 0 {
			break
		}

		logs.DebugLog(bd)
	}

	logs.StatusLog("Static files found in " + *fWeb)
	logs.StatusLog("System files found in " + *fSystem)

	defer func() {
		logs.StatusLog("Server correct shutdown")
	}()

	if !strings.HasPrefix(*fPort, ":") {
		*fPort = ":" + *fPort
	}
	listener, err := net.Listen("tcp", *fPort)
	if err != nil {
		logs.Fatal(err)
	}

	// badRoutings := AddRoutes(routes)
	// if len(badRoutings) > 0 {
	// 	logs.ErrorLog(apis.ErrRouteForbidden, badRoutings)
	// }

	ctxApis := NewCtxApis(0)

	apis := NewApis(ctxApis, routes, nil)
	logs.StatusLog(os.Getwd())
	cfg, err := httpGo.NewCfgHttp(path.Join(*fSystem, *fCfgPath, "httpgo.yml"))
	if err != nil {
		// not work without correct config
		logs.Fatal(err)
	}
	httpServer := httpGo.NewHttpgo(cfg, listener, apis)
	// services.InitServices("db_schema", "mail")

	err = httpServer.Run(
		false, "", "")
	// path.Join(*fSystem, *fCfgPath, "server.crt"),
	// path.Join(*fSystem, *fCfgPath, "server.key"))

	if err != nil {
		logs.ErrorLog(err)
	}

}

// version
var (
	Version string
	Build   string
	Branch  string
)

// HandleLogServer show status httpgo
// @/api/version/
func HandleVersion(ctx *RequestCtx) (interface{}, error) {

	return fmt.Sprintf("analytics-%s Version: %s, Build Time: %s", Branch, Version, Build), nil
}
