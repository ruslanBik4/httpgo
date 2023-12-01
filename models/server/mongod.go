/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Package server назначение модуля - читать и отдавать конфигурационные настройки
package server

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/ruslanBik4/logs"
)

type MongodConfig struct {
	systemPath  string
	wwwPath     string
	SessionPath string
	dbParams    struct {
		DB string `yaml:"dbName"`
	}
}

var mConfig *MongodConfig

// GetMongodConfig функция для получения конфигураций для mongod
func GetMongodConfig() *MongodConfig {

	if mConfig != nil {
		return mConfig
	}

	mConfig = &MongodConfig{}

	return mConfig
}
func (mConfig *MongodConfig) Init(fStatic, fWeb, fSession *string) error {

	mConfig.systemPath = *fStatic
	mConfig.wwwPath = *fWeb
	mConfig.SessionPath = *fSession

	f, err := os.Open(filepath.Join(mConfig.systemPath, "config/mongo.yml"))
	if err != nil {
		return err
	}
	fileInfo, _ := f.Stat()
	b := make([]byte, fileInfo.Size())
	if n, err := f.Read(b); err != nil {
		logs.ErrorLog(err, "n=", n)
		return err

	}

	if err := yaml.Unmarshal(b, &mConfig.dbParams); err != nil {
		return err
	}

	return nil
}

// The Data Source DB has a common format, like e.g. PEAR DB uses it,
// but without type-prefix (optional parts marked by squared brackets):
//
// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func (mConfig *MongodConfig) MongoDBName() string {
	return mConfig.dbParams.DB
}
