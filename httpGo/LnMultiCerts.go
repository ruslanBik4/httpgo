/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/valyala/fasthttp/reuseport"

	"github.com/ruslanBik4/httpgo/services"
	"github.com/ruslanBik4/logs"
)

// directory containing subfolders for each domain
const certBaseDir = "/etc/ssl/sites"

// certificate cache
type certMaps struct {
	certMap   map[string]*tls.Certificate
	certMutex sync.RWMutex
}

func newCertMaps() *certMaps {
	return &certMaps{
		certMap: make(map[string]*tls.Certificate),
	}
}

func (m *certMaps) readCertificates(domain string) (tls.Certificate, error) {
	certFile := filepath.Join(certBaseDir, domain, "fullchain.pem")
	keyFile := filepath.Join(certBaseDir, domain, "privkey.pem")

	_, err := os.Stat(certFile)
	if errors.Is(err, os.ErrNotExist) {
		certFile = filepath.Join(certBaseDir, domain, "server.crt")
		_, err = os.Stat(certFile)
	}
	if err != nil {
		return tls.Certificate{}, err
	}

	_, err = os.Stat(keyFile)
	if errors.Is(err, os.ErrNotExist) {
		keyFile = filepath.Join(certBaseDir, domain, "server.key")
		_, err = os.Stat(keyFile)
	}
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(certFile, keyFile)
}

func (m *certMaps) loadCertificates() error {
	m.certMutex.Lock()
	defer m.certMutex.Unlock()

	entries, err := os.ReadDir(certBaseDir)
	if err != nil {
		return err // "Cannot read cert directory: %v")
	}

	symLinks := make(map[string]struct{})
	for _, e := range entries {
		if !e.IsDir() {
			if e.Type().Type() == fs.ModeSymlink {
				// Resolve symlink (one hop)
				targetPath, err := os.Readlink(filepath.Join(certBaseDir, e.Name()))
				if err != nil {
					logs.ErrorLog(err)
					continue
				}
				symLinks[targetPath] = struct{}{}
			}
			continue
		}
		domain := e.Name()
		cert, err := m.readCertificates(domain)
		if err != nil {
			logs.ErrorLog(err, "Load cert for %s failed", domain)
			continue
		}

		m.certMap[strings.ToLower(domain)] = &cert
		logs.StatusLog("[CERT] Loaded certificate for %s", domain)
	}

	for name := range symLinks {
		if cert, ok := m.certMap[name]; ok {
			m.certMap[name] = cert
		}
	}

	return nil
}

func (m *certMaps) getCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.certMutex.RLock()
	defer m.certMutex.RUnlock()
	host := strings.ToLower(chi.ServerName)
	if cert, ok := m.certMap[host]; ok {
		return cert, nil
	}
	// fallback to any cert if unknown
	for _, c := range m.certMap {
		return c, nil
	}
	logs.StatusLog("[CERT] No certificate for %s", host, chi)
	return nil, nil
}

type LnMultiCerts struct {
	net.Listener
	certMaps *certMaps
	tlsCfg   *tls.Config
	fPort    string
}

func GetTSLListener(secure bool, fPort string) net.Listener {
	ln, err := reuseport.Listen("tcp", fPort)
	if err != nil {
		cmd := exec.Command("lsof", "-i", "-P", "-n")
		b, err1 := cmd.CombinedOutput()
		if err1 == nil {
			b, err1 = services.RunGrep(b, fmt.Sprintf(`':%s.*(LISTEN'`, fPort))
		}
		if err1 != nil {
			logs.ErrorLog(err1, string(b), cmd.String())
		}
		// port is occupied - work serve unpassable
		logs.Fatal(err, "'%s'", b)
	}

	if secure {
		tlsListener := LnMultiCerts{
			certMaps: newCertMaps(),
			fPort:    fPort,
		}
		err := tlsListener.certMaps.loadCertificates()
		if err != nil {
			logs.Fatal(err)
		}
		tlsListener.tlsCfg = &tls.Config{
			GetCertificate: tlsListener.certMaps.getCertificate,
			NextProtos:     []string{"h2", "http/1.1"},
			MinVersion:     tls.VersionTLS12,
		}

		tlsListener.Listener = tls.NewListener(ln, tlsListener.tlsCfg)

		logs.StatusLog("[*] HTTPS multi-site proxy running on %s", fPort)
		return tlsListener
	}
	return ln
}
