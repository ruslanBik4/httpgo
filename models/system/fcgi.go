// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// prepare & run request into php-fpm

package system

import (
	"bytes"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"bitbucket.org/PinIdea/fcgi_client"
	"github.com/pkg/errors"
	. "github.com/valyala/fasthttp"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/models/logs"
)

const internalRewriteFieldName = "travel"

var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
)

// FCGI structure for request in PHP-FPM
type FCGI struct {
	Sock string
	Env  func(ctx *RequestCtx) map[string]string
}

func (c *FCGI) defaultEnv(ctx *RequestCtx) map[string]string {
	return map[string]string{
		"REQUEST_METHOD":  string(ctx.Method()),
		"SCRIPT_FILENAME": string(ctx.URI().Path()),
		"SCRIPT_NAME":     string(ctx.URI().Path()),
		"QUERY_STRING":    string(ctx.URI().QueryString()),
	}
}

// Do run request
func (c *FCGI) Do(ctx *RequestCtx) (*http.Response, error) {
	const typeSckt = "unix" // or "unixgram" or "unixpacket"

	fcgi, err := fcgiclient.Dial(typeSckt, c.Sock)
	if err != nil {
		return nil, err
	}
	env := c.Env
	if env == nil {
		env = c.defaultEnv
	}
	params := env(ctx)

	switch string(ctx.Method()) {
	case "GET":
		return fcgi.Get(params)
	case "POST":
		b := ctx.PostBody()
		return fcgi.Post(params, params["CONTENT_TYPE"], bytes.NewReader(b), len(b))
	}

	return nil, apis.ErrWrongParamsList
}

// ServeHTTP get request response & render to output
func (c *FCGI) ServeHTTP(ctx *RequestCtx) (interface{}, error) {
	resp, err := c.Do(ctx)
	if err != nil {
		return nil, err
	}

	status, isStatus := resp.Header["Status"]
	location, isURL := resp.Header["Location"]
	if isStatus && (status[0] == "302 Found") && isURL {
		ctx.Redirect(location[0], StatusTemporaryRedirect)
		return nil, nil
	}

	for key, val := range resp.Header {
		ctx.Response.Header.Set(key, strings.Join(val, ";"))
	}
	ctx.Response.SetBodyStream(resp.Body, int(resp.ContentLength))

	return nil, nil
}

// NewFPM create new FCGI
func NewFPM(sock string) *FCGI {
	return &FCGI{Sock: sock}
}

// NewPHP create new FCGI for PHP scripts
func NewPHP(root string, priScript, sock string) *FCGI {
	return &FCGI{
		Sock: sock,
		Env: func(ctx *RequestCtx) map[string]string {

			ip, port := ctx.RemoteAddr().String(), ""
			if idx := strings.LastIndex(ip, ":"); idx > -1 {
				port = ip[idx+1:]
				ip = ip[:idx]
			}
			pathInfo, docURI := "", string(ctx.RequestURI())

			if idx := strings.Index(docURI, pathInfo); idx > -1 {
				docURI = docURI[len(pathInfo):]
			}
			// Some variables are unused but cleared explicitly to prevent
			// the parent environment from interfering.
			env := map[string]string{

				// Variables defined in CGI 1.1 spec
				"AUTH_TYPE":         "", // Not used
				"CONTENT_LENGTH":    strconv.Itoa(ctx.Request.Header.ContentLength()),
				"CONTENT_TYPE":      string(ctx.Request.Header.ContentType()),
				"GATEWAY_INTERFACE": "CGI/1.1",
				"PATH_INFO":         pathInfo,
				"QUERY_STRING":      string(ctx.URI().QueryString()),
				"REMOTE_ADDR":       ip,
				"REMOTE_HOST":       ip, // For speed, remote host lookups disabled
				"REMOTE_PORT":       port,
				"REMOTE_IDENT":      "", // Not used
				"REMOTE_USER":       "", // Not used
				"REQUEST_METHOD":    string(ctx.Method()),
				"SERVER_NAME":       string(ctx.Host()),
				"SERVER_PORT":       ":80", //TODO
				"SERVER_PROTOCOL":   "HTTP 1.1",
				"SERVER_SOFTWARE":   "httpGo 0.01",

				// Other variables
				"DOCUMENT_ROOT":   root,
				"DOCUMENT_URI":    docURI,
				"HTTP_HOST":       string(ctx.Host()), // added here, since not always part of headers
				"REQUEST_URI":     ctx.URI().String(),
				"SCRIPT_FILENAME": priScript,
				"SCRIPT_NAME":     priScript,
			}
			// compliance with the CGI specification that PATH_TRANSLATED
			// should only exist if PATH_INFO is defined.
			// Info: https://www.ietf.org/rfc/rfc3875 Page 14
			//if env["PATH_INFO"] != "" {
			//	env["PATH_TRANSLATED"] = filepath.Join(pathToYii, pathInfo) // Info: http://www.oreilly.com/openbook/cgi/ch02_04.html
			//}

			// Some web apps rely on knowing HTTPS or not
			if ctx.IsTLS() {
				env["HTTPS"] = "on"
				env["test"] = path.Base(root)
			}

			// Add all HTTP headers (except Caddy-Rewrite-Original-URI ) to env variables
			ctx.Request.Header.VisitAll(func(key, value []byte) {
				field, val := string(key), string(value)
				// /observe
				if strings.ToLower(field) == strings.ToLower(internalRewriteFieldName) {
					return
				}
				header := strings.ToUpper(field)
				header = headerNameReplacer.Replace(header)
				env["HTTP_"+header] = val
			})

			return env
		},
	}
}

// WriteError не уверен, что это должно быть здесь - должен быть какой общий механизм для выдачи такого
func WriteError(ctx *RequestCtx, err error) bool {
	if err == nil {
		return false
	}

	if os.IsNotExist(err) {
		ctx.SetStatusCode(http.StatusNotFound)
		return true
	}

	ctx.SetStatusCode(http.StatusInternalServerError)
	logs.ErrorLog(errors.Wrap(err, "SetStatusCode"))

	return true
}
