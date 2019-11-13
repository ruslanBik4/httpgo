// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package selenuim

import (
	"os/exec"
	"strings"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/tebeka/selenium"
)

type WD struct {
	brouwsers string
	path      string
	wd        selenium.WebDriver
}

func NewWD(brouwsers string, path string) (*WD, error) {
	caps := selenium.Capabilities{"browserName": brouwsers}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			cmd := exec.Command("/Users/ruslan/chromedriver")
			err = cmd.Start()
			if err == nil {
				wd, err = selenium.NewRemote(caps, "http://localhost:9515")
			}
		}
		logs.ErrorLog(err, err.Error())
		if err != nil {
			return nil, err
		}
	}
	return &WD{
		brouwsers: brouwsers,
		path:      path,
		wd:        wd,
	}, nil
}
