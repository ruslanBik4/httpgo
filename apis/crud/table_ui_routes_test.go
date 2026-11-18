/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ruslanBik4/httpgo/apis"
)

func TestCreateRoutesForTableSwitches(t *testing.T) {
	users, roles := newUsers()
	db := newDB(users, roles)
	routes := createRoutesForTable(db, "users", "/table/", users,
		[]DbRouteType{DbRouteType_Update, DbRouteType_Select}, []apis.InParam{ParamsLang})

	assert.NotContains(t, routes, "/table/users/set_bool", "the visible rows go through put/batch of the project")

	all := routes["/table/users/set_bool_all"]
	require.NotNil(t, all)
	assert.True(t, all.NeedAuth)
	assert.True(t, all.OnlyAdmin, "the whole table is changed by an administrator only")
	assert.Equal(t, apis.POST, all.Method)
	names := make([]string, len(all.Params))
	for i, p := range all.Params {
		names[i] = p.Name
	}
	assert.Equal(t, []string{"column", "value"}, names)

	assert.Contains(t, routes, "/table/users/card")

	t.Run("no switches without an update route or a boolean column", func(t *testing.T) {
		routes := createRoutesForTable(db, "users", "/table/", users, []DbRouteType{DbRouteType_Select}, nil)
		assert.NotContains(t, routes, "/table/users/set_bool_all")
		routes = createRoutesForTable(db, "roles", "/table/", roles, []DbRouteType{DbRouteType_Update}, nil)
		assert.NotContains(t, routes, "/table/roles/set_bool_all")
	})
}
