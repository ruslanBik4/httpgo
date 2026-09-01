/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package loadtest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
)

type HARFile struct {
	Log struct {
		Entries []HAREntry `json:"entries"`
	} `json:"log"`
}

type HAREntry struct {
	Request HARRequest `json:"request"`
}

type HARRequest struct {
	Method   string      `json:"method"`
	URL      string      `json:"url"`
	Headers  []HARHeader `json:"headers"`
	PostData *HARPost    `json:"postData"`
}

type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARPost struct {
	Text     string `json:"text"`
	Encoding string `json:"encoding"`
}

type H3LoadConfig struct {
	Requests       int           // total requests; targets repeat round-robin
	Concurrency    int           // concurrent H3 workers / connections
	RatePerSecond  int           // 0 = send as fast as possible
	RequestTimeout time.Duration // e.g. 30 * time.Second
	InsecureTLS    bool          // local self-signed certificate only
}

type H3LoadResult struct {
	Total       int
	Succeeded   int
	Failed      int
	StatusCode  map[int]int
	Errors      map[string]int
	MinLatency  time.Duration
	MaxLatency  time.Duration
	MeanLatency time.Duration
}

type h3Attempt struct {
	status  int
	err     error
	latency time.Duration
}

func RunHARHTTP3(
	ctx context.Context,
	harFile string,
	cfg H3LoadConfig,
) (H3LoadResult, error) {
	if cfg.Requests <= 0 {
		return H3LoadResult{}, fmt.Errorf("Requests must be greater than zero")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 30 * time.Second
	}

	data, err := os.ReadFile(harFile)
	if err != nil {
		return H3LoadResult{}, fmt.Errorf("read HAR: %w", err)
	}

	var har HARFile
	if err := json.Unmarshal(data, &har); err != nil {
		return H3LoadResult{}, fmt.Errorf("parse HAR JSON: %w", err)
	}
	if len(har.Log.Entries) == 0 {
		return H3LoadResult{}, fmt.Errorf("HAR contains no requests")
	}

	jobs := make(chan int)
	results := make(chan h3Attempt, cfg.Concurrency)

	var wg sync.WaitGroup
	for worker := 0; worker < cfg.Concurrency; worker++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// One H3 transport per worker gives the test several QUIC connections.
			transport := &http3.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: cfg.InsecureTLS, // #nosec G402 -- local test option
					MinVersion:         tls.VersionTLS13,
				},
			}
			defer transport.Close()

			client := &http.Client{
				Transport: transport,

				// Do not generate unrecorded requests by following redirects.
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			for index := range jobs {
				target := har.Log.Entries[index%len(har.Log.Entries)].Request
				started := time.Now()

				requestCtx, cancel := context.WithTimeout(ctx, cfg.RequestTimeout)
				req, reqErr := harRequestToHTTP3(requestCtx, target)
				if reqErr != nil {
					cancel()
					results <- h3Attempt{err: reqErr}
					continue
				}

				resp, reqErr := client.Do(req)
				cancel()

				attempt := h3Attempt{
					err:     reqErr,
					latency: time.Since(started),
				}

				if resp != nil {
					attempt.status = resp.StatusCode
					_, _ = io.Copy(io.Discard, resp.Body) // fully read response
					_ = resp.Body.Close()
				}

				results <- attempt
			}
		}()
	}

	go func() {
		defer close(jobs)

		var tick <-chan time.Time
		var ticker *time.Ticker

		if cfg.RatePerSecond > 0 {
			ticker = time.NewTicker(time.Second / time.Duration(cfg.RatePerSecond))
			defer ticker.Stop()
			tick = ticker.C
		}

		for i := 0; i < cfg.Requests; i++ {
			if tick != nil {
				select {
				case <-ctx.Done():
					return
				case <-tick:
				}
			}

			select {
			case <-ctx.Done():
				return
			case jobs <- i:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	result := H3LoadResult{
		StatusCode: make(map[int]int),
		Errors:     make(map[string]int),
	}

	var totalLatency time.Duration

	for attempt := range results {
		result.Total++

		if attempt.err != nil {
			result.Failed++
			result.Errors[attempt.err.Error()]++
			continue
		}

		result.StatusCode[attempt.status]++
		if attempt.status >= 200 && attempt.status < 400 {
			result.Succeeded++
		} else {
			result.Failed++
		}

		totalLatency += attempt.latency
		if result.MinLatency == 0 || attempt.latency < result.MinLatency {
			result.MinLatency = attempt.latency
		}
		if attempt.latency > result.MaxLatency {
			result.MaxLatency = attempt.latency
		}
	}

	if result.Total > 0 {
		result.MeanLatency = totalLatency / time.Duration(result.Total)
	}

	return result, nil
}

func harRequestToHTTP3(ctx context.Context, target HARRequest) (*http.Request, error) {
	var body []byte

	if target.PostData != nil {
		body = []byte(target.PostData.Text)

		if target.PostData.Encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(target.PostData.Text)
			if err != nil {
				return nil, fmt.Errorf("decode HAR base64 body: %w", err)
			}
			body = decoded
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		target.Method,
		target.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	for _, h := range target.Headers {
		// These are transport-managed or forbidden in HTTP/3.
		switch strings.ToLower(h.Name) {
		case "host", "content-length", "connection", "keep-alive",
			"proxy-connection", "transfer-encoding", "upgrade":
			continue
		}

		req.Header.Add(h.Name, h.Value)
	}

	return req, nil
}
