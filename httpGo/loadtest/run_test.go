/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package loadtest

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestRunHARHTTP3(t *testing.T) {
	type args struct {
		ctx     context.Context
		harFile string
		cfg     H3LoadConfig
	}
	tests := []struct {
		name    string
		args    args
		want    H3LoadResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunHARHTTP3(tt.args.ctx, tt.args.harFile, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunHARHTTP3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RunHARHTTP3() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_harRequestToHTTP3(t *testing.T) {
	type args struct {
		ctx    context.Context
		target HARRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Request
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := harRequestToHTTP3(tt.args.ctx, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("harRequestToHTTP3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("harRequestToHTTP3() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafariHAROverHTTP3(t *testing.T) {
	result, err := RunHARHTTP3(
		context.Background(),
		"localhost.har", //"ta.lexxinfo.com.har",
		H3LoadConfig{
			Requests:       10_000,
			Concurrency:    50,
			RatePerSecond:  200, // use 0 for maximum possible rate
			RequestTimeout: 30 * time.Second,
			InsecureTLS:    true, // localhost/self-signed only
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf(
		"total=%d success=%d failed=%d mean=%s min=%s max=%s status=%v errors=%v",
		result.Total,
		result.Succeeded,
		result.Failed,
		result.MeanLatency,
		result.MinLatency,
		result.MaxLatency,
		result.StatusCode,
		result.Errors,
	)
}
