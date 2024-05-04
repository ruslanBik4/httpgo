/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package layouts

import "github.com/pkg/errors"

var ErrSubMennuNotFound = errors.New("item has't subMenu")

type Menu []*ItemMenu

func (m Menu) Add(item *ItemMenu) Menu {
	return append(m, item)
}

func (m Menu) FindItem(name string) (*ItemMenu, int) {
	for i, item := range m {
		if item.Name == name {
			return item, i
		}
	}

	return nil, -1
}

type ItemMenu struct {
	Content, Link, Label, Name, OnClick, Title, Target string
	SubMenu                                            Menu
	Attr                                               map[string]string
}

func (m *ItemMenu) AddSubItem(item *ItemMenu) error {
	if m.SubMenu == nil {
		return ErrSubMennuNotFound
	}

	m.SubMenu = m.SubMenu.Add(item)
	return nil
}
