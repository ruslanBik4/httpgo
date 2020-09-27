// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// read file with seleniumCSS command and run with Chrome
package main

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"

	wd "github.com/ruslanBik4/httpgo/tools/selenius"
	"github.com/ruslanBik4/logs"
)

type tCommand struct {
	command, param string
}

// ErrFailTest has not valid parameters
type ErrFailTest struct {
	token, param string
}

func (err ErrFailTest) Error() string {
	return err.token + " with param " + err.param
}

var (
	result    []selenium.WebElement
	command   []tCommand
	values    = map[string]string{}
	fFilename = flag.String("filename", "/Users/ruslan/work/src/github.com/ruslanBik4/httpgo/views/selenium/test.sln", "file with css selenium rules")
	fWDPath   = flag.String("wd_path", "/Users/ruslan/chromedriver", "full path of chrome web-driver")
	fScrPath  = flag.String("path_scr", "./", "path to screenshot files")
)

const valPrefix = '@'

//todo: добавить в сценарий переменные, в частности, читать пароли отдельно из файла
//todo: добавить циклы  и ветвления
//todo: доабвить ассерты стандартных тестов ГО

func main() {
	flag.Parse()

	// Connect to the WebDriver instance running locally.
	wd, err := wd.NewWD("chrome", *fWDPath)
	if err != nil {
		logs.ErrorLog(err)
		return
	}

	b, err := ioutil.ReadFile(*fFilename)
	if err != nil {
		logs.ErrorLog(err, "")
		return
	}

	b = bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
	slBytes := bytes.Split(b, []byte("\n"))

	for _, line := range slBytes {

		// комментарии и пустые строки пропускаем
		if (len(line) == 0) || isComment(line) {
			continue
		}
		// завершение блока - вылопляем команды для селектора
		if bytes.Index(line, []byte("}")) > -1 {
			for _, elem := range result {
				for _, val := range command {
					err := wdCommand(val.command, wd, val.param)
					if err != nil {
						if err.Error() == "unknown command" {
							if !runCommand(val.command, val.param, elem) {
								log.Print(val.command)
							}
						} else {
							log.Print(err)
						}
					}
				}
			}
			result, command = nil, nil
			continue
		}

		tokens := bytes.Split(line, []byte("{"))
		if len(tokens) > 1 {
			// find selector
			result = getElement(wd, string(tokens[0]))
			command = make([]tCommand, 0)
		} else {
			var token, param string

			tokens = bytes.Split(line, []byte(":"))
			if len(tokens) > 1 {
				token, param = string(tokens[0]), string(tokens[1])
				if len(tokens) > 2 {
					param += ":" + string(tokens[2])
				}
			} else {
				token, param = string(line), ""
			}

			token, param = strings.TrimSpace(token), getParam(param)

			if token[0] == valPrefix {
				values[token[1:]] = param
			} else if command != nil {
				command = append(command, tCommand{command: token, param: param})
			} else {
				if err := wdCommand(token, wd, param); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

// line is comment
func isComment(line []byte) bool {
	line = bytes.TrimSpace(line)
	return bytes.HasPrefix(line, []byte("//")) ||
		(bytes.HasPrefix(line, []byte("/*")) && bytes.HasSuffix(line, []byte("*/")))
}

//
func getParam(param string) string {
	param = strings.TrimSpace(param)
	if (param > "") && (param[0] == valPrefix) {
		if value, ok := values[param[1:]]; ok {
			return value
		}

		return ""
	}

	return param
}

// выполнение общих команд Селениума (не привязанных к елементу страницы)
// возвращает ошибку исполнения, коли такая произойдет
func wdCommand(token string, wd *wd.WD, param string) (err error) {
	switch token {
	case "url":
		return openURL(wd, param)
	case "getalert":
		err = wd.AcceptAlert()
	case "maximize":
		err = wd.MaximizeWindow("")
	case "screenshot":
		wd.SaveScreenShoot(*fScrPath + time.Now().String() + ".jpg")
	case "pause":
		pInt, err := strconv.Atoi(param)
		if err != nil {
			return err
		}
		log.Print("pause ", pInt)
		time.Sleep(time.Millisecond * time.Duration(pInt))
	case "executescript":
		result, err := wd.ExecuteScript(param, nil)
		log.Print("script ", result)
		return err
	default:
		return errors.New("unknown command")
	}

	return err
}

// открыть URL
func openURL(wd *wd.WD, url string) error {
	// Navigate to the simple playground interface.
	logs.DebugLog(" %+v", url)
	if err := wd.Get(url); err != nil {
		logs.ErrorLog(err, url)
		return err
	}
	if status, err := wd.Status(); err != nil {
		logs.ErrorLog(err, "")
	} else {
		log.Printf("%#v", status)
	}
	log.Print("open " + url)

	return nil
}

// webElement select
func getElement(wd *wd.WD, token string) (result []selenium.WebElement) {

	if token == "activeElement" {
		if wElem, err := wd.ActiveElement(); err != nil {
			logs.ErrorLog(err, "")
		} else {
			result = append(result, wElem)
		}

		return

	}
	list := strings.Split(token, "::")
	token = list[0]
	if len(list) > 1 {
		stat := strings.TrimSpace(list[1])
		result = wd.FindElementBySelector(token)
		// addition parameters filtering result set
		switch stat {
		case "while":
			wd.SetAlertText("Ждем появления элемента \n" + token + `
			если долго не будет ничего происходит - просто закройте браузер!`)
			for ; len(result) < 1; result = wd.FindElementBySelector(token) {
				log.Print("while for element " + token)
				time.Sleep(time.Millisecond * 100)
			}
			wd.AcceptAlert()
		case "first":
			return result[:1]
		case "last":
			return result[len(result)-1:]
		default:
			temp := make([]selenium.WebElement, 0)
			for i, elem := range result {
				if runCommand(stat, "!continue", elem) {
					temp = append(temp, result[i])
				}
			}
			result = temp
		}
	} else {
		result = wd.FindElementBySelector(token)
	}
	log.Print("select ", len(result), " from ", token)

	return
}

// run Comman by webElement
func runCommand(token, param string, wElem selenium.WebElement) bool {
	var (
		slnCommands = map[string]func() error{
			"click":  wElem.Click,
			"clear":  wElem.Clear,
			"submit": wElem.Submit,
		}
		slnStat = map[string]func() (bool, error){
			"selected": wElem.IsSelected,
			"enabled":  wElem.IsEnabled,
			"visible":  wElem.IsDisplayed,
		}
		slnText = map[string]func() (string, error){"tag": wElem.TagName, "text": wElem.Text}
		slnCSS  = map[string]func(string) (string, error){"css": wElem.CSSProperty, "attr": wElem.GetAttribute}
		//{"move": wElem.MoveTo}
	)
	if command, ok := slnCommands[token]; ok {
		if err := command(); err != nil {
			log.Print(token)
			logs.ErrorLog(err, "")
		}
		log.Print("run command ", token)
	} else if command, ok := slnStat[token]; ok {
		if ok, err := command(); err != nil {
			log.Print(token)
			logs.ErrorLog(err, "")
		} else {

			if strings.HasPrefix(param, "!") {
				ok = !ok
				param = param[1:]
			}
			if ok && (param > "") {
				switch param {
				case "fail":
					panic(ErrFailTest{token: token, param: param})
				case "continue":
					return false
				default:
					log.Print(token+" is ", ok, " unknow param ", param)

				}
			}
		}
	} else if command, ok := slnText[token]; ok {
		if str, err := command(); err != nil {
			log.Print(token)
			logs.ErrorLog(err, "")
		} else {
			log.Print(token + "=" + str)
		}
	} else if command, ok := slnCSS[token]; ok {
		if ok, err := command(param); err != nil {
			log.Print(token)
			logs.ErrorLog(err, "")
		} else {
			log.Print(token + " is " + string(ok))
		}
	} else if token == "pause" {

		pInt, err := strconv.Atoi(param)
		if err != nil {
			logs.ErrorLog(err, "")
		}
		log.Print("pause ", pInt)
		time.Sleep(time.Millisecond * time.Duration(pInt))

	} else if token == "sendkey" {
		if err := wElem.SendKeys(param); err != nil {
			log.Print("sendkey err", err, param)
			logs.ErrorLog(err, "")
		}
		log.Print(token + ":" + param)
	} else if token == "active" {
		p, _ := wElem.LocationInView()
		if err := wElem.MoveTo(p.X, p.Y); err != nil {
			log.Print("moveto err", err, param)
			logs.ErrorLog(err, "")
		}
		log.Print(token + ":" + param)
	} else {
		log.Print("unknown command", token)
	}

	return true
}
