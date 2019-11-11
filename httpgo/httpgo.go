// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpgo

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/logs"
)

// HttpGo implement rest api http/https server for operation with storage
type HttpGo struct {
	mainServer *fasthttp.Server
	listener   net.Listener
	broadcast  chan string
	apis       *apis.Apis
	cfg        *CfgHttp
}

// NewHttpgo get configuration option from cfg
// listener to receive requests
func NewHttpgo(cfg *CfgHttp, listener net.Listener, apis *apis.Apis) *HttpGo {

	if apis.Ctx == nil {
		apis.Ctx = make(map[string]interface{}, 0)
	}
	apis.Ctx["ACC_VERSION"] = httpgoVersion

	cfg.Server.Handler = apis.Handler
	cfg.Server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logs.ErrorLog(err, ctx.String())
	}
	cfg.Server.Logger = &fastHTTPLogger{}

	logs.DebugLog("Server get files under %db size", cfg.Server.MaxRequestBodySize)

	return &HttpGo{
		mainServer: cfg.Server,
		listener:   listener,
		broadcast:  make(chan string),
		apis:       apis,
		cfg:        cfg,
	}
}

// Run starting http or https server according to secure
// certFile and keyFile are paths to TLS certificate and key files for https server
func (a *HttpGo) Run(secure bool, certFile, keyFile string) error {
	go a.listenOnShutdown()
	if secure {
		return a.mainServer.ServeTLS(a.listener, certFile, keyFile)
	}

	return a.mainServer.Serve(a.listener)
}

// listenOnShutdown implement correct shutdown server
func (a *HttpGo) listenOnShutdown() {
	ch := make(chan os.Signal)
	KillSignal := syscall.Signal(a.cfg.KillSignal)
	// syscall.SIGTTIN
	signal.Notify(ch, os.Interrupt, os.Kill, KillSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	logs.StatusLog("Shutdown service starting %v on signal '%v'", time.Now(), KillSignal)
	signShut := <-ch
	logs.StatusLog(signShut.String())

	close(a.broadcast)

	err := a.mainServer.Shutdown()
	if err != nil {
		logs.ErrorLog(err)
	}
	err = a.listener.Close()
	if err != nil {
		logs.ErrorLog(err)
	}
}

// fastHTTPLogger wrap logging server
type fastHTTPLogger struct {
	logs.LogsType
}

func (log *fastHTTPLogger) Printf(mess string, args ...interface{}) {
	args = append([]interface{}{mess}, args...)
	logs.DebugLog(args...)
}
