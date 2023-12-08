/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"os"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/ruslanBik4/logs"
)

// CfgHttp has some options for Acceptor work
type CfgHttp struct {
	*AccessConf `yaml:"Access" json:"access,omitempty"`
	fileCfg     string
	Domains     map[string]string `yaml:"Domains,omitempty" json:"Domains,omitempty"`
	KillSignal  int               `yaml:"KillSignal" json:"KillSignal,omitempty"`
	Server      *fasthttp.Server  `yaml:"Server,omitempty" json:"-"`
}

// NewCfgHttp create CfgHttp from config file
func NewCfgHttp(filename string) (cfgGlobal *CfgHttp, err error) {

	cfgGlobal, err = loadCfg(filename)
	if err != nil {
		return
	}

	if cfgGlobal == nil {
		return nil, errors.New("cfg httpgo is nil")
	}

	cfgGlobal.fileCfg = filename

	return
}

func (cfg *CfgHttp) Reload() (any, error) {
	cfg, err := loadCfg(cfg.fileCfg)
	if err != nil {
		return nil, err
	}

	cfg.AccessConf = cfg.AccessConf

	return cfg, nil
}

func loadCfg(filename string) (cfg *CfgHttp, err error) {
	f, err := os.Open(filename)
	if err != nil {
		logs.ErrorLog(err)
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			logs.ErrorLog(err, "close file '%s' failed", filename)
		}
	}()

	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		logs.ErrorLog(err, "decoding error")
		return nil, err
	}

	return
}
