/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package forms

type Button struct {
	Title    string
	Position bool
	Type     string
	OnClick  string
}

type BlockColumns struct {
	Id                 int
	Buttons            []Button
	Columns            []*ColumnDecor
	Multiple           bool
	Title, Description string
}

type FormField struct {
	Title, Action, Method, Description string
	HideBlock                          any
	Blocks                             []BlockColumns
}
