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
	"net/http"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/quic-go/quic-go/http3"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/gotools"
	. "github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/apis/crud"
	"github.com/ruslanBik4/httpgo/auth"
	"github.com/ruslanBik4/httpgo/httpGo/http2"
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
	h2Server   *http2.Server
	// h3Conn is the UDP socket of HTTP/3, bound in setHTTP3 so that Alt-Svc
	// is only advertised when HTTP/3 is really reachable
	h3Conn net.PacketConn
	// h3Upstream is the private loopback listener the HTTP/3 proxy forwards to
	h3Upstream net.Listener
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

	// No ResetUserValues wrapper for HTTP/2 any more: every HTTP/2 stream gets
	// a freshly reset RequestCtx, released only after its handler returns.
	// (It would also erase the key http2.Push uses to find the stream.)
	cfg.Server.Handler = apis.Handler

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

	// HTTP/2 (ALPN "h2"). Safe before the handler wrapping below: the
	// HTTP/2 server resolves cfg.Server.Handler per request.
	h2Server := NewHTTP2Server(cfg)
	if h2Server == nil {
		// never offer h2 in ALPN without a server for it
		withoutALPN(listener, H2TLSProto)
	}

	h := &HttpGo{
		mainServer: cfg.Server,
		listener:   listener,
		broadcast:  make(chan *string),
		apis:       apis,
		cfg:        cfg,
		store:      store,
		h2Server:   h2Server,
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
	if !h.cfg.HTTP3 {
		return
	}

	var tlsCfg *tls.Config
	fPort := ":443"
	if ln, ok := h.listener.(LnMultiCerts); ok {
		tlsCfg = ln.tlsCfg.Clone()
		fPort = ln.fPort
	} else {
		certMaps := newCertMaps() // $HTTPGO_CERT_DIR (list) or /etc/ssl/sites
		if err := certMaps.loadCertificates(); err != nil {
			// HTTP/3 is opt-in and the TCP server works without it:
			// disable only this half instead of serving UDP with a broken TLS config.
			logs.ErrorLog(err, "HTTP/3 disabled: failed to load TLS certificates")
			return
		}
		tlsCfg = &tls.Config{GetCertificate: certMaps.getCertificate}
	}
	tlsCfg.NextProtos = []string{"h3"}
	if h.cfg.HTTP3Proxy == nil {
		h.cfg.HTTP3Proxy = &HTTP3ProxyConfig{}
	}

	// Bind both sockets now, not in Run: a busy/unpublished UDP port must
	// switch HTTP/3 off instead of sending browsers to a dead Alt-Svc.
	udp, err := net.ListenPacket("udp", fPort)
	if err != nil {
		logs.ErrorLog(err, "HTTP/3 disabled: cannot listen UDP %s", fPort)
		return
	}

	altSvc, err := h3AltSvc(udp.LocalAddr().String())
	if err != nil {
		logs.ErrorLog(err, "HTTP/3 disabled")
		logs.ErrorLog(udp.Close())
		return
	}

	upstream, err := listenH3Upstream(h.cfg.HTTP3Proxy.UpstreamAddr)
	if err != nil {
		logs.ErrorLog(err, "HTTP/3 disabled: cannot listen upstream %q", h.cfg.HTTP3Proxy.UpstreamAddr)
		logs.ErrorLog(udp.Close())
		return
	}

	// the proxy targets the listener really opened (its port may be dynamic)
	proxyCfg := *h.cfg.HTTP3Proxy
	proxyCfg.UpstreamAddr = upstream.Addr().String()
	h3, err := NewHTTP3Proxy(tlsCfg, fPort, &proxyCfg)
	if err != nil {
		logs.ErrorLog(err, "HTTP/3 disabled")
		logs.ErrorLog(udp.Close())
		logs.ErrorLog(upstream.Close())
		return
	}

	h.h3Server, h.h3Conn, h.h3Upstream = h3, udp, upstream
	logs.StatusLog("[*] HTTP3 multi-site proxy running on %s (udp %s), upstream %s",
		fPort, udp.LocalAddr(), proxyCfg.UpstreamAddr)

	upstreamAddr := proxyCfg.UpstreamAddr
	handler := h.cfg.Server.Handler
	h.cfg.Server.Handler = func(ctx *fasthttp.RequestCtx) {
		if isH3Upstream(ctx, upstreamAddr) {
			// came through the HTTP/3 proxy: the client is on https even
			// though this hop is plain http (redirects build URLs from Scheme())
			ctx.URI().SetScheme("https")
		}
		ctx.Response.Header.Set("Alt-Svc", altSvc)
		handler(ctx)
	}
}

// isH3Upstream reports that ctx arrived on the private HTTP/3 upstream listener.
// Decided by the local address of the connection, not by X-Forwarded-Proto:
// that header can be sent by any client on the public TCP port.
func isH3Upstream(ctx *fasthttp.RequestCtx, upstreamAddr string) bool {
	addr := ctx.LocalAddr()
	return addr != nil && addr.String() == upstreamAddr
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
	h.runHTTP3()

	//todo change parameters type on
	go h.listenOnShutdown()
	if secure {
		return h.mainServer.ServeTLS(h.listener, certFile, keyFile)
	}

	return h.mainServer.Serve(h.listener)
}

// runHTTP3 starts the HTTP/3 server and its private upstream in background;
// the sockets were bound in setHTTP3. No-op when HTTP/3 is off.
func (h *HttpGo) runHTTP3() {
	if h.h3Server == nil {
		return
	}
	go func() {
		logs.StatusLog("starting HTTP3 server")
		if err := h.h3Server.Serve(h.h3Conn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.ErrorLog(err, "HTTP/3 server stopped")
		}
	}()
	go func() {
		logs.StatusLog("starting HTTP3 server upstream on %s", h.h3Upstream.Addr())
		if err := h.cfg.Server.Serve(h.h3Upstream); err != nil {
			logs.ErrorLog(err, "HTTP/3 upstream stopped")
		}
	}()
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

	// Must run alongside mainServer shutdown: fasthttp waits for every open
	// conn but can't close the ones it handed over to HTTP/2 via NextProto.
	if h.h2Server != nil {
		go func() {
			if err := h.h2Server.Shutdown(ctx); err != nil {
				logs.ErrorLog(err)
			}
		}()
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
	if h.h3Server == nil {
		return nil
	}
	err := h.h3Server.Shutdown(ctx)
	// http3.Server.Serve never closes a PacketConn it was given
	if h.h3Conn != nil {
		if errClose := h.h3Conn.Close(); err == nil {
			err = errClose
		}
	}
	return err
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

	logs.StatusLog(p, login)
	// ctx (not context.Background()) - *fasthttp.RequestCtx already
	// implements context.Context (Deadline/Done/Err/Value), so this is
	// httpServer.go's own real request context, passed straight through.
	challenge, err := p.BeginRegistration(ctx, login)
	if err != nil {
		// Was `return nil, ErrWrongParamsList`: every failure of
		// BeginRegistration (empty login, a failed INSERT into Users, a
		// webauthn config problem, ...) reached the browser as the same
		// "400 null", and nothing was logged. Log the real cause, and answer
		// the way login/begin below does (the error text as the body). Note:
		// this route is public, so once it works consider returning a fixed
		// text for store/DB errors and keeping the details in the log only.
		logs.ErrorLog(err, "passkey register/begin", login)
		return err.Error(), ErrWrongParamsList
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
