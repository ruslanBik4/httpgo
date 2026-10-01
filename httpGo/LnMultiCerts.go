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
	"slices"
	"strings"
	"sync"

	"github.com/valyala/fasthttp/reuseport"

	"github.com/ruslanBik4/httpgo/services"
	"github.com/ruslanBik4/logs"
)

// DefaultCertDir is a directory containing a subfolder per domain
// (<dir>/<domain>/fullchain.pem + privkey.pem, or server.crt + server.key).
const DefaultCertDir = "/etc/ssl/sites"

// CertDirEnv overrides DefaultCertDir when no directory is passed explicitly.
// Several directories are separated like $PATH: "/etc/ssl/sites:/home/app/certs".
const CertDirEnv = "HTTPGO_CERT_DIR"

// ResolveCertDirs picks the certificate directories:
// explicit arguments > $HTTPGO_CERT_DIR > DefaultCertDir.
// Empty entries are ignored, duplicates removed, order kept (it is the priority).
func ResolveCertDirs(dirs ...string) []string {
	res := cleanDirs(dirs)
	if len(res) == 0 {
		res = cleanDirs(filepath.SplitList(os.Getenv(CertDirEnv)))
	}
	if len(res) == 0 {
		res = []string{DefaultCertDir}
	}
	return res
}

func cleanDirs(dirs []string) []string {
	res := make([]string, 0, len(dirs))
	for _, d := range dirs {
		if d = strings.TrimSpace(d); d == "" {
			continue
		}
		if d = filepath.Clean(d); !slices.Contains(res, d) {
			res = append(res, d)
		}
	}
	return res
}

// certificate cache
type certMaps struct {
	baseDirs  []string
	certMap   map[string]*tls.Certificate
	certMutex sync.RWMutex
}

// newCertMaps creates a cache reading from baseDirs (none/empty -> ResolveCertDirs default).
func newCertMaps(baseDirs ...string) *certMaps {
	return &certMaps{
		baseDirs: ResolveCertDirs(baseDirs...),
		certMap:  make(map[string]*tls.Certificate),
	}
}

// firstExisting returns the first file of names that exists in dir.
func firstExisting(dir string, names ...string) (string, error) {
	var err error
	for _, name := range names {
		p := filepath.Join(dir, name)
		if _, err = os.Stat(p); err == nil {
			return p, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("none of %v found in %s: %w", names, dir, err)
}

func (m *certMaps) readCertificates(baseDir, domain string) (tls.Certificate, error) {
	dir := filepath.Join(baseDir, domain)

	certFile, err := firstExisting(dir, "fullchain.pem", "server.crt")
	if err != nil {
		return tls.Certificate{}, err
	}
	keyFile, err := firstExisting(dir, "privkey.pem", "server.key")
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(certFile, keyFile)
}

// loadCertificates scans every base directory in order. When a domain exists
// in several directories, the first one wins. A missing/unreadable directory
// is logged and skipped; it's an error only if no certificate loads at all.
func (m *certMaps) loadCertificates() error {
	m.certMutex.Lock()
	defer m.certMutex.Unlock()

	// alias domain -> target domain, e.g. www.example.com -> example.com.
	// Resolved after all directories are read, so an alias may point to a
	// domain found in another directory.
	symLinks := make(map[string]string)
	var dirErrs []error
	for _, baseDir := range m.baseDirs {
		if err := m.loadDir(baseDir, symLinks); err != nil {
			logs.ErrorLog(err)
			dirErrs = append(dirErrs, err)
		}
	}

	for alias, target := range symLinks {
		if _, ok := m.certMap[alias]; ok {
			continue // a real cert directory with that name wins over an alias
		}
		if cert, ok := m.certMap[target]; ok {
			m.certMap[alias] = cert
			logs.StatusLog("[CERT] %s uses certificate of %s", alias, target)
		}
	}

	if len(m.certMap) == 0 {
		return fmt.Errorf("no certificates loaded from %v (set %s or pass certDirs): %w",
			m.baseDirs, CertDirEnv, errors.Join(dirErrs...))
	}

	return nil
}

func (m *certMaps) loadDir(baseDir string, symLinks map[string]string) error {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("cannot read cert directory %s: %w", baseDir, err)
	}

	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if !e.IsDir() {
			if e.Type()&fs.ModeSymlink != 0 {
				// Resolve symlink (one hop)
				targetPath, err := os.Readlink(filepath.Join(baseDir, e.Name()))
				if err != nil {
					logs.ErrorLog(err)
					continue
				}
				if _, ok := symLinks[name]; !ok { // first directory wins
					symLinks[name] = strings.ToLower(filepath.Base(targetPath))
				}
			}
			continue
		}
		if _, ok := m.certMap[name]; ok {
			logs.DebugLog("[CERT] %s in %s skipped: already loaded from an earlier directory", e.Name(), baseDir)
			continue
		}
		cert, err := m.readCertificates(baseDir, e.Name())
		if err != nil {
			logs.ErrorLog(err, "Load cert for %s failed", e.Name())
			continue
		}

		m.certMap[name] = &cert
		logs.StatusLog("[CERT] Loaded certificate for %s from %s", e.Name(), baseDir)
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

// CertDirs reports the directories certificates are loaded from, in priority order.
func (ln LnMultiCerts) CertDirs() []string { return slices.Clone(ln.certMaps.baseDirs) }

// GetTSLListener listens on fPort; when secure, serves TLS with per-domain
// certificates searched in certDirs, in order (optional; default:
// $HTTPGO_CERT_DIR or /etc/ssl/sites).
func GetTSLListener(secure bool, fPort string, certDirs ...string) net.Listener {
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
			certMaps: newCertMaps(certDirs...),
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

		logs.StatusLog("[*] HTTPS multi-site proxy running on %s (certs: %v)", fPort, tlsListener.CertDirs())
		return tlsListener
	}
	return ln
}
