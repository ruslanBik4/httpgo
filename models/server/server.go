// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// назначение модуля - читать и одавать конфигурационные настройки

package server

import (
	"fmt"
 	yaml "gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
)

type serverConfig struct {
	StaticPath string
	WWWPath string
	SessionPath string
	dbParams struct {
		DB   string `yaml:"dbName"`
		User string `yaml:"dbUser"`
		Pass string `yaml:"dbPass"`
	}
}

var sConfig *serverConfig

func GetServerConfig() *serverConfig {

	if sConfig != nil {
		return sConfig
	} else {
		sConfig = &serverConfig{}
	}

	return sConfig
}
func (sConfig *serverConfig) Init(f_static, f_web, f_session *string) error{
	sConfig.StaticPath  = *f_static
	sConfig.WWWPath     = *f_web
	sConfig.SessionPath = *f_session

	f, err := os.Open(filepath.Join(sConfig.StaticPath, "config/db.yml" ))
	if err != nil {
		return err
	}
	fileInfo, _ := f.Stat()
	b  := make([]byte, fileInfo.Size())
	if n, err := f.Read(b); err != nil {
		log.Println(n)
		return err

	}

	if err := yaml.Unmarshal(b, &sConfig.dbParams); err != nil {
		return err
	}


	return nil
}
func  writeto(v interface{}) error {
  log.Println(v)
	return nil
}
//The Data Source DB has a common format, like e.g. PEAR DB uses it,
// but without type-prefix (optional parts marked by squared brackets):
//
//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func (sConfig *serverConfig) DNSConnection() string {
	return fmt.Sprintf("%s:%s@/%s?persistent", sConfig.dbParams.User, sConfig.dbParams.Pass, sConfig.dbParams.DB)
}
func (sConfig *serverConfig) DBName() string {
	return sConfig.dbParams.DB
}
