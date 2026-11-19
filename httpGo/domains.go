/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/ruslanBik4/gotools"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/logs"
)

var (
	ErrDomainEmpty  = errors.New("subdomain is empty")
	ErrDomainWrong  = errors.New("subdomain has wrong symbols")
	ErrTargetEmpty  = errors.New("target of subdomain is empty")
	ErrDomainExists = errors.New("subdomain already exists")
)

// Domains maps a subdomain onto its target, the "Domains" section of httpgo.yml:
//
//	lexx: /www/lexx/                  # local directory - serve its files
//	dev: :8080                        # local port - temporary redirect to <host>:8080
//	shop: https://shop.example.com/   # any other - permanent redirect to it
//
// A subdomain matches the leading labels of the request host: "lexx" matches
// "lexx.example.com", "toto.kid" matches "toto.kid.example.com"; the longest key wins.
//
// It is safe for concurrent use: requests read an immutable snapshot (one atomic
// load, no lock), Add/Set/Remove publish a new copy (copy-on-write) - subdomains
// can be inserted while the server runs (e.g. for started containers).
type Domains struct {
	mu   sync.Mutex // serializes writers only
	list atomic.Pointer[map[string]string]
}

// NewDomains creates Domains from map subdomain -> target (map is copied)
func NewDomains(list map[string]string) *Domains {
	d := &Domains{}
	d.reset(list)

	return d
}

// reset replaces all subdomains by list
func (d *Domains) reset(list map[string]string) {
	m := make(map[string]string, len(list))
	for sub, target := range list {
		m[normalizeSubdomain(sub)] = target
	}

	d.mu.Lock()
	d.list.Store(&m)
	d.mu.Unlock()
}

func (d *Domains) load() map[string]string {
	if d == nil {
		return nil
	}
	if m := d.list.Load(); m != nil {
		return *m
	}

	return nil
}

// Len returns count of subdomains
func (d *Domains) Len() int {
	return len(d.load())
}

// Get returns target of subdomain
func (d *Domains) Get(sub string) (string, bool) {
	target, ok := d.load()[normalizeSubdomain(sub)]
	return target, ok
}

// All returns copy of all subdomains
func (d *Domains) All() map[string]string {
	return maps.Clone(d.load())
}

// Add inserts new subdomain, returns ErrDomainExists if it is present already
// (use Set to replace it)
func (d *Domains) Add(sub, target string) error {
	return d.put(sub, target, false)
}

// Set inserts new subdomain or replaces target of existing one
func (d *Domains) Set(sub, target string) error {
	return d.put(sub, target, true)
}

func (d *Domains) put(sub, target string, replace bool) error {
	sub = normalizeSubdomain(sub)
	if err := validSubdomain(sub); err != nil {
		return errors.Wrapf(err, "'%s'", sub)
	}
	if target = strings.TrimSpace(target); target == "" {
		return errors.Wrapf(ErrTargetEmpty, "'%s'", sub)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	cur := d.load()
	if old, ok := cur[sub]; ok {
		if !replace {
			return errors.Wrapf(ErrDomainExists, "'%s' -> %s", sub, old)
		}
		if old == target {
			return nil
		}
	}

	m := make(map[string]string, len(cur)+1)
	maps.Copy(m, cur)
	m[sub] = target
	d.list.Store(&m)
	logs.DebugLog("subdomain %s -> %s", sub, target)

	return nil
}

// Remove deletes subdomain, returns false if it is absent
func (d *Domains) Remove(sub string) bool {
	sub = normalizeSubdomain(sub)

	d.mu.Lock()
	defer d.mu.Unlock()

	cur := d.load()
	if _, ok := cur[sub]; !ok {
		return false
	}

	m := maps.Clone(cur)
	delete(m, sub)
	d.list.Store(&m)
	logs.DebugLog("subdomain %s removed", sub)

	return true
}

// Match finds target for host (port is ignored), the longest matched subdomain wins
func (d *Domains) Match(host string) (sub, target string, ok bool) {
	list := d.load()
	if len(list) == 0 {
		return "", "", false
	}

	host = hostWithoutPort(host)
	// candidates are leading labels of host, from the longest:
	// "a.b.example.com" -> "a.b.example.com", "a.b.example", "a.b", "a"
	for end := len(host); end > 0; end = strings.LastIndexByte(host[:end], '.') {
		if target, ok := list[host[:end]]; ok {
			return host[:end], target, true
		}
	}

	return "", "", false
}

// Handler returns handler for split requests by subdomains,
// other requests pass to next handler.
// listener is used to detect that local port target is this server itself.
// Subdomains inserted later (Add/Set) work without re-creating handler.
func (d *Domains) Handler(listener net.Listener, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	ownAddr := ""
	if listener != nil {
		ownAddr = listener.Addr().String()
	}

	return func(ctx *fasthttp.RequestCtx) {
		host := gotools.BytesToString(ctx.Host())
		_, target, ok := d.Match(host)
		if !ok || host == target {
			next(ctx)
			return
		}

		switch {
		case isLocalDirectory(target):
			sendDomainFile(ctx, target)

		case !isLocalRedirect(target):
			logs.DebugLog("redirect", target)
			ctx.Redirect(target, fasthttp.StatusMovedPermanently)

		// local port of this server - serve it
		case strings.HasSuffix(ownAddr, target):
			next(ctx)

		default:
			// keep path & query, host:port of other local server.
			// Temporary: browsers cache 301 forever, but local ports change
			// (e.g. container of project rebuilt on another port)
			url := fmt.Sprintf("%s://%s%s%s",
				ctx.URI().Scheme(), hostWithoutPort(host), target, ctx.RequestURI())
			logs.DebugLog("redirect:", url)
			ctx.Redirect(url, fasthttp.StatusTemporaryRedirect)
		}
	}
}

func sendDomainFile(ctx *fasthttp.RequestCtx, dir string) {
	p := gotools.BytesToString(ctx.URI().Path())
	if p == "" || p == "/" {
		p = "index.html"
	}
	// path.Join cleans "..", Path() is normalized by fasthttp already
	fileName := path.Join(".", dir, p)
	logs.DebugLog(dir, p)
	if err := ctx.Response.SendFile(fileName); err != nil {
		logs.ErrorLog(err, ctx.String())
		return
	}

	ct, _ := views.GetContentType(ctx, fileName)
	ctx.Response.Header.SetContentType(ct)
}

// MarshalYAML implements yaml.Marshaler
func (d *Domains) MarshalYAML() (any, error) {
	return d.load(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Domains) UnmarshalYAML(node *yaml.Node) error {
	var m map[string]string
	if err := node.Decode(&m); err != nil {
		return err
	}
	d.reset(m)

	return nil
}

// MarshalJSON implements json.Marshaler
func (d *Domains) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.load())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Domains) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	d.reset(m)

	return nil
}

func normalizeSubdomain(sub string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(sub), "."))
}

// validSubdomain checks sub as DNS labels: letters, digits, '-', '_' and dots between labels
func validSubdomain(sub string) error {
	if sub == "" {
		return ErrDomainEmpty
	}
	for _, label := range strings.Split(sub, ".") {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return ErrDomainWrong
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
				return ErrDomainWrong
			}
		}
	}

	return nil
}

func hostWithoutPort(host string) string {
	if i := strings.LastIndexByte(host, ':'); i >= 0 && !strings.Contains(host[i:], "]") {
		return host[:i]
	}

	return host
}
