// Copyright 2021 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

func (m *blockListener) Addr() net.Addr {
	return m.Listener.Addr()
}

func (m *blockListener) Accept() (net.Conn, error) {
	for {
		c, err := m.Listener.Accept()
		if err != nil {
			return nil, err
		}

		addr := c.RemoteAddr().String()
		if m.isAllowIP(addr) || (len(m.AllowIP) == 0 && !m.isDenyIP(addr)) || len(m.AllowRoute) > 0 {
			return c, nil
		}

		logs.ErrorLog(fmt.Errorf("deny connect from addr %s", addr))
		err = c.Close()
		if err != nil {
			logs.ErrorLog(err, "close connection")
		}
	}
}
