/*
 * Copyright (c) 2024-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"fmt"
	"net"

	"github.com/ruslanBik4/logs"
)

type blockListener struct {
	net.Listener
	*AccessConf
}

func (b *blockListener) Addr() net.Addr {
	return b.Listener.Addr()
}

func (b *blockListener) Accept() (net.Conn, error) {
	for {
		c, err := b.Listener.Accept()
		if err != nil {
			return nil, err
		}

		addr := c.RemoteAddr().String()
		if b.isAllowIP(addr) || (len(b.AllowIP) == 0 && !b.isDenyIP(addr)) || len(b.AllowRoute) > 0 {
			return c, nil
		}

		//for path, msg := range b.QuickResponse {
		//
		//}
		logs.ErrorLog(fmt.Errorf("deny connect from addr %s", addr))
		err = c.Close()
		if err != nil {
			logs.ErrorLog(err, "close connection")
		}
	}
}
