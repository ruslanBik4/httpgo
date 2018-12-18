// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"os/exec"
	"time"

	"github.com/ruslanBik4/httpgo/views"
)

// HandleLogServer show status httpgo
// @/api/status/
func HandleStatusServer(w http.ResponseWriter, r *http.Request) {
	//ServerConfig := server.GetServerConfig()

	cmd := exec.Command("systemctl", "status", "httpgo")
	//cmd.Dir = ServerConfig.SystemPath()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		views.RenderInternalError(w, err)
	} else {
		views.RenderOutput(w, stdoutStderr)
	}
}

// HandleShowErrorsServer show services errors
// @/api/log/errors/
func HandleShowErrorsServer(w http.ResponseWriter, r *http.Request) {

	stdout, err := runJournal()
	if err == nil {
		var stdoutStderr []byte
		stdoutStderr, err = runGrep(getParamLog(r, "ERROR"), stdout)
		if err == nil {
			views.RenderOutput(w, stdoutStderr)
		}
	}
	if err != nil {
		views.RenderInternalError(w, err)
	}
}

// HandleShowStatusServer show status message
// @/api/log/status/
func HandleShowStatusServer(w http.ResponseWriter, r *http.Request) {

	stdout, err := runJournal()
	if err == nil {
		var stdoutStderr []byte
		stdoutStderr, err = runGrep(getParamLog(r, "STATUS"), stdout)
		if err == nil {
			views.RenderOutput(w, stdoutStderr)
		}
	}
	if err != nil {
		views.RenderInternalError(w, err)
	}
}

// HandleShowDebugServer show debug messages
// @/api/log/debug/
func HandleShowDebugServer(w http.ResponseWriter, r *http.Request) {

	stdout, err := runJournal()
	if err == nil {
		var stdoutStderr []byte
		stdoutStderr, err = runGrep(getParamLog(r, "DEBUG"), stdout)
		if err == nil {
			views.RenderOutput(w, stdoutStderr)
		}
	}

	if err != nil {
		views.RenderInternalError(w, err)
	}
}

// logs token for journalctl
const (
	startLogPram = "httpgo"
	sepLogParam  = ".*"
)

func runGrep(params string, stdout []byte) ([]byte, error) {
	cmd := exec.Command("grep", "-oE", params)
	stdin, err := cmd.StdinPipe()

	if err == nil {
		go func() {
			defer stdin.Close()
			_, err = stdin.Write(stdout)
		}()
		return cmd.CombinedOutput()

	}

	return nil, err
}
func getParamLog(r *http.Request, flagLog string) string {

	logParam := startLogPram + sepLogParam + flagLog
	if dateDebug := r.FormValue("date"); dateDebug > "" {
		logParam += sepLogParam + dateDebug
	} else {
		logParam += sepLogParam + time.Now().Format("2006/01/02")
	}
	if timeDebug := r.FormValue("time"); timeDebug > "" {
		logParam += sepLogParam + timeDebug
	}

	return logParam + sepLogParam
}
func runJournal() ([]byte, error) {
	//ServerConfig := server.GetServerConfig()
	//cmd.Dir = ServerConfig.SystemPath()
	return exec.Command("journalctl", "-u", "httpgo").Output()
}
