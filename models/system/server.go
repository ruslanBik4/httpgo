// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package system

type serverConfig struct {
	StaticPath string
	WWWPath string
	SessionPath string
}

var ServerConfig serverConfig

func (sConfig *serverConfig) Init(f_static, f_web, f_session *string) {
	sConfig.StaticPath  = *f_static
	sConfig.WWWPath     = *f_web
	sConfig.SessionPath = *f_session
}
