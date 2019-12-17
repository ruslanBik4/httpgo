// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// инициализация и запуск веб-сервера, подключение основных хандлеров
package main

import (
	"flag"
	"fmt"
	"go/types"
	"io/ioutil"
	"net"
	_ "net/http/pprof"
	"os"
	// "syscall"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/valyala/fasthttp"

	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/httpGo"
	"github.com/ruslanBik4/httpgo/logs"
	_ "github.com/ruslanBik4/httpgo/models/api/v1"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/db/qb"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
)

//go:generate qtc -dir=views/templates

const fpmSocket = "/var/run/php5-fpm.sock"

var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
	// ErrIndexMissingSplit describes an index configuration error.
	//ErrIndexMissingSplit = errors.New("configured index file(s) must include split value")

	cacheMu sync.RWMutex
	cache   = map[string][]byte{}
	routes  = ApiRoutes{
		"moreParams": {
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
		// "/godoc/":        handlerGoDoc,
		// "/recache":       handlerRecache,
		// "/update/":       handleUpdate,
		// "/test/":         handleTest,
		// "/api/firebird/": HandleFirebird,
		// "/fonts/":        fonts.HandleGetFont,
		// "/query/":        db.HandlerDBQuery,
		// "/menu/":         handlerMenu,
		// "/show/forms/":   handlerForms,
		// "/user/signup/":  users.HandlerSignUp,
		// "/user/signin/":  users.HandlerSignIn,
		// "/user/signout/": users.HandlerSignOut,
		// "/user/active/":  users.HandlerActivateUser,
		// "/user/profile/": users.HandlerProfile,
		// //"/user/oauth/":    users.HandlerQauth2,
		// "/user/GoogleCallback/": users.HandleGoogleCallback,
		// "/components/":          handlerComponents,
	}
)

func testValue(ctx *RequestCtx) interface{} {
	return ctx.Method()
}

// DefaultHandler работа по умолчанию - кеширования общих файлов в частности, обработчики для php-fpm & php
type DefaultHandler struct {
	fpm       *system.FCGI
	php       *system.FCGI
	cache     []string
	whitelist []string
}

// NewDefaultHandler create default handler for read static files
func NewDefaultHandler() *DefaultHandler {
	handler := &DefaultHandler{
		fpm: system.NewFPM(fpmSocket),
		php: system.NewPHP(*fWeb, "index.php", *fPort, fpmSocket),
		cache: []string{
			".svg", ".css", ".js", ".map", ".ico",
		},
		whitelist: []string{
			".jpg", ".jpeg", ".png", ".gif", ".ttf", ".pdf", ".json", ".htm", ".html", ".txt", ".mp4",
		},
	}
	// read from flags
	cacheExt := *fCache
	p := strings.Index(cacheExt, ";")
	for p > 0 {

		handler.cache = append(handler.cache, cacheExt[:p])
		cacheExt = cacheExt[p:]
		p = strings.Index(cacheExt, ";")
	}
	return handler
}
func (h *DefaultHandler) toCache(ext string) bool {
	for _, name := range h.cache {
		if ext == name {
			return true
		}
	}
	return false
}
func (h *DefaultHandler) toServe(ext string) bool {
	for _, name := range h.whitelist {
		if ext == name {
			return true
		}
	}
	return false
}
func (h *DefaultHandler) ServeHTTP(ctx *RequestCtx) {

	// defer system.Catch(ctx)

	switch ctx.URI().String() {
	case "/":
		ctx.Redirect("/customer/", StatusTemporaryRedirect)
		return

		//p := &pages.IndexPageBody{Title: "Главная страница"}
		////для авторизованного пользователя - сразу показать его данные на странице
		//p.Content = fmt.Sprintf("<script>afterLogin({login:'%d',sex:'0'})</script>", userID)
		//var menu db.MenuItems
		//menu.GetMenu("indexTop")
		//
		//p.TopMenu = make(map[string]string, len(menu.Items))
		//for _, item := range menu.Items {
		//	p.TopMenu[item.Title] = "/menu/" + item.Name + "/"
		//
		//}
		//views.RenderTemplate(ctx, "index", p)
		// спецвойска
	case "/polymer.html":
		ServeFile(ctx, filepath.Join(*fSystem, "views/components/polymer/polymer.html"))
	case "/polymer-mini.html":
		ServeFile(ctx, filepath.Join(*fSystem, "views/components/polymer/polymer-mini.html"))
	case "/polymer-micro.html":
		ServeFile(ctx, filepath.Join(*fSystem, "views/components/polymer/polymer-micro.html"))
	case "/status", "/ping", "/pong":
		h.fpm.ServeHTTP(ctx)
	default:
		filename := strings.TrimLeft(ctx.URI().String(), "/")
		ext := filepath.Ext(filename)

		if strings.HasPrefix(ext, ".php") {
			h.php.ServeHTTP(ctx)
		} else if h.toCache(ext) {
			serveAndCache(filename, ctx)
		} else if h.toServe(ext) {
			ServeFile(ctx, filepath.Join(*fWeb, filename))
		} else if fileName := filepath.Join(*fWeb, filename); ext == "" {
			ServeFile(ctx, fileName)
		} else {
			h.php.ServeHTTP(ctx)
		}
	}
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

const _24K = (1 << 10) * 24

func handleUpdate(ctx *RequestCtx) {

}

func handlerForms(ctx *RequestCtx) {
	// views.RenderTemplate(ctx, r.FormValue("name")+"Form", &pages.IndexPageBody{Title: r.FormValue("email")})
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

		p := &layouts.MenuOwnerBody{Title: idMenu, TopMenu: make(map[string]*layouts.ItemMenu, 0)}

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &layouts.ItemMenu{Link: "/menu/" + item.Name + "/"}

		}

		// return into parent menu if he occurent
		if menu.Self.ParentID > 0 {
			p.TopMenu["< на уровень выше"] = &layouts.ItemMenu{Link: fmt.Sprintf("/menu/%d/", menu.Self.ParentID)}
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
		// views.RenderInternalError(w, err)
	} else {
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		} else {
			return output, nil
		}
		ctx.Redirect("http://localhost:6060", StatusPermanentRedirect)
	}

	return nil, nil
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

	logs.StatusLog("Static files found in " + *fWeb)
	logs.StatusLog("System files found in " + *fSystem)

	defer func() {
		logs.StatusLog("Server correct shutdown")
	}()

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
// func HandleVersion(ctx *fasthttp.RequestCtx) (interface{}, error) {
//
// 	return fmt.Sprintf("analytics-%s Version: %s, Build Time: %s", Branch, Version, Build), nil
// }
