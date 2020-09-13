// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package system

import (
	"path/filepath"
	"strings"
)

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
