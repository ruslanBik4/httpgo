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

	i := 0
	for (status != "ready") && (i < 1000) {
		time.Sleep(5)
		i++
		status = services.Status("schema")
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
	f_port    = flag.String("port", ":8080", "host address to listen on")
	f_static  = flag.String("path", "/Users/ruslan/work/src/github.com/ruslanBik4/httpgo", "path to static files")
	f_web     = flag.String("web", "/Users/ruslan/PhpstormProjects/thetravel/web", "path to web files")
	f_session = flag.String("sessionPath", "/var/lib/php/session", "path to store sessions data")
	f_cache   = flag.String("cacheFileExt", `.eot;.ttf;.woff;.woff2;.otf;`, "file extensions for caching HTTPGO")
	f_chePath = flag.String("cachePath", "css;js;fonts;images", "path to cached files")
)

func init() {
	flag.Parse()
	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(f_static, f_web, f_session); err != nil {
		fmt.Println(err)
	}
	services.InitServices()
}
