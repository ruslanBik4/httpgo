// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// назначение модуля - читать и одавать конфигурационные настройки

package server

import (
	"fmt"
	"github.com/ruslanBik4/httpgo/models/logs"
	yaml "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type serverConfig struct {
	systemPath  string
	wwwPath     string
	SessionPath string
	dbParams    struct {
		DB   string `yaml:"dbName"`
		User string `yaml:"dbUser"`
		Pass string `yaml:"dbPass"`
		AllowAllFiles string `yaml:"allowAllFiles"`
		AllowCleartextPasswords string `yaml:"allowCleartextPasswords"`
		AllowNativePasswords string `yaml:"allowNativePasswords"`
		AllowOldPasswords string `yaml:"allowOldPasswords"`
		Charset string `yaml:"charset"`
		Collation string `yaml:"collation"`
		ClientFoundRows string `yaml:"clientFoundRows"`
		ColumnsWithAlias string `yaml:"columnsWithAlias"`
		InterpolateParams string `yaml:"interpolateParams"`
		Loc string `yaml:"loc"`
		MaxAllowedPacket string `yaml:"maxAllowedPacket"`
		MultiStatements string `yaml:"multiStatements"`
		ParseTime string `yaml:"parseTime"`
		ReadTimeout string `yaml:"readTimeout"`
		RejectReadOnly string `yaml:"rejectReadOnly"`
		Strict string `yaml:"strict"`
		Timeout string `yaml:"timeout"`
		Tls string `yaml:"tls"`
		WriteTimeout string `yaml:"writeTimeout"`
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
func (sConfig *serverConfig) Init(f_static, f_web, f_session *string) error {
	sConfig.systemPath = *f_static
	sConfig.wwwPath = *f_web
	sConfig.SessionPath = *f_session

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

	return fmt.Sprintf("%s:%s@/%s?allowAllFiles=%s&allowCleartextPasswords=%s&allowNativePasswords=%s&" +
		"allowOldPasswords=%s&charset=%s&collation=%s&clientFoundRows=%s&columnsWithAlias=%s&" +
		"interpolateParams=%s&loc=%s&maxAllowedPacket=%s&multiStatements=%s&parseTime=%s&" +
		"readTimeout=%s&rejectReadOnly=%s&strict=%s&timeout=%s&tls=%s&writeTimeout=%s",
		sConfig.dbParams.User, sConfig.dbParams.Pass, sConfig.dbParams.DB,
		sConfig.dbParams.AllowAllFiles, sConfig.dbParams.AllowCleartextPasswords, sConfig.dbParams.AllowNativePasswords,
		sConfig.dbParams.AllowOldPasswords,
		sConfig.dbParams.Charset, sConfig.dbParams.Collation, sConfig.dbParams.ClientFoundRows,
		sConfig.dbParams.ColumnsWithAlias, sConfig.dbParams.InterpolateParams,
		sConfig.dbParams.Loc, sConfig.dbParams.MaxAllowedPacket, sConfig.dbParams.MultiStatements,
		sConfig.dbParams.ParseTime,sConfig.dbParams.ReadTimeout,
		sConfig.dbParams.RejectReadOnly, sConfig.dbParams.Strict, sConfig.dbParams.Timeout, sConfig.dbParams.Tls,
		sConfig.dbParams.WriteTimeout)
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
