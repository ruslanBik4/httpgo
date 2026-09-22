/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
package httpGo

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/auth"
)

// --- regForwardedFor --------------------------------------------------

func Test_regForwardedFor(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		matches bool
	}{
		{"simple for=", `for=192.0.2.60;proto=http;by=203.0.113.43`, "192.0.2.60", true},
		// The bracket alternative (\[[^\]]+\]) matches greedily up to the
		// FIRST closing bracket and stops there - it does not also grab a
		// trailing :port outside the brackets, even though the quoted value
		// as a whole includes one. Verified against Go's actual regexp
		// engine (not just read off the pattern) since this is easy to get
		// wrong by eye: RE2/Go regexp uses leftmost-first alternation, so
		// once \[[^\]]+\] matches, the permissive second alternative -
		// which WOULD consume the trailing :port - is never tried.
		{"quoted bracketed IPv6 with port - captures only the bracketed part", `for="[2001:db8:cafe::17]:4711"`, "[2001:db8:cafe::17]", true},
		{"case-insensitive scheme name", `For=192.0.2.1`, "192.0.2.1", true},
		{"optional whitespace after for=", `for= 192.0.2.1`, "192.0.2.1", true},
		{"no for= at all", `by=203.0.113.43;proto=http`, "", false},
		{"first hop of a multi-hop header", `for=192.0.2.60, for=198.51.100.17`, "192.0.2.60", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := regForwardedFor.FindStringSubmatch(tt.header)
			if !tt.matches {
				assert.Nil(t, m)
				return
			}
			if assert.NotNil(t, m) {
				assert.Equal(t, tt.want, m[1])
			}
		})
	}
}

// --- extractClientIP ----------------------------------------------------

func Test_extractClientIP_XForwardedFor(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"single address", "203.0.113.5", "203.0.113.5"},
		{"multi-hop, takes first", "203.0.113.5, 70.41.3.18, 150.172.238.178", "203.0.113.5"},
		{"trims surrounding whitespace", "  203.0.113.5  , 70.41.3.18", "203.0.113.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.Set("X-Forwarded-For", tt.header)
			assert.Equal(t, tt.want, extractClientIP(ctx))
		})
	}
}

func Test_extractClientIP_Forwarded(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Forwarded", `for=192.0.2.60;proto=http;by=203.0.113.43`)
	assert.Equal(t, "192.0.2.60", extractClientIP(ctx))
}

func Test_extractClientIP_XProxyUserIp(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("X-ProxyUser-Ip", "198.51.100.1")
	assert.Equal(t, "198.51.100.1", extractClientIP(ctx))
}

func Test_extractClientIP_Precedence(t *testing.T) {
	t.Run("X-Forwarded-For wins over Forwarded and X-ProxyUser-Ip", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("X-Forwarded-For", "203.0.113.5")
		ctx.Request.Header.Set("Forwarded", `for=192.0.2.60`)
		ctx.Request.Header.Set("X-ProxyUser-Ip", "198.51.100.1")
		assert.Equal(t, "203.0.113.5", extractClientIP(ctx))
	})

	t.Run("Forwarded wins over X-ProxyUser-Ip when no X-Forwarded-For", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Forwarded", `for=192.0.2.60`)
		ctx.Request.Header.Set("X-ProxyUser-Ip", "198.51.100.1")
		assert.Equal(t, "192.0.2.60", extractClientIP(ctx))
	})
}

// Test_extractClientIP_RemoteAddrFallback exercises the final branch (no
// X-Forwarded-For/Forwarded/X-ProxyUser-Ip header at all) by giving the ctx
// a real remote address via fasthttp.RequestCtx.Init - the documented way to
// build a *RequestCtx for tests without a live network connection (a bare
// &fasthttp.RequestCtx{} has no underlying net.Conn, so ctx.Conn() on it
// would panic). Not independently re-verified against fasthttp's source in
// this sandbox (no network access to fetch the dependency), but Init's
// signature and purpose are stable, long-documented fasthttp public API.
func Test_extractClientIP_RemoteAddrFallback(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	addr := &net.TCPAddr{IP: net.ParseIP("203.0.113.99"), Port: 12345}
	ctx.Init(&fasthttp.Request{}, addr, nil)

	assert.Equal(t, addr.String(), extractClientIP(ctx))
}

// --- filterIPs ----------------------------------------------------------

func Test_filterIPs(t *testing.T) {
	tests := []struct {
		name string
		cur  []string
		rm   []string
		want []string
	}{
		{"removes a present entry", []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{"no-op for an absent entry", []string{"a", "b"}, []string{"z"}, []string{"a", "b"}},
		{"removes multiple entries", []string{"a", "b", "c", "d"}, []string{"b", "d"}, []string{"a", "c"}},
		{"empty removal list changes nothing", []string{"a", "b"}, []string{}, []string{"a", "b"}},
		{"removing everything leaves empty", []string{"a", "b"}, []string{"a", "b"}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// filterIPs reuses curIPs' backing array (curIPs[:0]) - pass a
			// fresh copy each run so one case can't corrupt another's input.
			cur := append([]string(nil), tt.cur...)
			got := filterIPs(cur, tt.rm)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- isLocalRedirect / isLocalDirectory ----------------------------------

func Test_isLocalRedirect(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"leading colon (port only)", ":8080", true},
		{"leading slash (path)", "/some/path", true},
		{"neither", "example.com", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isLocalRedirect(tt.ip))
		})
	}
}

