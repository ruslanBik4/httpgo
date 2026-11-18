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

	setBool := routes["/table/users/set_bool"]
	require.NotNil(t, setBool)
	assert.True(t, setBool.NeedAuth)
	assert.False(t, setBool.OnlyAdmin, "the visible rows may be changed by any editor")
	assert.Equal(t, apis.POST, setBool.Method)
	names := make([]string, len(setBool.Params))
	for i, p := range setBool.Params {
		names[i] = p.Name
	}
	assert.Equal(t, []string{"column", "value", "ids"}, names)

	all := routes["/table/users/set_bool_all"]
	require.NotNil(t, all)
	assert.True(t, all.OnlyAdmin, "the whole table is changed by an administrator only")
	assert.Len(t, all.Params, 2, "no ids")

	assert.Contains(t, routes, "/table/users/card")

	t.Run("no switches without an update route or a boolean column", func(t *testing.T) {
		routes := createRoutesForTable(db, "users", "/table/", users, []DbRouteType{DbRouteType_Select}, nil)
		assert.NotContains(t, routes, "/table/users/set_bool")
		routes = createRoutesForTable(db, "roles", "/table/", roles, []DbRouteType{DbRouteType_Update}, nil)
		assert.NotContains(t, routes, "/table/roles/set_bool")
	})
}
