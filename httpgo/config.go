// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpgo

import (
	"io/ioutil"
	"os"

	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

// CfgHttp has some options for Acceptor work
type CfgHttp struct {
	// list tokens to check requests
	KillSignal int              `yaml:"KillSignal"`
	Server     *fasthttp.Server `yaml:"Server"`
}

// NewCfgHttp create CfgHttp from config file
func NewCfgHttp(filename string) (cfgGlobal *CfgHttp, err error) {

	f, err := os.Open(filename)
	if err == nil {

		defer func() {
			err := f.Close()
			if err != nil {
				logs.ErrorLog(err)
			}
		}()
		var buf []byte
		buf, err = ioutil.ReadAll(f)
		if err == nil {
			err = yaml.Unmarshal(buf, &cfgGlobal)
		}
	}
	if err != nil {
		logs.ErrorLog(err)
		return nil, err
	}

	return
}
