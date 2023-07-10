/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestAccessConf_Allow(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		ctx  *fasthttp.RequestCtx
		addr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.Allow(tt.args.ctx, tt.args.addr); got != tt.want {
				t.Errorf("Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_Deny(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		ctx  *fasthttp.RequestCtx
		addr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.Deny(tt.args.ctx, tt.args.addr); got != tt.want {
				t.Errorf("Deny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_IsAccessConf(t *testing.T) {
	tests := []struct {
		name       string
		AccessConf *AccessConf
		want       bool
	}{
		// TODO: Add test cases.
		{
			"nil",
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.AccessConf
			if got := cfg.IsAccess(); got != tt.want {
				t.Errorf("IsAccessConf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_Reload(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	tests := []struct {
		name    string
		fields  fields
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			got, err := cfg.Reload()
			if (err != nil) != tt.wantErr {
				t.Errorf("Reload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reload() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_isAllowIP(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		addr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.isAllowIP(tt.args.addr); got != tt.want {
				t.Errorf("isAllowIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_isAllowRoute(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.isAllowRoute(tt.args.ctx); got != tt.want {
				t.Errorf("isAllowRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_isDenyIP(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		addr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.isDenyIP(tt.args.addr); got != tt.want {
				t.Errorf("isDenyIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessConf_isDenyRoute(t *testing.T) {
	type fields struct {
		fileCfg    string
		Domains    map[string]string
		KillSignal int
		Server     *fasthttp.Server
		AccessConf *AccessConf
	}
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CfgHttp{
				fileCfg:    tt.fields.fileCfg,
				Domains:    tt.fields.Domains,
				KillSignal: tt.fields.KillSignal,
				Server:     tt.fields.Server,
				AccessConf: tt.fields.AccessConf,
			}
			if got := cfg.isDenyRoute(tt.args.ctx); got != tt.want {
				t.Errorf("isDenyRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCfgHttp(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"1",
			"../config/httpgo.yml.sample",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfgGlobal, err := NewCfgHttp(tt.filename)
			if !assert.Equal(t, tt.wantErr, err != nil) {
				t.Error(err)
				return
			}

			t.Log(gotCfgGlobal)
			assert.Implements(t, (*Access)(nil), gotCfgGlobal)
			s, err := jsoniter.MarshalToString(gotCfgGlobal)
			t.Log(s, err)

		})
	}
}
