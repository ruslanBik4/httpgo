// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server назначение модуля - читать и отдавать конфигурационные настройки
package server

import (
	"fmt"
	"github.com/ruslanBik4/httpgo/models/logs"
	yaml "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"time"
)

type serverConfig struct {
	systemPath  string
	wwwPath     string
	SessionPath string
	dbParams    struct {
		DB   string `yaml:"dbName"`
		User string `yaml:"dbUser"`
		Pass string `yaml:"dbPass"`
		Prot string `yaml:"dbProt"`
		ConfigParams map[string]string `yaml:"configParams"`
	}
	StartTime time.Time
}

var sConfig *serverConfig
// GetServerConfig return reference on server config structure
func GetServerConfig() *serverConfig {

	if sConfig != nil {
		return sConfig
	}

	sConfig = &serverConfig{}
	return sConfig
}
func (sConfig *serverConfig) Init(fStatic, fWeb, fSession *string) error {
	sConfig.systemPath = *fStatic
	sConfig.wwwPath = *fWeb
	sConfig.SessionPath = *fSession
	sConfig.StartTime = time.Now()

	f, err := os.Open(filepath.Join(sConfig.systemPath, "config/db.yml"))
	if err != nil {
		return err
	}
	fileInfo, _ := f.Stat()
	b := make([]byte, fileInfo.Size())
	if n, err := f.Read(b); err != nil {

		logs.ErrorLog(err, "n=", n)
		return err

	}

	if err := yaml.Unmarshal(b, &sConfig.dbParams); err != nil {
		logs.ErrorLog(err, b, &sConfig.dbParams)
		return err
	}

	return nil
}

//The Data Source DB has a common format, like e.g. PEAR DB uses it,
// but without type-prefix (optional parts marked by squared brackets):
//
//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func (sConfig *serverConfig) DNSConnection() string {

	addConfig := sConfig.generateConfigString()
	return fmt.Sprintf("%s:%s@%s/%s?maximumpoolsize&%s", sConfig.dbParams.User, sConfig.dbParams.Pass, sConfig.dbParams.Prot,
		sConfig.dbParams.DB, addConfig)
}
func (sConfig *serverConfig) DBName() string {
	return sConfig.dbParams.DB
}
func (sConfig *serverConfig) SystemPath() string {
	return sConfig.systemPath
}
func (sConfig *serverConfig) WWWPath() string {
	return sConfig.wwwPath
}
func (sConfig *serverConfig) generateConfigString() (result string) {

	var separator string
	for key, val := range sConfig.dbParams.ConfigParams {
		result += separator + key + "=" + val
		separator = "&"
	}

	return result
}