/*
 * Copyright (c) 2023-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import "fmt"

const ShowVersion = "/api/version()"

// version
var (
	Version   string
	HTTPGOVer string
	GoVersion string
	OSVersion string
	Build     string
	Branch    string
)

func GetAppTitle(name string) string {
	return fmt.Sprintf("%s (%s) Version: %s, Build Time: %s", name, Branch, Version, Build)
}
