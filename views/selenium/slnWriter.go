// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/tebeka/selenium"
	"os"
	"flag"
	_ "time"
	"log"
	"time"
	"strings"
)

type currentElem struct {
	elem selenium.WebElement
	Selector string
}
var (
	//command [] tCommand
	fFileName    = flag.String("filename", "new.sln", "file with css selenium rules")
	fURL  = flag.String("url", "vps-20777.vps-default-host.net/extranet/", "path to screenshot files")
)

func main()  {
	defer func() {
		err := recover()
		if err, ok := err.(error); ok {
			//wd.SetAlertText(err.Error())
			log.Printf("%#v, %T", err, err)
			time.Sleep(time.Millisecond * 5000)
		}
	}()
	flag.Parse()
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer wd.Quit()

	if !strings.HasPrefix( *fURL, "http://") {
		*fURL = "http://" + *fURL
	}

	if err := wd.Get( *fURL ); err != nil {
		panic(err)
	}


	ioWriter, err := os.Create(*fFileName)
	ioWriter.Write([]byte( "url: " + *fURL + "\n"))

	defer ioWriter.Close()

	wd.MaximizeWindow("")

	var activeElem = currentElem{}
	time.Sleep(time.Millisecond * 1000)

	resp, err := wd.ExecuteScriptRaw(`$("body").click(function() {
		$(".sln_writer").removeClass("sln_writer");
		$(event.target).addClass("sln_writer").focus();
		return true;})`, nil)
	time.Sleep(time.Millisecond * 1000)
	log.Print( string(resp) )
	log.Print( wd.Status() )
	if err != nil {
		log.Print(err)
	}
	for count, url := 0, *fURL; (url > "") && (count < 1000) && (err == nil); url, err = wd.CurrentURL() {
			wd.AcceptAlert()
		elem, err := wd.FindElement(selenium.ByClassName, "sln_writer")
		if err != nil {
			err = nil
			continue
		}
		newElem, err := saveNewElement(elem, url)
		if activeElem.Selector != newElem.Selector {
			if activeElem.Selector > "" {
				ioWriter.Write([]byte(activeElem.Selector + "{\nclick\n}\n"))
			}
			activeElem = newElem
		}


	}
	log.Print( wd.Status() )
}

func saveNewElement(elem selenium.WebElement, url string) (result currentElem, err error){
	var id, tag, name, class, href string
	var inputFields = map[string] string { "input": "", "select": "", "textarea": "", }

	result.elem = elem
	id, err = elem.GetAttribute("id")
	if err != nil {
		if strings.HasPrefix( err.Error(), "" ) {

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
		if _, ok := inputFields[ tag ]; ok  {
			name, err = elem.CSSProperty("name")
			if err == nil {
				result.Selector = tag + "[name=" + name + "]"
			}
		} else if tag == "a" {
			href, err = elem.GetAttribute("href")
			if href > "" {
				if strings.HasPrefix(href, url) {
					href =strings.TrimPrefix(href, url)
				}
				result.Selector = tag + "[href='" + href + "']"
			}
		} else {
			class, err = elem.GetAttribute("class")
			class      = strings.Replace(class, "sln_writer", "", 0)
			if class > "" {
				result.Selector = tag + "." + class + "::first"
			} else {

				result.Selector = tag + "::first"
			}

		}
	}

	return
}