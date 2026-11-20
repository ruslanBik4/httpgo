/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeCert(t *testing.T, dir, cn string, serial int64) {
	t.Helper()
	writeCertNames(t, dir, cn, serial, cn)
}

// writeCertNames writes a certificate valid for names (they may be wildcards)
func writeCertNames(t *testing.T, dir, cn string, serial int64, names ...string) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tpl := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: cn}, DNSNames: names,
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(90 * 24 * time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	kb, _ := x509.MarshalECPrivateKey(key)
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "fullchain.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600)
	os.WriteFile(filepath.Join(dir, "privkey.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}), 0o600)
}

func serial(t *testing.T, m *certMaps, host string) int64 {
	c, err := m.getCertificate(&tls.ClientHelloInfo{ServerName: host})
	if err != nil || c == nil {
		t.Fatalf("no certificate for %s: %v", host, err)
	}
	x, _ := x509.ParseCertificate(c.Certificate[0])
	return x.SerialNumber.Int64()
}

// the layout of Let's Encrypt: live/<domain>/ links into archive/<domain>/
func TestCertReload(t *testing.T) {
	root := t.TempDir()
	live, archive := filepath.Join(root, "live"), filepath.Join(root, "archive", "shop.example.com")
	writeCert(t, archive, "shop.example.com", 1)
	os.MkdirAll(filepath.Join(live, "shop.example.com"), 0o700)
	for _, f := range []string{"fullchain.pem", "privkey.pem"} {
		os.Symlink(filepath.Join("..", "..", "archive", "shop.example.com", f), filepath.Join(live, "shop.example.com", f))
	}
	os.WriteFile(filepath.Join(live, "README"), []byte("certbot"), 0o644)
	other := filepath.Join(root, "sites")
	writeCert(t, filepath.Join(other, "localhost"), "localhost", 7)

	m := newCertMaps(other, live)
	if err := m.loadCertificates(); err != nil {
		t.Fatal(err)
	}
	if s := serial(t, m, "shop.example.com"); s != 1 {
		t.Fatalf("serial %d", s)
	}
	// certbot renews: new files behind the same links
	writeCert(t, archive, "shop.example.com", 2)
	if s := serial(t, m, "shop.example.com"); s != 1 {
		t.Fatalf("changed before the reload: %d", s)
	}
	if err := m.reload(); err != nil {
		t.Fatal(err)
	}
	if s := serial(t, m, "shop.example.com"); s != 2 {
		t.Fatalf("not reloaded: %d", s)
	}
	if s := serial(t, m, "localhost"); s != 7 {
		t.Fatalf("other dir: %d", s)
	}
	// a read that fails (all the certificates gone for a moment): the old ones stay
	os.RemoveAll(root)
	if err := m.reload(); err == nil {
		t.Fatal("reload of nothing")
	}
	if s := serial(t, m, "shop.example.com"); s != 2 {
		t.Fatalf("lost: %d", s)
	}
}

func TestCertReloadEnv(t *testing.T) {
	for v, want := range map[string]time.Duration{"": DefaultCertReload, "6h": 6 * time.Hour, "0": 0, "bad": DefaultCertReload} {
		t.Setenv(CertReloadEnv, v)
		if got := certReload(); got != want {
			t.Errorf("%q: %v, want %v", v, got, want)
		}
	}
}

// a subdomain is served by the certificate that names it, not by the directory of the exact name
func TestCertSubdomains(t *testing.T) {
	root := t.TempDir()
	writeCert(t, filepath.Join(root, "default"), "default", 1)
	// a wildcard certificate lives in the directory of the base domain
	writeCertNames(t, filepath.Join(root, "example.com"), "example.com", 2, "example.com", "*.example.com")
	// a directory named after the wildcard, the certificate has the names too
	writeCertNames(t, filepath.Join(root, "*.shop.org"), "shop.org", 3, "shop.org", "*.shop.org")
	// one certificate with several concrete names
	writeCertNames(t, filepath.Join(root, "multi"), "a.multi.net", 4, "a.multi.net", "b.multi.net")

	m := newCertMaps(root)
	if err := m.loadCertificates(); err != nil {
		t.Fatal(err)
	}
	for host, want := range map[string]int64{
		"example.com":           2,
		"EXAMPLE.com":           2,
		"api.example.com":       2,
		"api.example.com.":      2, // absolute name
		"a.b.example.com":       2, // nested: the nearest wildcard up the tree
		"shop.org":              3,
		"www.shop.org":          3,
		"b.multi.net":           4,
		"unknown.net":           1, // fallback
		"":                      1, // no SNI
		"c.multi.net":           1, // not listed in the certificate
		"example.com.evil.test": 1, // a suffix is not a parent
	} {
		if got := serial(t, m, host); got != want {
			t.Errorf("%q: serial %d, want %d", host, got, want)
		}
	}

	// the names are indexed again after a reload
	writeCertNames(t, filepath.Join(root, "example.com"), "example.com", 5, "example.com", "*.example.com")
	if err := m.reload(); err != nil {
		t.Fatal(err)
	}
	if got := serial(t, m, "api.example.com"); got != 5 {
		t.Errorf("after reload: serial %d, want 5", got)
	}
}
