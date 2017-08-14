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
)
var 	fFilename    = flag.String("filename", "test.css", "file with css selenium rules")
var (
	wElem       []selenium.WebElement
)

func main() {
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer wd.Quit()

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

	b = bytes.Replace(b, []byte("\r\n"), []byte("\n"), -1)
	str := string(b)
	token, param := "", ""
	isCommand := false

	for i, s := range str {
		if i > n {
			log.Print("i>n")
			break
		}
		switch string(s) {
		case "{":
			wElem, err = wd.FindElements(selenium.ByCSSSelector, token)
			if err != nil {
				log.Print(token)
				panic(err)
			}
			log.Print(token)
			token, param, isCommand = "", "", false
		case "}":
			token, param, isCommand, wElem = "", "", false, nil
		case ":":
			isCommand = (wElem != nil) || (token == "url")
			if !isCommand {
				token += string(s)
			}
		case "\n":
			if token == "url" {
				// Navigate to the simple playground interface.
				if err := wd.Get("http://" + param); err != nil {
					panic(err)
				}
				log.Print("open " + param)

				img, err := wd.Screenshot()
				if err != nil {
					output, err := os.Create("screeshot.jpg")
					if err != nil {
						log.Print(err)
					} else {
						output.Write(img)
						output.Close()
					}
				}
				wd.MaximizeWindow("")

			} else if (token > "") && (wElem != nil) {
				for i, elem := range wElem {
					if (elem !=nil) && !runCommand(token, param, elem) {
						wElem[i] = nil
					}
				}
			}
			token, param, isCommand = "", "", false
		case " ":
			continue
		default:
			if isCommand {
				param += string(s)
			} else {
				token += string(s)
			}
		}
	}
}
func runCommand(token, param string, wd selenium.WebElement) bool{
var (	slnCommands  = map[string] func() error {
	"click": wd.Click,
	"clear": wd.Clear,
	"submit": wd.Submit, }
	slnStat  = map[string] func() (bool, error) {"selected": wd.IsSelected, "enabled": wd.IsEnabled, "displayed": wd.IsDisplayed, }
	slnText  = map[string] func() (string, error) { "tag": wd.TagName, "text": wd.Text, }
	slnCSS   = map[string] func(string) (string, error) {"css": wd.CSSProperty, "attr": wd.GetAttribute, }
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
					log.Print(token+" is ", ok, param)

				}
			} else {
				log.Print(token+" is ", ok)
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
		if err := wd.SendKeys(param); err != nil {
			log.Print("sendkey err", err, param)
			panic(err)
		}
		log.Print(token + ":" + param)
	} else {
		log.Print("unknow command", token)
	}

	return true
}