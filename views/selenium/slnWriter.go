/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tebeka/selenium"

	"github.com/ruslanBik4/logs"
)

type currentElem struct {
	elem     selenium.WebElement
	Selector string
}

var (
	//command [] tCommand
	fFileName = flag.String("filename", "new.sln", "file with css selenium rules")
	fWDPath   = flag.String("wd_path", "/Users/ruslan/chromedriver", "full path of chrome web-driver")
	fURL      = flag.String("url", "https://161.97.144.240:4443/", "path to screenshot files")
)

func main() {
	defer func() {
		switch err := recover().(type) {
		case nil:
		case error:
			logs.ErrorStack(err, "stop my work %#v, %T", err)
			time.Sleep(time.Millisecond * 5000)
			//wd.SetAlertText(err.Error())

		default:
			logs.StatusLog("recover: %v", err)
		}
	}()

	flag.Parse()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		logs.ErrorLog(err, err.Error())
		if strings.Contains(err.Error(), "connection refused") {
			cmd := exec.Command(*fWDPath)
			err = cmd.Start()
			if err == nil {
				wd, err = selenium.NewRemote(caps, "http://localhost:9515")
			}
		}
		if err != nil {
			panic(err) // panic is used only as an example and is not otherwise recommended.
		}
	}
	defer wd.Quit()

	// if !strings.HasPrefix(*fURL, "http://") {
	// 	*fURL = "http://" + *fURL41i1U9Ojlv0lBcy58J_cRA==
	// }

	if err := wd.Get(*fURL); err != nil {
		panic(err)
	}

	ioWriter, err := os.Create(*fFileName)
	ioWriter.Write([]byte("url: " + *fURL + "\n"))

	defer ioWriter.Close()

	wd.MaximizeWindow("")

	st, err := wd.Status()
	logs.DebugLog(" %+v", st)
	time.Sleep(time.Millisecond * 1000)

	for activeElem, err := wd.ActiveElement(); err == nil && activeElem == nil; {
		logs.DebugLog(" %+v", activeElem)
		activeElem, err = wd.ActiveElement()
	}

	var types = map[string]string{
		"email":    "bik4ruslan@gmail.com",
		"password": "41i1U9Ojlv0lBcy58J_cRA==",
	}

	pass := findElementBySelector(wd, `input`)
	for _, elem := range pass {
		typeName, err := elem.GetAttribute("type")
		elem.Click()
		elem.Clear()
		err = elem.SendKeys(types[typeName])
		if err != nil {
			logs.ErrorLog(err, typeName, types[typeName])
		}

	}

	btn := findElementBySelector(wd, `button`)

	btn[0].Click()

	resp, err := wd.ExecuteScriptRaw(`document.body.addEventListener("click", 
		function(event){
		document.getElementsByClassName("sln_writer").classList.remove("sln_writer");
		event.target.classList.add("sln_writer");
		event.target.focus();
		// return true;
})`, nil)

	if err != nil {
		logs.ErrorLog(err)
	}

	logs.DebugLog(string(resp))
	time.Sleep(time.Millisecond * 1000)

	var s string
	_, err = fmt.Scanln(&s)
	if err != nil {
		logs.ErrorLog(err)
	}
	logs.DebugLog(" %+v", s)
	logs.DebugLog(wd.Status())

	var activeElem = currentElem{}
	for count, url := 0, *fURL; (url > "") && (count < 1000) && (err == nil); url, err = wd.CurrentURL() {
		wd.AcceptAlert()
		elem, err := wd.FindElement(selenium.ByClassName, "sln_writer")
		if err != nil {
			err = nil
			continue
		}
		newElem, err := saveNewElement(elem, url)
		if err != nil {
			continue
		}
		if activeElem.Selector != newElem.Selector {
			if activeElem.Selector > "" {
				ioWriter.Write([]byte(activeElem.Selector + "{\nclick\n}\n"))
			}
			activeElem = newElem
		}
		count++
	}
	log.Print(wd.Status())
}

func saveNewElement(elem selenium.WebElement, url string) (result currentElem, err error) {
	var id, tag, name, class, href string
	var inputFields = map[string]string{"input": "", "select": "", "textarea": ""}

	result.elem = elem
	id, err = elem.GetAttribute("id")
	if err != nil {
		if strings.HasPrefix(err.Error(), "") {

		} else {
			panic(err) // panic is used only as an example and is not otherwise recommended.
		}
	}
	if id > "" {
		result.Selector = "#" + id
		return
	}
	tag, err = elem.TagName()
	if err == nil {
		if _, ok := inputFields[tag]; ok {
			name, err = elem.CSSProperty("name")
			if err == nil {
				result.Selector = tag + "[name=" + name + "]"
			}
		} else if tag == "a" {
			href, err = elem.GetAttribute("href")
			if href > "" {
				if strings.HasPrefix(href, url) {
					href = strings.TrimPrefix(href, url)
				}
				result.Selector = tag + "[href='" + href + "']"
			}
		} else {
			class, err = elem.GetAttribute("class")
			class = strings.Replace(class, "sln_writer", "", -1)
			if class > "" {
				result.Selector = tag + "." + class + "::first"
			} else {

				result.Selector = tag + "::first"
			}

		}
	}

	return
}

// find element by selector & panic if error occupiers
func findElementBySelector(wd selenium.WebDriver, token string) []selenium.WebElement {
	wElements, err := wd.FindElements(selenium.ByTagName, token)
	if err != nil {
		logs.ErrorLog(err, token)
		return nil
	}

	return wElements
}
