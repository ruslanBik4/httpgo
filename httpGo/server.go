// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpGo

import (
	"fmt"
	"go/types"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/logs"
)

// HttpGo implement rest api http/https server for operation with storage
type HttpGo struct {
	mainServer *fasthttp.Server
	listener   net.Listener
	broadcast  chan string
	apis       *Apis
	cfg        *CfgHttp
}

var regIp = regexp.MustCompile(`for=s*(\d+\.?)+,`)

// NewHttpgo get configuration option from cfg
// listener to receive requests
func NewHttpgo(cfg *CfgHttp, listener net.Listener, apis *Apis) *HttpGo {

	if apis.Ctx == nil {
		apis.Ctx = make(map[string]interface{}, 0)
	}

	apis.Ctx["ACC_VERSION"] = httpgoVersion

	// cfg.Server.HeaderReceived = func(header *fasthttp.RequestHeader) fasthttp.RequestConfig {
	// 	uri := header.RequestURI()
	// 	if bytes.HasPrefix(uri, []byte("https")) {
	//
	// 	}
	// 	logs.StatusLog(string(uri))
	// 	return fasthttp.RequestConfig{}
	// }
	// cfg.Server.NextProto("https", func(c net.Conn) error {
	// 	n := c.LocalAddr().Network()
	// 	if strings.HasPrefix(n, "https") {
	//
	// 	}
	// 	logs.StatusLog(n)
	// 	return nil
	// })
	cfg.Server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logs.ErrorLog(err, ctx.String())
		// if  !bytes.Equal(ctx.Request.URI().Scheme(), []byte("http")) {
		// 	uri := ctx.Request.URI()
		// 	uri.SetScheme("http")
		// 	ctx.RedirectBytes(uri.FullURI(), fasthttp.StatusFound)
		// }
	}
	cfg.Server.Logger = &fastHTTPLogger{}

	logs.DebugLog("Server get files under %d size", cfg.Server.MaxRequestBodySize)
	if len(cfg.Domains) == 0 {
		cfg.Server.Handler = apis.Handler
	} else {
		logs.DebugLog("Subdomains is %+v", cfg.Domains)
		cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			for subD, ip := range cfg.Domains {
				const delim = ":"
				host := string(ctx.Host())
				if host != ip && strings.HasPrefix(host, subD) {

					if ip[:1] == delim && !strings.HasSuffix(listener.Addr().String(), ip) {
						url := fmt.Sprintf("%s://%s%s/", ctx.URI().Scheme(), host, ip)
						logs.DebugLog("redirect:", url)
						ctx.Redirect(url, fasthttp.StatusMovedPermanently)
						return
					}

					if ip[:1] != delim {
						ctx.Redirect(ip, fasthttp.StatusMovedPermanently)
						return
					}
				}
			}

			apis.Handler(ctx)
		}
	}

	if cfg.IsAccess() {
		if cfg.ChkConn {
			listener = &blockListener{
				listener,
				cfg.AccessConf,
			}
		}
		handler := cfg.Server.Handler
		cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			ipClient := ctx.Request.Header.Peek("X-Forwarded-For")
			addr := string(ipClient)
			if len(ipClient) == 0 {
				ipClient = ctx.Request.Header.Peek("Forwarded")
				ips := regIp.FindSubmatch(ipClient)

				if len(ips) == 0 {
					addr = string(ctx.Request.Header.Peek("X-ProxyUser-Ip"))
					if len(addr) == 0 {
						addr = ctx.Conn().RemoteAddr().String()
					}
				} else {
					addr = string(ips[0])
				}
			}

			if cfg.Allow(ctx, addr) || !cfg.Deny(ctx, addr) {
				handler(ctx)
				return
			}

			logs.DebugLog(addr, ctx.Request.Header.String(), cfg)
			ctx.Error(cfg.Mess, fasthttp.StatusForbidden)
		}
	}
	// add cfg refresh routers, ignore errors
	apisRoute := createAdminRoutes(cfg)

	_ = apis.AddRoutes(apisRoute)

	return &HttpGo{
		mainServer: cfg.Server,
		listener:   listener,
		broadcast:  make(chan string),
		apis:       apis,
		cfg:        cfg,
	}
}

func createAdminRoutes(cfg *CfgHttp) ApiRoutes {
	params := []InParam{
		{
			Name: "allow_ip",
			Type: NewSliceTypeInParam(types.String),
		},
		{
			Name: "deny_ip",
			Type: NewSliceTypeInParam(types.String),
		},
		{
			Name: "msg",
			Type: NewTypeInParam(types.String),
		},
	}

	return ApiRoutes{
		"/httpgo/cfg/reload": {
			Desc: "reload cfg of httpgo from starting config file",
			Fnc: func(ctx *fasthttp.RequestCtx) (interface{}, error) {
				return cfg.Reload()
			},
		},
		"/httpgo/cfg/": {
			Desc: "show config of httpGo",
			Fnc: func(ctx *fasthttp.RequestCtx) (interface{}, error) {
				return cfg, nil
			},
		},
		"/httpgo/cfg/add_ip": {
			Desc: "show config of httpGo",
			Fnc: func(ctx *fasthttp.RequestCtx) (interface{}, error) {
				if ips, ok := ctx.UserValue("allow_ip").([]string); ok {
					cfg.AllowIP = append(cfg.AllowIP, ips...)
				}
				if ips, ok := ctx.UserValue("deny_ip").([]string); ok {
					cfg.DenyIP = append(cfg.DenyIP, ips...)
				}

				if msg, ok := ctx.UserValue("msg").(string); ok {
					cfg.Mess = msg
				}

				return cfg, nil
			},
			Multipart: true,
			Method:    POST,
			OnlyAdmin: true,
			Params:    params,
		},
		"/httpgo/cfg/rm_ip": {
			Desc: "show config of httpGo",
			Fnc: func(ctx *fasthttp.RequestCtx) (interface{}, error) {
				if ips, ok := ctx.UserValue("allow_ip").([]string); ok {
					cfg.AllowIP = filterIPs(cfg.AllowIP, ips)
				}

				if ips, ok := ctx.UserValue("deny_ip").([]string); ok {
					cfg.DenyIP = filterIPs(cfg.DenyIP, ips)
				}

				return cfg, nil
			},
			Multipart: true,
			Method:    POST,
			OnlyAdmin: true,
			Params:    params,
		},
	}
}

func filterIPs(curIPs []string, ips []string) []string {

	tmpIps := curIPs[:0]
	for _, ip := range curIPs {
		isRm := false
		for _, rmIp := range ips {
			if rmIp == ip {
				isRm = true
				break
			}
		}
		if !isRm {
			tmpIps = append(tmpIps, ip)
		}
	}

	return tmpIps
}

// Run starting http or https server according to secure
// certFile and keyFile are paths to TLS certificate and key files for https server
func (a *HttpGo) Run(secure bool, certFile, keyFile string) error {
	//todo change patameters type on
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
	signal.Notify(ch, KillSignal, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	logs.StatusLog("Shutdown service starting %v on signal '%v'", time.Now(), KillSignal)
	signShut := <-ch
	logs.StatusLog("Shutdown service get signal: " + signShut.String())

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

	if strings.Contains(mess, "error") {
		if strings.Contains(mess, "serving connection") {
			logs.StatusLog(append([]interface{}{mess}, args...)...)
		} else {
			logs.ErrorLog(errors.New(mess), args...)
		}
	} else {
		logs.DebugLog(append([]interface{}{mess}, args...)...)
	}
}