func Test_isLocalDirectory(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"leading slash", "/var/www", true},
		{"trailing slash", "static/", true},
		{"neither", "example.com", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isLocalDirectory(tt.ip))
		})
	}
}

// --- generic cookie helpers ------------------------------------------------
//
// passkey.go (auth package) no longer knows anything about cookies at all -
// it only hands back/accepts a plain session-id string. cookieValue/
// setCookie/clearCookie are the ONLY place that actually reads/writes any
// cookie on a real fasthttp request/response - not passkey-specific, so
// they're exercised directly against fasthttp.RequestCtx with an arbitrary
// name/path/ttl, plus the exact passkey call shape (createAdminRoutes'
// handlers pass auth.PasskeySessionCookie/"/webauthn/"/auth.PasskeySessionTTL)
// as its own case.

func Test_cookieValue(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.SetCookie("some_cookie", "abc123")
		assert.Equal(t, "abc123", cookieValue(ctx, "some_cookie"))
	})

	t.Run("missing", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		assert.Empty(t, cookieValue(ctx, "some_cookie"))
	})

	t.Run("passkey session cookie", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.SetCookie(auth.PasskeySessionCookie, "abc123")
		assert.Equal(t, "abc123", cookieValue(ctx, auth.PasskeySessionCookie))
	})
}

func Test_setCookie(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	setCookie(ctx, "some_cookie", "some-value", "/some/path/", 90*time.Second)

	c := &fasthttp.Cookie{}
	c.SetKey("some_cookie")
	if assert.True(t, ctx.Response.Header.Cookie(c), "setCookie must set a cookie the response actually carries") {
		assert.Equal(t, "some-value", string(c.Value()))
		assert.Equal(t, "/some/path/", string(c.Path()))
		assert.True(t, c.HTTPOnly())
		assert.True(t, c.Secure())
		assert.Equal(t, fasthttp.CookieSameSiteStrictMode, c.SameSite())
		assert.Equal(t, 90, c.MaxAge())
	}
}

func Test_setCookie_PasskeySession(t *testing.T) {
	// Exactly how handleWebAuthnRegisterBegin/handleWebAuthnLoginBegin call
	// setCookie - pins down that the generic helper reproduces the old
	// passkey-specific setPasskeySessionCookie's behavior exactly.
	ctx := &fasthttp.RequestCtx{}
	setCookie(ctx, auth.PasskeySessionCookie, "new-session-id", "/webauthn/", auth.PasskeySessionTTL)

	c := &fasthttp.Cookie{}
	c.SetKey(auth.PasskeySessionCookie)
	if assert.True(t, ctx.Response.Header.Cookie(c)) {
		assert.Equal(t, "new-session-id", string(c.Value()))
		assert.Equal(t, "/webauthn/", string(c.Path()))
		assert.Equal(t, int(auth.PasskeySessionTTL.Seconds()), c.MaxAge())
	}
}

func Test_clearCookie(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	clearCookie(ctx, "some_cookie")

	// DelClientCookie writes a cookie with an empty value and a max-age of
	// 0/negative expiry - fasthttp's own way of telling the browser to drop
	// it - rather than simply omitting a Set-Cookie header entirely.
	c := &fasthttp.Cookie{}
	c.SetKey("some_cookie")
	if assert.True(t, ctx.Response.Header.Cookie(c), "clearCookie must still write a Set-Cookie header (the deletion instruction), not merely omit one") {
		assert.Empty(t, string(c.Value()))
	}
}

// --- fastHTTPLogger's error-classifying predicates -----------------------

func Test_isTLSError(t *testing.T) {
	assert.True(t, isTLSError(errors.New("tls: bad certificate")))
	assert.False(t, isTLSError(errors.New("some other error")))
}

func Test_isReadError(t *testing.T) {
	assert.True(t, isReadError(errors.New("read: connection reset by peer")))
	assert.True(t, isReadError(errors.New("unexpected EOF")))
	assert.False(t, isReadError(errors.New("some other error")))
}

func Test_isHeaderError(t *testing.T) {
	assert.True(t, isHeaderError(errors.New("error when reading request headers: too large")))
	assert.False(t, isHeaderError(errors.New("some other error")))
}

func Test_isMPFBodyError(t *testing.T) {
	assert.True(t, isMPFBodyError(errors.New("cannot read multipart/form-data body: unexpected EOF")))
	assert.False(t, isMPFBodyError(errors.New("some other error")))
}

func Test_isUnsupportedContent(t *testing.T) {
	assert.True(t, isUnsupportedContent(errors.New("unsupported Content-Encoding: br")))
	assert.False(t, isUnsupportedContent(errors.New("some other error")))
}
