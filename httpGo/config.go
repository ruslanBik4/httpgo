// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpGo

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"

	"github.com/ruslanBik4/logs"
)

type AccessConf struct {
	ChkConn    bool     `yaml:"ChkConn"`
	AllowIP    []string `yaml:"Allow"`
	DenyIP     []string `yaml:"Deny"`
	Mess       string   `yaml:"Mess"`
	AllowRoute []string `yaml:"AllowRoute"`
	DenyRoute  []string `yaml:"DenyRoute"`
}

// CfgHttp has some options for Acceptor work
type CfgHttp struct {
	fileCfg string
	Domains map[string]string `yaml:"Domains"`
	// list tokens to check requests
	KillSignal int              `yaml:"KillSignal"`
	Server     *fasthttp.Server `yaml:"Server"`
	Access     AccessConf       `yaml:"Access"`
}

// NewCfgHttp create CfgHttp from config file
func NewCfgHttp(filename string) (cfgGlobal *CfgHttp, err error) {

	var buf []byte

	buf, err = ioutil.ReadFile(filename)
	if err != nil {
		logs.ErrorLog(err)
		return nil, err
	}

	err = yaml.Unmarshal(buf, &cfgGlobal)
	if err != nil {
		return nil, err
	}

	if cfgGlobal == nil {
		return nil, errors.New("cfg httpgo is nil")
	}

	cfgGlobal.fileCfg = filename

	return
}

func (cfg *CfgHttp) isAllowRoute(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	for _, str := range cfg.Access.AllowRoute {
		if strings.HasPrefix(path, str) ||
			((strings.Index(str, "?") > -1) &&
				strings.HasPrefix(path+"?"+ctx.QueryArgs().String(), str)) {

			return true
		}
	}

	return false
}

func (cfg *CfgHttp) isDenyRoute(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	for _, str := range cfg.Access.DenyRoute {
		if strings.HasPrefix(path, str) ||
			((strings.Index(str, "?") > -1) &&
				strings.HasPrefix(path+"?"+ctx.QueryArgs().String(), str)) {

			return true
		}
	}

	return false
}

func (cfg *CfgHttp) Allow(ctx *fasthttp.RequestCtx, addr string) bool {

	if len(cfg.Access.DenyRoute) > 0 {
		if !cfg.isDenyRoute(ctx) {
			return true
		}
	}

	return cfg.isAllowIP(addr)
}

func (cfg *CfgHttp) isAllowIP(addr string) bool {
	for _, str := range cfg.Access.AllowIP {
		if strings.HasPrefix(addr, str) {
			return true
		}
	}

	return false
}

func (cfg *CfgHttp) Deny(ctx *fasthttp.RequestCtx, addr string) bool {

	if len(cfg.Access.AllowRoute) > 0 {
		if cfg.isAllowRoute(ctx) {
			return false
		}
	}

	return cfg.isDenyIP(addr)
}

func (cfg *CfgHttp) isDenyIP(addr string) bool {
	for _, str := range cfg.Access.DenyIP {
		if strings.HasPrefix(addr, str) {
			return true
		}
	}

	return false
}

func (cfg *CfgHttp) IsAccess() bool {
	return len(cfg.Access.AllowIP) > 0 || len(cfg.Access.DenyIP) > 0
}

func (cfg *CfgHttp) Reload() (interface{}, error) {

	buf, err := ioutil.ReadFile(cfg.fileCfg)
	if err != nil {
		return nil, err
	}

	var cfgGlobal CfgHttp
	err = yaml.Unmarshal(buf, &cfgGlobal)
	if err != nil {
		return nil, err
	}

	cfg.Access = cfgGlobal.Access

	return cfg.Access, nil
}
