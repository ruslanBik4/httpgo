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
var 	fFilename    = flag.String("filename", "test.css", "file with css selenium rules")
type tCommand struct {
	command, param string
}
var (
	wElements []selenium.WebElement
 	command [] tCommand
)

func main() {
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer wd.Quit()
	defer func() {
		err := recover()
		if err, ok := err.(error); ok {
			time.Sleep(time.Millisecond * 5000)
			log.Print(err)
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
		if bytes.Index(line, []byte("}") ) > -1 {
			for _, elem := range wElements {
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
			wElements, command = nil, nil
			continue
		}

		tokens := bytes.Split(line, []byte("{"))
		if len(tokens) > 1 {
			// find selector
			wElements = getElement(wd, string(tokens[0]))
			command = make([] tCommand, 0)
		} else {
			var token, param string

			tokens = bytes.Split(line, []byte(":"))
			if len(tokens) > 1 {
				token, param = string(tokens[0]), string(tokens[1])
			} else {
				token, param = string(line), ""
			}

			token, param = strings.TrimSpace(token), strings.TrimSpace(param)

			if command != nil {
				command = append(command, tCommand{ command: token, param: param } )
			} else {
				err := wdCommand(token, wd, param)
				log.Print(err)
			}
		}
	}
}
func wdCommand(token string, wd selenium.WebDriver, param string) (err error){
	switch token {
	case "url":
		openURL(wd, param)
	case "getalert":
		err = wd.AcceptAlert()
	case "maximize":
		err = wd.MaximizeWindow("")
	case "screenshot":
		wd.Screenshot()
	case "executescript":
		result, err := wd.ExecuteScript(param, nil)
		log.Print("script ", result)
		return err
	default:
		return errors.New("unknow command")
	}

	return err
}
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
	img, err := wd.Screenshot()
	if err != nil {
		output, err := os.Create("/Users/ruslan/work/src/github.com/ruslanBik4/httpgo/views/selenium/screeshot.jpg")
		if err != nil {
			log.Print(err)
		} else {
			output.Write(img)
			output.Close()
		}
	}
}
// webElement select
func getElement(wd selenium.WebDriver, token string) (result []selenium.WebElement){

	list := strings.Split(token, "::")
	token = list[0]
	wElements, err := wd.FindElements(selenium.ByCSSSelector, token)
	if err != nil {
		log.Print(token)
		panic(err)
	}
	if len(list) > 1 {
		stat := strings.TrimSpace(list[1])
		for i, elem := range wElements {
			if runCommand(stat, "!continue", elem) {
				result = append(result, wElements[i])
			}
		}
	} else {
		result = wElements
	}
	log.Print("select ", len(wElements), " from ", token)

	return
}
func runCommand(token, param string, wElem selenium.WebElement) bool{
var (	slnCommands  = map[string] func() error {
	"click":  wElem.Click,
	"clear":  wElem.Clear,
	"submit": wElem.Submit, }
	slnStat  = map[string] func() (bool, error) {"selected": wElem.IsSelected, "enabled": wElem.IsEnabled, "visible": wElem.IsDisplayed, }
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
					log.Fatal(token)
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