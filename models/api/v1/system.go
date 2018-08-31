// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/views/templates/system"
	"net/http"
	"os/exec"
)

// HandleUpdateServer update httpgo & {project} code from git  & build new version httpgo
// branch - name git branch
// @/api/update/?branch={branch}
func HandleUpdateServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("./webserver.sh", "update")
	cmd.Dir = ServerConfig.SystemPath()
	if r.FormValue("branch") > "" {
		cmd.Args = append(cmd.Args, r.FormValue("branch"))
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
	}

}

// HandleRestartServer restart service
// @/api/restart/
func HandleRestartServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("systemctl", "restart", "httpgo")
	cmd.Dir = ServerConfig.SystemPath()
	cmd.Stdout = w

	err := cmd.Start()
	//if err == nil {
	//	if stdoutStderr, err := cmd.Output(); err == nil {
	//		views.RenderOutput(w, stdoutStderr)
	//	}
	//}
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		w.Write([]byte("restart server go"))
	}
}

// HandleLogServer show log httpgo
// @/api/log/
func HandleLogServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("systemctl", "status", "httpgo")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
	}
}
// HandleUpdateSource incremental update
// update httpgo & {project} co de from git  & build new version httpgo
// branch - name git branch
// @/api/update/source/
func HandleUpdateSource(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("./webserver.sh", "pull")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
		arr := []string{"/api/update/test/", "/api/update/build/", "/api/restart/"}
		system.WriteAddRescanJS(w, arr)
	}

}

// HandleUpdateTest run tests project
func HandleUpdateTest(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("./webserver.sh", "test")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
	}

}

// HandleUpdateBuild build project
func HandleUpdateBuild(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("./webserver.sh", "build")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
	}

}
