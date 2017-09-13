// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qb

import (
	"flag"
	"fmt"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/services"
	"testing"
	"time"
)

func TestQBCreate(t *testing.T) {

	status := services.Status("schema")

	for i := 0; (status != "ready") && (i < 1000); status = services.Status("schema") {
		time.Sleep(5)
		i++

	}
	t.Log(status)
	qb := CreateEmpty()

	qb.AddTable("a", "rooms").AddField("name", "title").AddField("num", "id")

	v, err := qb.SelectToMultidimension()

	if err != nil {
		t.Error(err)
	} else {
		t.Log(v)
		t.Skipped()
	}
}

var (
	fPort    = flag.String("port", ":8080", "host address to listen on")
	fStatic  = flag.String("path", "./", "path to static files")
	fWeb     = flag.String("web", "/Users/ruslan/PhpstormProjects/thetravel/web", "path to web files")
	fSession = flag.String("sessionPath", "/var/lib/php/session", "path to store sessions data")
	fCache   = flag.String("cacheFileExt", `.eot;.ttf;.woff;.woff2;.otf;`, "file extensions for caching HTTPGO")
	fChePath = flag.String("cachePath", "css;js;fonts;images", "path to cached files")
)

func init() {
	flag.Parse()
	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(fStatic, fWeb, fSession); err != nil {
		fmt.Println(err)
	}
	services.InitServices()
}
