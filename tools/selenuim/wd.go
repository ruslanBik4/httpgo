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
	browsers string
	path     string
	wd       selenium.WebDriver
}

func NewWD(browsers string, path string) (*WD, error) {
	caps := selenium.Capabilities{"browserName": browsers}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			cmd := exec.Command("/Users/ruslan/chromedriver")
			err = cmd.Start()
			if err == nil {
				wd, err = selenium.NewRemote(caps, "http://localhost:9515")
			}
		}

		if err != nil {
			logs.ErrorLog(err, err.Error())
			return nil, err
		}
	}
	return &WD{
		browsers: browsers,
		path:     path,
		wd:       wd,
	}, nil
}

// find element by selector & panic if error occupiers
func (wd *WD) findElementBySelector(token string) []selenium.WebElement {
	wElements, err := wd.wd.FindElements(selenium.ByCSSSelector, token)
	if err != nil {
		logs.ErrorLog(err, token)
		return nil
	}

	logs.DebugLog(" %+v", wElements[0])
	return wElements
}
