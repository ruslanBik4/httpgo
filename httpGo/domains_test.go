/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"
)

func TestDomains_Match(t *testing.T) {
	d := NewDomains(map[string]string{
		"lexx":                "/www/lexx/",
		"toto.kid":            "/experimental/toto/",
		"toto":                ":8081",
		"DEV":                 ":8080",
		"createx.gomc.dev":    "/www/Createx/",
		"star-bucks.gomc.dev": "/www/StarBucks/",
	})

	tests := []struct {
		host, sub, target string
		ok                bool
	}{
		{"lexx.gozaurus.com", "lexx", "/www/lexx/", true},
		{"lexx.gozaurus.com:8443", "lexx", "/www/lexx/", true},
		{"toto.kid.gozaurus.com", "toto.kid", "/experimental/toto/", true},
		{"toto.gozaurus.com", "toto", ":8081", true},
		{"dev.localhost", "dev", ":8080", true},
		{"createx.gomc.dev", "createx.gomc.dev", "/www/Createx/", true},
		{"star-bucks.gomc.dev", "star-bucks.gomc.dev", "/www/StarBucks/", true},
		{"lexx", "lexx", "/www/lexx/", true},
		{"lexxx.gozaurus.com", "", "", false},
		{"www.lexx.com", "", "", false},
		{"", "", "", false},
		{"[::1]:443", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			sub, target, ok := d.Match(tt.host)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.sub, sub)
			assert.Equal(t, tt.target, target)
		})
	}
}

func TestDomains_AddSetRemove(t *testing.T) {
	d := NewDomains(nil)
	require.NoError(t, d.Add("Shop", ":9001"))
	assert.True(t, errors.Is(d.Add("shop", ":9002"), ErrDomainExists))
	assert.True(t, errors.Is(d.Add("", ":9002"), ErrDomainEmpty))
	assert.True(t, errors.Is(d.Add("sh op", ":9002"), ErrDomainWrong))
	assert.True(t, errors.Is(d.Add("-shop", ":9002"), ErrDomainWrong))
	assert.True(t, errors.Is(d.Add("new", " "), ErrTargetEmpty))

	target, ok := d.Get("shop")
	assert.True(t, ok)
	assert.Equal(t, ":9001", target)

	require.NoError(t, d.Set("shop", ":9002"))
	target, _ = d.Get("shop")
	assert.Equal(t, ":9002", target)

	assert.True(t, d.Remove("shop"))
	assert.False(t, d.Remove("shop"))
	assert.Equal(t, 0, d.Len())

	var zero Domains
	require.NoError(t, zero.Add("a", ":1"))
	assert.Equal(t, 1, zero.Len())

	var nilD *Domains
	_, _, ok = nilD.Match("a.b")
	assert.False(t, ok)
}

func TestDomains_Concurrent(t *testing.T) {
	d := NewDomains(nil)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				_ = d.Set("p", ":9000")
				d.Remove("p")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				d.Match("p.example.com")
			}
		}()
	}
	wg.Wait()
}

func TestDomains_YAML_JSON(t *testing.T) {
	var cfg struct {
		Domains *Domains `yaml:"Domains"`
	}
	require.NoError(t, yaml.Unmarshal([]byte("Domains:\n  lexx: /www/lexx/\n  dev: :8080\n"), &cfg))
	assert.Equal(t, map[string]string{"lexx": "/www/lexx/", "dev": ":8080"}, cfg.Domains.All())

	b, err := json.Marshal(cfg.Domains)
	require.NoError(t, err)
	d := &Domains{}
	require.NoError(t, json.Unmarshal(b, d))
	assert.Equal(t, cfg.Domains.All(), d.All())

	out, err := yaml.Marshal(cfg.Domains)
	require.NoError(t, err)
	assert.Contains(t, string(out), "lexx: /www/lexx/")
}

type fakeListener struct {
	net.Listener
	addr string
}

func (l fakeListener) Addr() net.Addr { return fakeAddr(l.addr) }

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

func TestDomains_Handler(t *testing.T) {
	d := NewDomains(map[string]string{
		"www": ":443",
		"dev": ":8080",
		"ext": "https://example.org/",
	})
	nextCalled := false
	h := d.Handler(fakeListener{addr: "[::]:443"}, func(*fasthttp.RequestCtx) { nextCalled = true })

	// as server gets it: Host header + origin-form request URI
	run := func(host, uri string) *fasthttp.RequestCtx {
		nextCalled = false
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.SetHost(host)
		ctx.Request.SetRequestURI(uri)
		h(ctx)
		return ctx
	}

	run("www.gozaurus.com", "/a")
	assert.True(t, nextCalled, "own port must be served")

	run("other.gozaurus.com", "/a")
	assert.True(t, nextCalled, "unknown subdomain must be served")

	ctx := run("dev.gozaurus.com", "/a/b?x=1")
	assert.False(t, nextCalled)
	assert.Equal(t, fasthttp.StatusTemporaryRedirect, ctx.Response.StatusCode())
	assert.Equal(t, "http://dev.gozaurus.com:8080/a/b?x=1", string(ctx.Response.Header.Peek("Location")))

	// inserted at runtime, port of the Host must not be doubled
	require.NoError(t, d.Add("proj", ":9100"))
	ctx = run("proj.gozaurus.com:443", "/")
	assert.False(t, nextCalled)
	assert.Equal(t, "http://proj.gozaurus.com:9100/", string(ctx.Response.Header.Peek("Location")))

	ctx = run("ext.gozaurus.com", "/")
	assert.False(t, nextCalled)
	assert.Equal(t, fasthttp.StatusMovedPermanently, ctx.Response.StatusCode())
	assert.Equal(t, "https://example.org/", string(ctx.Response.Header.Peek("Location")))
}
