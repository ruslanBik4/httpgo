// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seleniгs

import (
	"log"
	"os"
	"os/exec"
	"strings"

	. "github.com/tebeka/selenium"

	"github.com/ruslanBik4/httpgo/logs"
)

type WD struct {
	browsers string
	path     string
	wd       WebDriver
}

func NewWD(browsers string, path string) (*WD, error) {
	caps := Capabilities{"browserName": browsers}
	wd, err := NewRemote(caps, "http://localhost:9515")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			cmd := exec.Command(path)
			err = cmd.Start()
			if err == nil {
				wd, err = NewRemote(caps, "http://localhost:9515")
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

// создает скриншот текущего окна браузера и сохраняет его в папке программы
func (wd *WD) SetAlertText(text string) error {
	return wd.wd.SetAlertText(text)
}

// Status returns various pieces of information about the server environment.
func (wd *WD) Status() (*Status, error) {
	return wd.wd.Status()
}

// Get navigates the browser to the provided URL.
func (wd *WD) Get(url string) error {
	return wd.wd.Get(url)
}

// создает скриншот текущего окна браузера и сохраняет его в папке программы
func (wd *WD) SaveScreenShoot(filename string) {
	img, err := wd.wd.Screenshot()
	if err == nil {
		var output *os.File
		output, err = os.Create(filename)
		if err == nil {
			defer output.Close()
			_, err = output.Write(img)
		}
	}
	if err != nil {
		log.Print(err)
	}
}

func (wd *WD) AcceptAlert() error {
	return wd.wd.AcceptAlert()
}

func (wd *WD) ActiveElement() (WebElement, error) {
	return wd.wd.ActiveElement()
}

func (wd *WD) ExecuteScript(script string, args []interface{}) (interface{}, error) {
	return wd.wd.ExecuteScript(script, args)
}

func (wd *WD) MaximizeWindow(title string) error {
	return wd.wd.MaximizeWindow(title)
}

// find element by selector & panic if error occupiers
func (wd *WD) FindElementBySelector(token string) []WebElement {
	wElements, err := wd.wd.FindElements(ByCSSSelector, token)
	if err != nil {
		logs.ErrorLog(err, token)
		return nil
	}

	return wElements
}
