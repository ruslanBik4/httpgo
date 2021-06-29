// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpGo

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"

	"github.com/ruslanBik4/logs"
)

// CfgHttp has some options for Acceptor work
type CfgHttp struct {
	*AccessConf `yaml:"Access"json:"access,omitempty"`
	fileCfg     string
	Domains     map[string]string `yaml:"Domains,omitempty"json:"Domains,omitempty"`
	KillSignal  int               `yaml:"KillSignal"json:"KillSignal,omitempty"`
	Server      *fasthttp.Server  `json:"-"yaml:"Server,omitempty"`
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
