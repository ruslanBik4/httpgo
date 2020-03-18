// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpGo

import (
	"fmt"
	"net"
	"strings"
	
	"github.com/ruslanBik4/httpgo/logs"
)

type blockListener struct {
	net.Listener
	*CfgHttp
}

func (m *blockListener) Addr() net.Addr {
	return m.Addr()
}

func (m *blockListener) Accept() (net.Conn, error) {

	for {
		c, err := m.Listener.Accept()
		if err != nil {
			return nil, err
		}
		
		addr := c.RemoteAddr().String()
		//todo: add chk deny later
		if m.isAllowIP(addr) {
			return c, nil
		}
		
		logs.ErrorLog(fmt.Errorf("Deny connect from addr %s", addr))
		c.Close()

	}
	
	
	return nil, nil
}

func (m *blockListener) Close() error {
	return m.Close()
}
