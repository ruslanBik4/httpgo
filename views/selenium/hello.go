// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/tebeka/selenium"
	"time"
	"log"
	"strings"
)

func main()  {

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get("http://vps-20777.vps-default-host.net/extranet/"); err != nil {
		panic(err)
	}

	// Get a reference to the text box containing code.
	elem, err := wd.FindElement(selenium.ByCSSSelector, "#searchField")
	if err != nil {
		log.Print("#searchField")
		panic(err)
	}
	// Remove the boilerplate code already in the text box.
	if err := elem.Clear(); err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond * 3000)

	// Enter some new code in text box.
	err = elem.SendKeys(`Алупка`)
	if err != nil {
		log.Print("Алупка")
		panic(err)
	}

	// Click the run button.
	btn, err := wd.FindElement(selenium.ByCSSSelector, "a.collapsed[href='#collapse525']")
	if err != nil {
		log.Print("a.collapsed")
		panic(err)
	}
	log.Print(btn.Text())
	if err := btn.Click(); err != nil {
		log.Print("click")
		panic(err)
	}

    link, err := btn.GetAttribute("href")
	if err != nil {
		log.Print("href")
		panic(err)
	}
	pos := strings.Index(link, "#")
	link = link[pos:]
	btn, err = wd.FindElement(selenium.ByCSSSelector,  link + " > div > ul > li:nth-child(1) > a")
	if err != nil {
		log.Print("href")
		panic(err)
	}
	log.Print(btn.Text())
	if err := btn.Click(); err != nil {
		log.Print("click")
		panic(err)
	}
	time.Sleep(time.Millisecond * 1000)

	// Wait for the program to finish running and get the output.
	outputDiv, err := wd.FindElement(selenium.ByCSSSelector, "select[name=id_prop_hotels_type_list]")
	if err != nil {
		log.Print(link)
		panic(err)
	}
	output := ""
	for {
		output, err = outputDiv.Text()
		if err != nil {
			panic(err)
		}
		if output != "Waiting for remote server..." {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	log.Print(output)

}