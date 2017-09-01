// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// read file with seleniumCSS command and run with Chrome
package main

import (
	"github.com/tebeka/selenium"
	"flag"
	"os"
	"log"
	"bytes"
	"time"
	"strconv"
	"strings"
	"errors"
)
type tCommand struct {
	command, param string
}
type ErrFailTest struct {
	token, param string
}

func (err ErrFailTest) Error() string {
	return err.token + " wtih param " + err.param
}
type ErrUnknowCommand error
var (
	result    []selenium.WebElement
 	command   [] tCommand
	values     = map[string] string {}
	fFilename     = flag.String("filename", "test.sln", "file with css selenium rules")
	fScrPath    = flag.String("path_scr", "./", "path to screenshot files")
)
const valPrefix = '@'
//todo: добавить в сценарий переменные, в частности, читать пароли отдельно из файла
//todo: добавить циклы  и ветвления
//todo: доабвить ассерты стандартных тестов ГО

func main() {
	flag.Parse()
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer wd.Quit()
	defer func() {
		saveScreenShoot(wd)
		err := recover()
		if err, ok := err.(error); ok {
			wd.SetAlertText(err.Error())
			log.Print(err)
			time.Sleep(time.Millisecond * 5000)
		}
	}()

	ioReader, err := os.Open(*fFilename)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	stat, err := ioReader.Stat()
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	b := make([]byte, stat.Size())
	n, err := ioReader.Read(b)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}

	log.Print(n)
	b = bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
	slBytes := bytes.Split(b, []byte("\n"))

	for _, line := range slBytes {

		// комментарии и пустые строки пропускаем
		if (len(line) == 0) || isComment(line) {
			continue
		}
		// завершение блока - вылопляем команды для селектора
		if bytes.Index(line, []byte("}") ) > -1 {
			for _, elem := range result {
				for _, val := range command {
					err := wdCommand(val.command, wd, val.param)
					if err != nil {
						if err.Error() == "unknow command" {
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
			command = make([] tCommand, 0)
		} else {
			var token, param string

			tokens = bytes.Split(line, []byte(":"))
			if len(tokens) > 1 {
				token, param = string(tokens[0]), string(tokens[1])
			} else {
				token, param = string(line), ""
			}

			token, param = strings.TrimSpace(token), getParam(param)

			if token[0] == valPrefix {
				values[token[1:]] = param
			} else if command != nil {
				command = append(command, tCommand{ command: token, param: param } )
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
		( bytes.HasPrefix(line, []byte("/*")) && bytes.HasSuffix(line, []byte("*/")) )
}
//
func getParam(param string) string {
	param = strings.TrimSpace(param)
	if (param > "") && (param[0] == valPrefix) {
		if value, ok := values[param[1:]]; ok {
			return value
		} else {
			return ""
		}
	}

	return param
}
// выполнение обзих команд Селениума (не привязанных к елементу страницы)
// возвращает ошибку исполнения, коли такая произойдет
func wdCommand(token string, wd selenium.WebDriver, param string) (err error){
	switch token {
	case "url":
		openURL(wd, param)
	case "getalert":
		err = wd.AcceptAlert()
	case "maximize":
		err = wd.MaximizeWindow("")
	case "screenshot":
		saveScreenShoot(wd)
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
		return errors.New("unknow command")
	}

	return err
}
// открыть URL
func openURL(wd selenium.WebDriver, param string) {
	// Navigate to the simple playground interface.
	if err := wd.Get("http://" + param); err != nil {
		panic(err)
	}
	if status, err := wd.Status(); err != nil {
		panic(err)
	} else {
		log.Printf("%#v", status)
	}
	log.Print("open " + param)
}
// создает скриншот текущего окна браузера и сохраняет его в папке программы
func saveScreenShoot(wd selenium.WebDriver)  {
	img, err := wd.Screenshot()
	if err == nil {
		output, err := os.Create(*fScrPath + time.Now().String() + ".jpg")
		if err == nil {
			_, err = output.Write(img)
			output.Close()
		}
	}
	if err != nil {
		log.Print(err)
	}
}
// find element by selector & panic if error occupiers
func findElementBySelector(wd selenium.WebDriver, token string) []selenium.WebElement {
	wElements, err := wd.FindElements(selenium.ByCSSSelector, token)
	if err != nil {
		log.Print(token)
		panic(err)
	}

	return wElements
}
// webElement select
func getElement(wd selenium.WebDriver, token string) (result []selenium.WebElement){

	if token == "activeElement" {
		if wElem, err := wd.ActiveElement(); err != nil {
			panic(err)
		} else {
			result = append(result, wElem)
		}

		return

	}
	list := strings.Split(token, "::")
	token = list[0]
	if (len(list) > 1) {
		stat := strings.TrimSpace(list[1])
		result = findElementBySelector(wd, token)
		// addition parameters filtering result set
		switch stat {
		case "while":
			wd.SetAlertText("Ждем появления элемента \n" + token + `
			если долго не будет ничего происходит - просто закройте браузер!`)
			for ; len(result) < 1; result = findElementBySelector(wd, token) {
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
		result = findElementBySelector(wd, token)
	}
	log.Print("select ", len(result), " from ", token)

	return
}
// run Comman by webElement
func runCommand(token, param string, wElem selenium.WebElement) bool{
var (	slnCommands  = map[string] func() error {
							"click":  wElem.Click,
							"clear":  wElem.Clear,
							"submit": wElem.Submit,
						}
	slnStat  = map[string] func() (bool, error) {
		"selected": wElem.IsSelected,
		"enabled": wElem.IsEnabled,
		"visible": wElem.IsDisplayed,
	}
	slnText  = map[string] func() (string, error) { "tag": wElem.TagName, "text": wElem.Text, }
	slnCSS   = map[string] func(string) (string, error) {"css": wElem.CSSProperty, "attr": wElem.GetAttribute, }
	//{"move": wElem.MoveTo}
)
	if command, ok := slnCommands[token]; ok {
		if err := command(); err != nil {
			log.Print(token)
			panic(err)
		}
		log.Print("run command ", token)
	} else if command, ok := slnStat[token]; ok {
		if ok, err := command(); err != nil {
			log.Print(token)
			panic(err)
		} else {

			if strings.HasPrefix( param, "!") {
				ok = !ok
				param = param[1:]
			}
			if ok && (param > "") {
				switch param {
				case "fail":
					panic(ErrFailTest{token: token, param: param} )
				case "continue":
					return false
				default:
					log.Print( token + " is ", ok, " unknow param ", param)

				}
			}
		}
	} else if command, ok := slnText[token]; ok {
		if str, err := command(); err != nil {
			log.Print(token)
			panic(err)
		} else {
			log.Print(token +"=" + str)
		}
	} else if command, ok := slnCSS[token]; ok {
		if ok, err := command(param); err != nil {
			log.Print(token)
			panic(err)
		} else {
			log.Print(token + " is " + string(ok))
		}
	} else if token == "pause" {

		pInt, err := strconv.Atoi(param)
		if err != nil {
			panic(err)
		}
		log.Print("pause ", pInt)
		time.Sleep(time.Millisecond * time.Duration(pInt))

	} else  if token == "sendkey" {
		if err := wElem.SendKeys(param); err != nil {
			log.Print("sendkey err", err, param)
			panic(err)
		}
		log.Print(token + ":" + param)
	} else {
		log.Print("unknow command", token)
	}

	return true
}