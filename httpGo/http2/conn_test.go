/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package http2

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	xhttp2 "golang.org/x/net/http2"
)

func selfSigned(t testing.TB) (certPEM, keyPEM []byte) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	kb, _ := x509.MarshalECPrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb})
}

// startServer runs fasthttp with our HTTP/2 NextProto — exactly like httpGo.setHTTP2.
func startServer(t testing.TB, h fasthttp.RequestHandler, cfg *Config) (addr string, stop func()) {
	cert, key := selfSigned(t)
	srv := &fasthttp.Server{Handler: h}
	h2 := NewServer(h, cfg)
	srv.NextProto(H2, h2.ServeConn)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		err := srv.ServeTLSEmbed(ln, cert, key)
		if err != nil {
			t.Fatal(err)
		}
	}() //nolint:errcheck
	return ln.Addr().String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		go func() { _ = h2.Shutdown(ctx) }()
		if err := srv.ShutdownWithContext(ctx); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	}
}

const H2 = "h2"

func h2Client() *http.Client {
	return &http.Client{Transport: &xhttp2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"h2"}},
	}, Timeout: 30 * time.Second}
}

func echoHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/echo":
		ctx.Response.Header.Set("X-Method", string(ctx.Method()))
		ctx.Response.Header.Set("X-Proto", string(ctx.Request.Header.Protocol()))
		ctx.Response.Header.Set("X-Host", string(ctx.Host()))
		ctx.Response.Header.Set("X-Cookie-A", string(ctx.Request.Header.Cookie("a")))
		ctx.Response.Header.Set("X-Cookie-B", string(ctx.Request.Header.Cookie("b")))
		ctx.Response.Header.Set("X-Query", string(ctx.QueryArgs().Peek("q")))
		ctx.Response.Header.Set("X-TLS", fmt.Sprint(ctx.IsTLS()))
		ctx.Response.Header.Set("X-Big-Len", fmt.Sprint(len(ctx.Request.Header.Peek("X-Big"))))
		ctx.SetContentType("application/octet-stream")
		ctx.SetBody(ctx.PostBody())
	case "/big":
		ctx.SetBody(bytes.Repeat([]byte("0123456789abcdef"), 20<<20/16)) // 20 MiB
	case "/stream":
		ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
			for i := range 5 {
				fmt.Fprintf(w, "chunk-%d\n", i)
				_ = w.Flush()
				time.Sleep(10 * time.Millisecond)
			}
		})
	case "/slow":
		time.Sleep(200 * time.Millisecond)
		ctx.SetBodyString("slow")
	case "/panic":
		panic("boom")
	case "/nocontent":
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func TestBasic(t *testing.T) {
	addr, stop := startServer(t, echoHandler, nil)
	defer stop()
	cl := h2Client()

	req, _ := http.NewRequest("POST", "https://"+addr+"/echo?q=42", strings.NewReader("hello h2"))
	req.AddCookie(&http.Cookie{Name: "a", Value: "1"})
	req.AddCookie(&http.Cookie{Name: "b", Value: "2"})
	req.Header.Set("X-Big", strings.Repeat("x", 40<<10)) // forces CONTINUATION frames
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	check := map[string]string{
		"X-Method": "POST", "X-Proto": "HTTP/2.0", "X-Cookie-A": "1", "X-Cookie-B": "2",
		"X-Query": "42", "X-TLS": "true", "X-Big-Len": "40960", "X-Host": addr,
	}
	if resp.ProtoMajor != 2 || string(body) != "hello h2" {
		t.Fatalf("proto=%s body=%q", resp.Proto, body)
	}
	for k, v := range check {
		if got := resp.Header.Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if resp.Header.Get("Date") == "" || resp.ContentLength != 8 {
		t.Errorf("date=%q cl=%d", resp.Header.Get("Date"), resp.ContentLength)
	}

	resp, err = cl.Get("https://" + addr + "/nocontent")
	if err != nil || resp.StatusCode != 204 {
		t.Fatalf("204: %v %v", err, resp)
	}
	resp.Body.Close()
}

func TestLargeBodies(t *testing.T) {
	addr, stop := startServer(t, echoHandler, &Config{MaxRequestBodySize: 16 << 20})
	defer stop()
	cl := h2Client()

	// upload 8 MiB > our 1 MiB stream window: exercises receive-side WINDOW_UPDATEs
	up := bytes.Repeat([]byte("u"), 8<<20)
	resp, err := cl.Post("https://"+addr+"/echo", "application/octet-stream", bytes.NewReader(up))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Equal(got, up) {
		t.Fatalf("echo mismatch: %d bytes", len(got))
	}

	// download 20 MiB > client window: exercises send-side flow control blocking
	resp, err = cl.Get("https://" + addr + "/big")
	if err != nil {
		t.Fatal(err)
	}
	n, _ := io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if n != 20<<20 {
		t.Fatalf("big: got %d", n)
	}
}

func TestTooLarge(t *testing.T) {
	addr, stop := startServer(t, echoHandler, &Config{MaxRequestBodySize: 1 << 10})
	defer stop()
	resp, err := h2Client().Post("https://"+addr+"/echo", "text/plain", bytes.NewReader(make([]byte, 4<<10)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 413 {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestConcurrentStreams(t *testing.T) {
	addr, stop := startServer(t, echoHandler, nil)
	defer stop()
	cl := h2Client()
	var wg sync.WaitGroup
	start := time.Now()
	errs := make(chan error, 200)
	for i := range 200 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := "/echo"
			if i%10 == 0 {
				path = "/slow"
			}
			msg := fmt.Sprintf("msg-%d", i)
			resp, err := cl.Post("https://"+addr+path, "text/plain", strings.NewReader(msg))
			if err != nil {
				errs <- err
				return
			}
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if path == "/echo" && string(b) != msg {
				errs <- fmt.Errorf("got %q want %q", b, msg)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	// 20 slow (200ms) requests multiplexed on one conn must run in parallel.
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("not multiplexed: took %v", d)
	}
}

func TestStreamingAndPanicAndCancel(t *testing.T) {
	addr, stop := startServer(t, echoHandler, nil)
	defer stop()
	cl := h2Client()

	resp, err := cl.Get("https://" + addr + "/stream")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if strings.Count(string(b), "chunk-") != 5 {
		t.Fatalf("stream body %q", b)
	}

	if _, err := cl.Get("https://" + addr + "/panic"); err == nil {
		t.Fatal("expected stream reset on panic")
	}

	// client cancels a download mid-way (RST_STREAM from client)
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://"+addr+"/big", nil)
	resp, err = cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.CopyN(io.Discard, resp.Body, 1<<20)
	cancel()
	resp.Body.Close()

	// connection still healthy
	resp, err = cl.Post("https://"+addr+"/echo", "text/plain", strings.NewReader("after"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(b) != "after" {
		t.Fatalf("after cancel: %q", b)
	}
}

func BenchmarkGET(b *testing.B) {
	addr, stop := startServer(b, func(ctx *fasthttp.RequestCtx) { ctx.SetBodyString("ok") }, nil)
	defer stop()
	cl := h2Client()
	url := "https://" + addr + "/"
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := cl.Get(url)
			if err != nil {
				b.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
