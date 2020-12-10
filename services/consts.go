// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

// STATUS_PREPARING - стстус Сервиса - "Подготовка сервиса"
const STATUS_PREPARING = "preparing data"

// STATUS_ERROR - стстус Сервиса - "Ошибка"
const STATUS_ERROR = "error"

// STATUS_READY - стстус Сервиса - "Готово"
const STATUS_READY = "ready"

// param of log showing
const (
	paramsSystemctlUnit = "unit"
	paramDate           = "date"
	paramTime           = "time"
	paramAgo            = "ago"
	paramPatter         = "pattern"
)
const (
	ShowStatus         = "/api/status/"
	ShowDBStatus       = "/api/status/db"
	ShowPsqlLog        = "/api/status/psql"
	ShowStatusServices = "/api/status/services"
	ShowDebugLog       = "/api/log/debug/"
	ShowErrorsLog      = "/api/log/errors/"
	ShowFEUpdateLog    = "/api/log/errors/front/update/"
	ShowInfoLog        = "/api/log/info/"
)
