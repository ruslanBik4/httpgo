/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package layouts

import (
	"github.com/pkg/errors"
	"slices"
)

var ErrSubMennuNotFound = errors.New("item has't subMenu")

type Menu []*ItemMenu

func (m *Menu) Add(item *ItemMenu) *Menu {
	*m = append(*m, item)

	return m
}

func (m Menu) FindItem(name string) (*ItemMenu, int) {
	i := slices.IndexFunc(m, func(item *ItemMenu) bool {
		return item.Name == name
	})
	if i < 0 {
		return nil, -1
	}

	return m[i], i
}

type ItemMenu struct {
	Content, Class, Link, Label, Name, OnClick, Title, Target string
	SubMenu                                                   Menu
	Attr                                                      map[string]string
}

func (m *ItemMenu) AddSubItem(item *ItemMenu) error {
	if m.SubMenu == nil {
		return ErrSubMennuNotFound
	}

	m.SubMenu.Add(item)

	return nil
}
