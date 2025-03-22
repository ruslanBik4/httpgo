/*
 * Copyright (c) 2022-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"fmt"
	"go/types"
	"mime/multipart"
	"net"
	"os"
	"os/signal"
	"path"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/domsolutions/http2"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/gotools"
	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/logs"
)

// HttpGo implement rest api http/https server for operation with storage
type HttpGo struct {
	mainServer *fasthttp.Server
	listener   net.Listener
	broadcast  chan string
	apis       *Apis
	cfg        *CfgHttp
	store      *Store
}

var regIp = regexp.MustCompile(`for=s*(\d+\.?)+,`)

var GoVersion string

// NewHttpgo get configuration option from cfg
// listener to receive requests
func NewHttpgo(cfg *CfgHttp, listener net.Listener, apis *Apis) *HttpGo {

	if cfg.HTTP2 != nil {
		http2.ConfigureServer(cfg.Server, *cfg.HTTP2)
		logs.StatusLog("set HTTP2 server configuration")
	}

	if cfg.PortRedirect > "" {
		RunRedirectNoSecure(cfg.PortRedirect)
	}

	if apis.Ctx == nil {
		apis.Ctx = make(map[string]any, 0)
	}

	apis.Ctx[ApiVersion] = httpgoVersion
	if cfg.Server != nil {
		apis.Ctx[ServerName] = fmt.Sprintf("%v HTTPGO/%v (CentOS) backend by Go %v", cfg.Server.Name, httpgoVersion, GoVersion)
	}

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
	cfg.Server.ContinueHandler = func(header *fasthttp.RequestHeader) bool {

		logs.StatusLog("has Continue !", header)
		return true
	}
	cfg.Server.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
		logs.ErrorLog(err, ctx.String())
		switch err {
		case fasthttp.ErrBodyTooLarge:
			ctx.SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
		case fasthttp.ErrNoMultipartForm, fasthttp.ErrNoArgValue:
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.Response.SetBodyString(err.Error())
		default:
			// if  !bytes.Equal(ctx.Request.URI().Scheme(), []byte("http")) {
			// 	uri := ctx.Request.URI()
			// 	uri.SetScheme("http")
			// 	ctx.RedirectBytes(uri.FullURI(), fasthttp.StatusFound)
			// }
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}
	cfg.Server.Logger = &fastHTTPLogger{}
	cfg.Server.KeepHijackedConns = true
	cfg.Server.CloseOnShutdown = true

	var h *HttpGo
	if len(cfg.Domains) == 0 {
		cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			apis.Handler(ctx)
		}
	} else {
		logs.DebugLog("Subdomains is %+v", cfg.Domains)
		cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			for subD, ip := range cfg.Domains {
				host := gotools.BytesToString(ctx.Host())
				if host != ip && strings.HasPrefix(host, subD) {
					if isLocalDirectory(ip) {
						p := gotools.BytesToString(ctx.URI().Path())
						if p == "" || p == "/" {
							p = "index.html"
						}
						logs.StatusLog(ip, p)
						fileName := path.Join(".", ip, p)
						err := ctx.Response.SendFile(fileName)
						if err != nil {
							logs.ErrorLog(err, ctx.String())
							return
						}
						ct, fileName := views.GetContentType(ctx, fileName)
						ctx.Response.Header.SetContentType(ct)
						return
					}

					if !isLocalRedirect(ip) {
						ctx.Redirect(ip, fasthttp.StatusMovedPermanently)
						logs.DebugLog("redirect", ip)
						return
					}

					if !strings.HasSuffix(listener.Addr().String(), ip) {
						url := fmt.Sprintf("%s://%s%s/", ctx.URI().Scheme(), host, ip)
						logs.DebugLog("redirect:", url)
						ctx.Redirect(url, fasthttp.StatusMovedPermanently)
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
			addr := gotools.BytesToString(ipClient)
			if len(ipClient) == 0 {
				ipClient = ctx.Request.Header.Peek("Forwarded")
				ips := regIp.FindSubmatch(ipClient)

				if len(ips) == 0 {
					addr = gotools.BytesToString(ctx.Request.Header.Peek("X-ProxyUser-Ip"))
					if len(addr) == 0 {
						addr = ctx.Conn().RemoteAddr().String()
					}
				} else {
					addr = gotools.BytesToString(ips[0])
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

	store := NewStore()
	apis.Ctx.AddValue(AppStore, store)
	// add cfg refresh routers, ignore errors
	apisRoute := createAdminRoutes(cfg)

	_ = apis.AddRoutes(apisRoute)

	h = &HttpGo{
		mainServer: cfg.Server,
		listener:   listener,
		broadcast:  make(chan string),
		apis:       apis,
		cfg:        cfg,
		store:      store,
	}
	logs.DebugLog("Server get files under %d size", cfg.Server.MaxRequestBodySize)

	return h
}

const separator = "/"

func isLocalRedirect(ip string) bool {
	const delim = ":"
	return strings.HasPrefix(ip, delim) || strings.HasPrefix(ip, separator)
}

func isLocalDirectory(ip string) bool {
	return strings.HasPrefix(ip, separator) || strings.HasSuffix(ip, separator)
}

func createAdminRoutes(cfg *CfgHttp) ApiRoutes {
	allowedParams := []InParam{
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
			Desc: `# HttpGo managements
reload cfg of httpgo from starting config file`,
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
				return cfg.Reload()
			},
		},
		"/httpgo/cfg/": {
			Desc: `# HttpGo managements
show config of httpGo`,
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
				return cfg, nil
			},
		},
		"/httpgo/cfg/add_ip": {
			Desc: `# HttpGo managements
add IP addresses into config of httpGo`,
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
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
			Params:    allowedParams,
		},
		"/httpgo/cfg/rm_ip": {
			Desc: `# HttpGo managements
remove IP addresses show config of httpGo`,
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
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
			Params:    allowedParams,
		},
		"/httpgo/store/": {
			Desc: " # store",
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
				id := ctx.UserValue(crud.ParamsID.Name).(int32)
				name := ctx.UserValue(crud.ParamsName.Name).(string)
				return ctx.UserValue(AppStore).(*Store).Get(uint64(id), name), nil
			},
			Params: []InParam{
				crud.ParamsID,
				crud.ParamsName,
			},
		},
		"/httpgo/store/put": {
			Desc: " # store",
			Fnc: func(ctx *fasthttp.RequestCtx) (any, error) {
				val := ctx.UserValue("blob").([]*multipart.FileHeader)
				name := ctx.UserValue(crud.ParamsName.Name).(string)
				return ctx.UserValue(AppStore).(*Store).Set(ctx, name, val), nil
			},
			Method:    POST,
			Multipart: true,
			Params: []InParam{
				crud.ParamsName,
				crud.NewFileParam("blob", "file to saving in store"),
			},
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
func (h *HttpGo) Run(secure bool, certFile, keyFile string) error {

	h.apis.Https = secure
	h.apis.StartTime = time.Now()
	//todo change parameters type on
	go h.listenOnShutdown()
	if secure {
		return h.mainServer.ServeTLS(h.listener, certFile, keyFile)
	}

	return h.mainServer.Serve(h.listener)
}

// listenOnShutdown implement correct shutdown server
func (h *HttpGo) listenOnShutdown() {
	ch := make(chan os.Signal)
	KillSignal := syscall.Signal(h.cfg.KillSignal)
	// syscall.SIGTTIN
	signal.Notify(ch, KillSignal, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	logs.StatusLog("Shutdown service starting %v on signal '%v'", time.Now(), KillSignal)
	signShut := <-ch

	logs.StatusLog("Shutdown service get signal: " + signShut.String())
	close(h.broadcast)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err := h.mainServer.ShutdownWithContext(ctx)
	if err != nil {
		logs.ErrorLog(err)
	}

	err = h.listener.Close()
	if err != nil {
		logs.ErrorLog(err)
	}
}

// fastHTTPLogger wrap logging server
type fastHTTPLogger struct {
	logs.LogsType
}

func (log *fastHTTPLogger) Printf(mess string, args ...any) {

	if strings.Contains(mess, "error") {
		if slices.ContainsFunc(args, func(a any) bool {
			s, ok := a.(error)
			return ok && strings.Contains(s.Error(), "tls: unknown certificate")
		}) {
			//	nothing to tell :-)
		} else if strings.Contains(mess, "serving connection") {
			logs.StatusLog(fmt.Sprintf(mess, args...))
		} else {
			logs.ErrorLog(errors.New(mess), args...)
		}
	} else {
		logs.DebugLog(append([]any{mess}, args...)...)
	}
}
