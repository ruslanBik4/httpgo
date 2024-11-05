/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"strings"

	"github.com/valyala/fasthttp"
)

type Access interface {
	Allow(ctx *fasthttp.RequestCtx, addr string) bool
}

type AccessConf struct {
	ChkConn       bool              `yaml:"ChkConn"`
	AllowIP       []string          `yaml:"Allow"`
	DenyIP        []string          `yaml:"Deny"`
	Mess          string            `yaml:"Mess"`
	AllowRoute    []string          `yaml:"AllowRoute"`
	DenyRoute     []string          `yaml:"DenyRoute"`
	QuickResponse map[string]string `yaml:"QuickResponse"`
}

func (cfg *AccessConf) isAllowRoute(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	for _, str := range cfg.AllowRoute {
		if strings.HasPrefix(path, str) ||
			((strings.Index(str, "?") > -1) &&
				strings.HasPrefix(path+"?"+ctx.QueryArgs().String(), str)) {

			return true
		}
	}

	return false
}

func (cfg *AccessConf) isDenyRoute(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	for _, str := range cfg.DenyRoute {
		if strings.HasPrefix(path, str) ||
			((strings.Index(str, "?") > -1) &&
				strings.HasPrefix(path+"?"+ctx.QueryArgs().String(), str)) {

			return true
		}
	}

	return false
}

func (cfg *AccessConf) Allow(ctx *fasthttp.RequestCtx, addr string) bool {

	return cfg.isAllowRoute(ctx) || cfg.isAllowIP(addr)
}

func (cfg *AccessConf) Deny(ctx *fasthttp.RequestCtx, addr string) bool {

	return cfg.isDenyRoute(ctx) || cfg.isDenyIP(addr)
}

func (cfg *AccessConf) isAllowIP(addr string) bool {
	for _, str := range cfg.AllowIP {
		if strings.HasPrefix(addr, str) {
			return true
		}
	}

	return false
}

func (cfg *AccessConf) isDenyIP(addr string) bool {
	for _, str := range cfg.DenyIP {
		if strings.HasPrefix(addr, str) {
			return true
		}
	}

	return false
}

func (cfg *AccessConf) IsAccess() bool {
	return cfg != nil && (len(cfg.AllowIP) > 0 || len(cfg.DenyIP) > 0)
}
