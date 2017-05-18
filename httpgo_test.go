// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
	"flag"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/models/services"
	"fmt"
)

func TestMain(t *testing.T) {
	t.Skipped()
}
func init() {
	flag.Parse()
	ServerConfig := server.GetServerConfig()
	if err := ServerConfig.Init(f_static, f_web, f_session); err != nil {
		fmt.Println(err)
	}
	services.InitServices()
}
