// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// хранит кешированные данные БД (пока только справочники)
// для использования при отдаче данных
package cache

// хранит структуру полей - стоит продумать, как хранить еще и ключи
var ListsCache map[string]map[string]string

func GetListRecord(tableName string) map[string]string {
	if list, ok := ListsCache[tableName]; ok {
		return list
	}

	return nil
}

func init() {
	ListsCache = make(map[string]map[string]string, 0)
}
