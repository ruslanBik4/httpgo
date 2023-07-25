/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package auth

import (
	"time"
)

const UserValueToken = "UserToken"
const tokenExpires = 60 * 60 * 24 * time.Second
const authManager = "auth"
