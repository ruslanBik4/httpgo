// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"github.com/ruslanBik4/httpgo/models/server"
	"github.com/ruslanBik4/httpgo/views"
	"net/http"
	"os/exec"
)

func renderOutput(w http.ResponseWriter, stdoutStderr []byte) {

	views.WriteHeaders(w)
	w.Write([]byte("<pre>"))
	w.Write(bytes.Replace(stdoutStderr, []byte("\n"), []byte("<br>"), 0))
	w.Write([]byte("</pre>"))
}
func HandleUpdateServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("./webserver.sh", "update")
	cmd.Dir = ServerConfig.SystemPath()
	if r.FormValue("branch") > "" {
		cmd.Args = append(cmd.Args, r.FormValue("branch"))
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		renderOutput(w, stdoutStderr)
	}

}
func HandleLogServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("systemctl", "status", "httpgo")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		renderOutput(w, stdoutStderr)
	}
}
func HandleRestartServer(w http.ResponseWriter, r *http.Request) {
	ServerConfig := server.GetServerConfig()

	cmd := exec.Command("systemctl", "restart", "httpgo")
	cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		renderOutput(w, stdoutStderr)
	}
}
