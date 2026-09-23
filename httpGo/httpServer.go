/*
 * Copyright (c) 2022-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"crypto/tls"
	"fmt"
	"go/types"
	"mime/multipart"
	"net"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/domsolutions/http2"
	"github.com/pkg/errors"
	"github.com/quic-go/quic-go/http3"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/gotools"
	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/logs"
)

// HttpGo implement rest api http/https server for operation with storage
type HttpGo struct {
	mainServer *fasthttp.Server
	listener   net.Listener
	broadcast  chan *string
	apis       *Apis
	cfg        *CfgHttp
	store      *Store
	rdServer   *fasthttp.Server
	h3Server   *http3.Server
}

// regForwardedFor extracts the "for=" identifier from a single element of an
// RFC 7239 Forwarded header, e.g. `for=192.0.2.60;proto=http;by=203.0.113.43`
// or `for="[2001:db8:cafe::17]:4711"`. Two real bugs fixed vs. the previous
// pattern (`for=s*(\d+\.?)+,`):
//  1. `s*` was very likely meant to be `\s*` (optional whitespace after
//     "for=", which RFC 7239 explicitly permits) - as written it matched zero
//     or more literal "s" characters, which happens to still "work" only
//     because zero of them is allowed.
//  2. The trailing `,` made the whole pattern require a SECOND forwarded-for
//     element after the one being captured - a single-hop Forwarded header
//     (the common case: exactly one proxy) has no comma at all, so it never
//     matched and silently fell through to the X-ProxyUser-Ip/remote-addr
//     fallback instead of the value the header actually carried.
//
// Bracketed IPv6 (`[::1]`) and a bare quoted value are both matched by the
// permissive `[^;,\s"]+` class; callers get back the raw for= value as-is
// (still possibly wrapped in `[...]` or a trailing `:port`), same shape the
// old code produced.
var regForwardedFor = regexp.MustCompile(`(?i)for=\s*"?(\[[^\]]+\]|[^;,\s"]+)"?`)

// extractClientIP centralizes what blockingHandler used to do inline, fixing
// two correctness bugs along the way (see regForwardedFor's doc comment for
// the Forwarded-header one):
//
//   - X-Forwarded-For can legitimately carry a comma-separated hop chain
//     ("client, proxy1, proxy2" - RFC 7239's non-standardized predecessor).
//     The previous code used the ENTIRE raw header value as "the" client IP,
//     which is not a single address at all once more than one hop is
//     present - almost certainly never matching a clean-IP allow/deny list.
//     This takes the first (left-most / original-client) entry, same
//     convention as every other XFF consumer, and trims whitespace.
//   - The Forwarded-header branch used to return the ENTIRE regex match,
//     including the literal "for=" prefix and trailing comma, as "the" IP
//     (Go's regexp only keeps the LAST capture of a repeated group like
//     `(\d+\.?)+`, so the old code fell back to group 0, the whole match).
//     This returns the actual captured for= value.
//
// SECURITY: every one of these headers (X-Forwarded-For, Forwarded,
// X-ProxyUser-Ip) is caller-supplied and trivially spoofable by any client
// that talks to this server directly - fixing the parsing bugs above does
// NOT make IP-based allow/deny (blockingHandler, cfg.Allow/cfg.Deny) safe
// against spoofing on its own. It is only meaningful when this server sits
// behind a reverse proxy/load balancer that you control, which OVERWRITES
// these headers with the real client address before forwarding (never just
// appends to whatever the client sent) - and even then, only the header your
// specific proxy actually sets should be trusted; the others should arguably
// be ignored rather than tried as a fallback chain. If httpgo is ever
// reachable directly (no proxy in front, or the proxy doesn't scrub these
// headers), a client can set X-Forwarded-For to an allow-listed IP and walk
// straight past cfg.Allow/cfg.Deny. Worth confirming which of these header
// checks the deployment actually needs before relying on this for anything
// more sensitive than coarse logging/rate-limiting hints.
func extractClientIP(ctx *fasthttp.RequestCtx) string {
	if xff := ctx.Request.Header.Peek("X-Forwarded-For"); len(xff) > 0 {
		first, _, _ := strings.Cut(gotools.BytesToString(xff), ",")
		return strings.TrimSpace(first)
	}

	if fwd := ctx.Request.Header.Peek("Forwarded"); len(fwd) > 0 {
		if m := regForwardedFor.FindSubmatch(fwd); m != nil {
			return gotools.BytesToString(m[1])
		}
	}

	if proxyIP := ctx.Request.Header.Peek("X-ProxyUser-Ip"); len(proxyIP) > 0 {
		return gotools.BytesToString(proxyIP)
	}

	return ctx.Conn().RemoteAddr().String()
}

// NewHttpgo get configuration option from cfg
// listener to receive requests
func NewHttpgo(cfg *CfgHttp, listener net.Listener, apis *Apis) *HttpGo {

	cfg.Server.Handler = apis.Handler
	if cfg.HTTP2 != nil {
		http2.ConfigureServer(cfg.Server, *cfg.HTTP2)
		logs.StatusLog("set HTTP2 server configuration")
		//reset user values for HTTP/2
		cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.ResetUserValues()
			apis.Handler(ctx)
		}
	}

	if apis.Ctx == nil {
		apis.Ctx = NewCtxApis(8)
	}

	apis.Ctx[ApiVersion] = HTTPGOVer
	if cfg.Server != nil {
		apis.Ctx[views.ServerName] = fmt.Sprintf(
			"%v HTTPGO/%v (%s) backend by Golang %v(%s) builded on %s",
			cfg.Server.Name,
			HTTPGOVer,
			runtime.GOOS,
			runtime.Version(),
			runtime.Compiler,
			OSVersion,
		)
	}

	//todo cfg.Server.HeaderReceived = func(header *fasthttp.RequestHeader) fasthttp.RequestConfig {
	// 	uri := header.RequestURI()
	// 	if bytes.HasPrefix(uri, []byte("https")) {
	//
	// 	}
	// 	logs.StatusLog(string(uri))
	// 	return fasthttp.RequestConfig{}
	// }
	cfg.Server.ContinueHandler = func(header *fasthttp.RequestHeader) bool {
		// Was StatusLog - a "100-continue" header arrives on every large/
		// chunked upload that expects one, so at StatusLog level (presumably
		// always-on, unlike DebugLog) this was pure noise on any server that
		// takes file uploads at any real volume.
		logs.DebugLog("has Continue !", header)
		return true
	}
	cfg.Server.ErrorHandler = renderError
	cfg.Server.Logger = &fastHTTPLogger{}
	cfg.Server.KeepHijackedConns = true
	cfg.Server.CloseOnShutdown = true

	if len(cfg.Domains) > 0 {
		logs.DebugLog("Subdomains is %+v", cfg.Domains)
		cfg.Server.Handler = spliDomainsHandler(cfg, listener, cfg.Server.Handler)
	}

	store := NewStore()
	apis.Ctx.AddValue(AppStore, store)
	// add cfg refresh routers, ignore errors
	apisRoute := createAdminRoutes(cfg)

	_ = apis.AddRoutes(apisRoute)

	h := &HttpGo{
		mainServer: cfg.Server,
		listener:   listener,
		broadcast:  make(chan *string),
		apis:       apis,
		cfg:        cfg,
		store:      store,
	}

	h.setHTTP3()

	if cfg.IsAccess() {
		if cfg.ChkConn {
			listener = &blockListener{
				listener,
				cfg.AccessConf,
			}
		}

		cfg.Server.Handler = blockingHandler(cfg, cfg.Server.Handler)
	}

	logs.DebugLog("Server get files under %d size", cfg.Server.MaxRequestBodySize)
	if cfg.PortRedirect > "" {
		h.rdServer = RunRedirectNoSecure(cfg)
	}

	return h
}

func spliDomainsHandler(cfg *CfgHttp, listener net.Listener, handler fasthttp.RequestHandler) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
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

		handler(ctx)
	}
}

func (h *HttpGo) setHTTP3() {
	if h.cfg.HTTP3 {
		var tlsCfg *tls.Config
		fPort := ":443"
		if ln, ok := h.listener.(LnMultiCerts); ok {
			tlsCfg = ln.tlsCfg.Clone()
			fPort = ln.fPort
		} else {
			certMaps := newCertMaps()
			if err := certMaps.loadCertificates(); err != nil {
				// Was: log the error and press on anyway, building a
				// tls.Config around a certMaps that may hold zero usable
				// certificates - every HTTP/3 handshake would then fail one
				// at a time in production instead of failing loudly, once,
				// at startup. HTTP/3 is opt-in (h.cfg.HTTP3) and the TCP
				// server already came up fine without it, so disabling just
				// this half of the server is strictly safer than serving UDP
				// port 443 with a broken TLS config.
				logs.ErrorLog(err, "HTTP/3 disabled: failed to load TLS certificates")
				return
			}
			tlsCfg = &tls.Config{GetCertificate: certMaps.getCertificate}
		}
		tlsCfg.NextProtos = []string{"h3"}
		if h.cfg.HTTP3Proxy == nil {
			h.cfg.HTTP3Proxy = &HTTP3ProxyConfig{}
		}

		if h3, err := NewHTTP3Proxy(
			tlsCfg,
			fPort,
			h.cfg.HTTP3Proxy,
		); err != nil {
			// Was stdlib log.Fatal - every other fatal/error path in this
			// file goes through the app's own logs package (consistent
			// formatting/output/whatever sinks logs.Fatal is wired to);
			// stdlib log.Fatal here was writing to a different, unrelated
			// destination for what is otherwise identical "can't start
			// HTTP/3, give up" behavior (both still call os.Exit(1)).
			logs.Fatal(err)
		} else {
			h.h3Server = h3
		}

		logs.StatusLog("[*] HTTP3 multi-site proxy running on %s, %v", fPort, h.h3Server.Addr)
		h3AltSvc := fmt.Sprintf(`h3="%s"; ma=86400`, fPort)
		handler := h.cfg.Server.Handler
		h.cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.Response.Header.Set("Alt-Svc", h3AltSvc)
			handler(ctx)
		}
	}
}

func blockingHandler(cfg *CfgHttp, handler fasthttp.RequestHandler) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		addr := extractClientIP(ctx)

		if cfg.Allow(ctx, addr) || !cfg.Deny(ctx, addr) {
			handler(ctx)
			return
		}

		logs.DebugLog(addr, ctx.Request.Header.String(), cfg)
		ctx.Error(cfg.Mess, fasthttp.StatusForbidden)
	}
}

// Run starting http or https server according to secure
// certFile and keyFile are paths to TLS certificate and key files for https server
func (h *HttpGo) Run(secure bool, certFile, keyFile string) error {

	h.apis.Https = secure
	h.apis.StartTime = time.Now()
	if h.h3Server != nil {
		go func() {
			logs.StatusLog("starting HTTP3 server")
			if err := h.h3Server.ListenAndServe(); err != nil {
				logs.ErrorLog(err, "HTTP/3 server stopped")
			}
		}()
		//start upstream
		go func() {
			logs.StatusLog("starting HTTP3 server upstream")
			listener, err := reuseport.Listen("tcp4", h.cfg.HTTP3Proxy.UpstreamAddr)
			if err != nil {
				logs.Fatal(err)
			}
			if err := h.cfg.Server.Serve(listener); err != nil {
				logs.ErrorLog(err, "HTTP/3 upstream stopped")
			}
		}()
	}

	//todo change parameters type on
	go h.listenOnShutdown()
	if secure {
		return h.mainServer.ServeTLS(h.listener, certFile, keyFile)
	}

	return h.mainServer.Serve(h.listener)
}

// listenOnShutdown implement correct shutdown server
func (h *HttpGo) listenOnShutdown() {
	// Was `make(chan os.Signal)` - unbuffered. os/signal's own doc is
	// explicit that Notify's channel "should be buffered" - the runtime
	// delivers a signal by a non-blocking send, so with no buffer and no
	// receiver ready at that exact instant (the goroutine hasn't reached
	// `<-ch` yet, or is busy handling a previous signal), the signal is
	// simply dropped and this shutdown path never runs at all. A capacity
	// of 1 is what signal.Notify's own docs recommend.
	ch := make(chan os.Signal, 1)
	KillSignal := syscall.Signal(h.cfg.KillSignal)
	// syscall.SIGTTIN
	//
	// Was also listed here: syscall.SIGKILL, and syscall.SIGINT twice.
	// SIGKILL can never be caught, blocked, or ignored by any process (both
	// POSIX and Go's own os/signal docs say so) - registering it with
	// Notify is not an error, it just never does anything, which reads as
	// "this process handles SIGKILL gracefully" when it can't possibly.
	// SIGINT was listed twice for no effect either way; removed the
	// duplicate.
	signal.Notify(ch, KillSignal, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	logs.StatusLog("Shutdown service starting %v on signal '%v'", time.Now(), KillSignal)
	signShut := <-ch

	logs.StatusLog("Shutdown service get signal: " + signShut.String())
	close(h.broadcast)

	// Was `ctx, _ := context.WithTimeout(...)` - discarding the cancel func
	// is a real leak (staticcheck/go vet both flag this as "lostcancel"):
	// the context's internal timer keeps running until the 5s deadline
	// fires regardless, instead of being released the moment shutdown
	// actually finishes early.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := h.rdServerShutdownWithContext(ctx); err != nil {
		logs.ErrorLog(err)
	}

	if err := h.h3ServerShutdownWithContext(ctx); err != nil {
		logs.ErrorLog(err)
	}

	if err := h.mainServer.ShutdownWithContext(ctx); err != nil {
		logs.ErrorLog(err)
	}

}

func (h *HttpGo) rdServerShutdownWithContext(ctx context.Context) error {
	if h.rdServer != nil {
		return h.rdServer.ShutdownWithContext(ctx)
	}

	return nil
}

func (h *HttpGo) h3ServerShutdownWithContext(ctx context.Context) error {
	if h.h3Server != nil {
		return h.h3Server.Shutdown(ctx)
	}
	return nil
}

const separator = "/"

func isLocalRedirect(ip string) bool {
	const delim = ":"
	return strings.HasPrefix(ip, delim) || strings.HasPrefix(ip, separator)
}

func isLocalDirectory(ip string) bool {
	return strings.HasPrefix(ip, separator) || strings.HasSuffix(ip, separator)
}

// webAuthnManager resolves the app-wide auth manager (set per-request by
// apis.Apis.Handler via auth.SetAuthManager) and asserts it to
// auth.FncPasskey. Registering the four /webauthn/* routes below
// unconditionally in createAdminRoutes - rather than in a per-project
// generated routes.go - means passkey support is "standard equipment" on
// every NewHttpgo() server, exactly like /httpgo/cfg/* already is: a
// server wired with a plain *auth.AuthBearer simply has these four routes
// return ErrRouteForbidden (no passkey support configured), while one
// wired with a *auth.WebAuthnPasskey (see auth/passkey.go) gets working
// endpoints with zero extra route registration in the project itself.
func webAuthnManager(ctx *fasthttp.RequestCtx) (auth.FncPasskey, error) {
	x, ok := auth.GetAuthManager(ctx)
	if !ok {
		return nil, ErrRouteForbidden
	}
	p, ok := x.(auth.FncPasskey)
	if !ok {
		return nil, ErrRouteForbidden
	}

	return p, nil
}

// cookieValue reads the named cookie's raw value off the incoming request.
// Not passkey-specific - any handler needing to read one of its own cookies
// back off the request can call this the same way the passkey endpoints
// below do (cookieValue(ctx, auth.PasskeySessionCookie)).
func cookieValue(ctx *fasthttp.RequestCtx, name string) string {
	return string(ctx.Request.Header.Cookie(name))
}

// setCookie writes name/value onto the response, scoped to path and valid
// for maxAge from now. HttpOnly/Secure/SameSite=Strict is a fixed policy
// here, not a parameter - every current caller wants exactly that posture
// for a short-lived, server-correlated session id; a cookie needing looser
// attributes (e.g. readable by JS, cross-site) would need its own variant
// rather than passing different flags into this one.
func setCookie(ctx *fasthttp.RequestCtx, name, value, path string, maxAge time.Duration) {
	c := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(c)
	c.SetKey(name)
	c.SetValue(value)
	c.SetPath(path)
	c.SetHTTPOnly(true)
	c.SetSecure(true)
	c.SetSameSite(fasthttp.CookieSameSiteStrictMode)
	c.SetMaxAge(int(maxAge.Seconds()))
	ctx.Response.Header.SetCookie(c)
}

// clearCookie tells the browser to drop the named cookie.
func clearCookie(ctx *fasthttp.RequestCtx, name string) {
	// DelClientCookie (not DelCookie) - this tells the *browser* to drop the
	// cookie it already has; DelCookie only removes it from this response's
	// own headers before sending, which wouldn't touch what the browser is
	// already holding.
	ctx.Response.Header.DelClientCookie(name)
}

// handleWebAuthnRegisterBegin -> POST /webauthn/register/begin. Not
// NeedAuth: CheckAndRun would run FncAuth.Auth(ctx) BEFORE this handler
// ever runs, which requires the caller to already hold a valid Bearer
// token - but this route's whole job is registering a passkey for a login
// that may have never authenticated before (a brand new account's very
// first passkey doubles as sign-up). It identifies the account via the
// crud.ParamsLogin ("login") param declared on the route below instead -
// see BeginRegistration's own doc comment in passkey.go for the full
// reasoning, and crud.ParamsLogin's doc comment for an open question about
// whether that param actually gets populated for this request's real
// Content-Type.
func handleWebAuthnRegisterBegin(ctx *fasthttp.RequestCtx) (any, error) {
	p, err := webAuthnManager(ctx)
	if err != nil {
		return nil, err
	}
	login := GetValue[string](ctx, &crud.ParamsLogin)

	challenge, err := p.BeginRegistration(ctx, login)
	if err != nil {
		return nil, ErrWrongParamsList
	}

	setCookie(ctx, auth.PasskeySessionCookie, challenge.SessionID, "/webauthn/", auth.PasskeySessionTTL)
	return challenge.Body, nil
}

// handleWebAuthnRegisterFinish -> POST /webauthn/register/finish. Also not
// NeedAuth - FinishRegistration was never gated by Bearer auth in the
// first place (it resolves its caller from the ceremony's session cookie,
// see takeSession in passkey.go), so the route-level NeedAuth this had
// before never actually did anything for it either.
func handleWebAuthnRegisterFinish(ctx *fasthttp.RequestCtx) (any, error) {
	p, err := webAuthnManager(ctx)
	if err != nil {
		return nil, err
	}

	sessionID := cookieValue(ctx, auth.PasskeySessionCookie)
	clearCookie(ctx, auth.PasskeySessionCookie)
	return p.FinishRegistration(ctx, sessionID, ctx.PostBody())
}

// handleWebAuthnLoginBegin -> POST /webauthn/login/begin (no auth yet - the
// request body is {"login": "..."}, that's the whole point of this route).
func handleWebAuthnLoginBegin(ctx *fasthttp.RequestCtx) (any, error) {
	p, err := webAuthnManager(ctx)
	if err != nil {
		return nil, err
	}

	login := GetValue[string](ctx, &crud.ParamsLogin)

	challenge, err := p.BeginLogin(ctx, login)
	if err != nil {
		return crud.ErrWrongParamsResult(err.Error(), login)
	}

	setCookie(ctx, auth.PasskeySessionCookie, challenge.SessionID, "/webauthn/", auth.PasskeySessionTTL)
	return challenge.Body, nil
}

// handleWebAuthnLoginFinish -> POST /webauthn/login/finish (no auth yet).
func handleWebAuthnLoginFinish(ctx *fasthttp.RequestCtx) (any, error) {
	p, err := webAuthnManager(ctx)
	if err != nil {
		return nil, err
	}

	sessionID := cookieValue(ctx, auth.PasskeySessionCookie)
	clearCookie(ctx, auth.PasskeySessionCookie)
	return p.FinishLogin(ctx, sessionID, ctx.PostBody())
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
		ShowVersion: {
			Fnc:
			// HandleLogServer show status httpgo
			// @/api/version/
			func(*fasthttp.RequestCtx) (any, error) {
				return GetAppTitle(cfg.Server.Name), nil
			},
			Desc: "view version server",
		},
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
		// Passkey (WebAuthn) ceremony endpoints - see webAuthnManager's doc
		// comment above for why these live here rather than in a
		// per-project routes.go, and auth/passkey.go for what each Fnc
		// actually does. Exact paths passkey-auth.js/user.js already call.
		//
		// Neither register route is NeedAuth: register/begin identifies the
		// account via the "login" param below rather than an existing
		// Bearer token (see handleWebAuthnRegisterBegin's doc comment for
		// why), and register/finish never read the Bearer-auth-derived
		// ctx.UserValue at all - it resolves its caller from the ceremony's
		// own session cookie.
		"/webauthn/register/begin": {
			Desc:   "begin passkey registration (login param identifies the account; a login with no existing passkey is registered as a new account)",
			Fnc:    handleWebAuthnRegisterBegin,
			Method: POST,
			Params: []InParam{crud.ParamsLogin},
		},
		"/webauthn/register/finish": {
			Desc:   "finish passkey registration",
			Fnc:    handleWebAuthnRegisterFinish,
			Method: POST,
		},
		"/webauthn/login/begin": {
			Desc:   "begin passkey login (no auth yet)",
			Fnc:    handleWebAuthnLoginBegin,
			Method: POST,
			Params: []InParam{crud.ParamsLogin},
		},
		"/webauthn/login/finish": {
			Desc:   "finish passkey login (no auth yet)",
			Fnc:    handleWebAuthnLoginFinish,
			Method: POST,
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
		"/httpgo/store/sse": {
			Desc:           " # store",
			Fnc:            HandleNoticeSSE,
			IsServerEvents: true,
			WithCors:       true,
			Params: []InParam{
				crud.ParamsIDReq,
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

// fastHTTPLogger wrap logging server
type fastHTTPLogger struct {
	logs.LogsType
}

func (log *fastHTTPLogger) Printf(mess string, args ...any) {

	if strings.Contains(mess, "error") {
		if slices.ContainsFunc(args, func(a any) bool {
			err, ok := a.(error)
			return ok && (isTLSError(err) || isReadError(err) || isHeaderError(err) || isMPFBodyError(err) || isUnsupportedContent(err))
		}) {
			// These are the well-known "client did something a bit odd"
			// noise fasthttp itself logs on essentially every deployment
			// (a bad TLS handshake probe, a client that hung up mid-request,
			// ...) - genuinely not worth a log line at any level in the
			// common case. Kept fully silent as before; if this class ever
			// needs to be *investigated* rather than just ignored, this is
			// the one spot to add a DebugLog(mess, args...) back in.
		} else if strings.Contains(mess, "serving connection") {
			logs.DebugLog(fmt.Sprintf(mess, args...))
			if len(args) > 2 {
				logs.DebugLog("%#v", args[2])
			}

		} else {
			logs.ErrorLog(errors.New(mess), args...)
		}
	} else {
		logs.DebugLog(append([]any{mess}, args...)...)
	}
}

func renderError(ctx *fasthttp.RequestCtx, err error) {
	switch err {
	case fasthttp.ErrBodyTooLarge:
		ctx.SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
	case fasthttp.ErrNoMultipartForm, fasthttp.ErrNoArgValue:
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.SetBodyString(err.Error())
	default:
		switch {
		case isReadError(err):
			logs.DebugLog(err)
			ctx.SetStatusCode(fasthttp.StatusExpectationFailed)

		case isTLSError(err):
			logs.DebugLog("%v", err)
			ctx.SetStatusCode(fasthttp.StatusConflict)

		case isHeaderError(err):
			logs.DebugLog("%v", err)
			ctx.SetStatusCode(fasthttp.StatusRequestHeaderFieldsTooLarge)

		case isMPFBodyError(err) || isUnsupportedContent(err):
			logs.DebugLog("%v", err)
			ctx.SetStatusCode(fasthttp.StatusNotAcceptable)

		default:
			logs.ErrorLog(err, ctx.String())
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}
}

func isTLSError(err error) bool {
	return strings.Contains(err.Error(), "tls: ")
}

func isReadError(err error) bool {
	return strings.Contains(err.Error(), "read: ") || strings.Contains(err.Error(), "unexpected EOF")
}

func isHeaderError(err error) bool {
	return strings.Contains(err.Error(), "error when reading request headers")
}
func isMPFBodyError(err error) bool {
	return strings.Contains(err.Error(), "cannot read multipart/form-data body")
}
func isUnsupportedContent(err error) bool {
	return strings.Contains(err.Error(), "unsupported Content-Encoding")
}
