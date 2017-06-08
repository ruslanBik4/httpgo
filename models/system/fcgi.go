package system

import (
	"net/http"
	"bitbucket.org/PinIdea/fcgi_client"
	"io"
	"path/filepath"
	"strings"
	"os"
	"github.com/ruslanBik4/httpgo/models/logs"
)
const internalRewriteFieldName  = "travel"
var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
)
type FCGI struct{
	Sock string
	Env func(r *http.Request) map[string]string
}
func (c *FCGI) defaultEnv(r *http.Request) map[string]string {
	return map[string]string{
		"REQUEST_METHOD":    r.Method,
		"SCRIPT_FILENAME":   r.URL.Path,
		"SCRIPT_NAME":       r.URL.Path,
		"QUERY_STRING":      r.URL.RawQuery,
	}
}
func (c *FCGI) Do(r *http.Request) (*http.Response, error){
	const typeSckt = "unix" // or "unixgram" or "unixpacket"

	fcgi, err := fcgiclient.Dial(typeSckt, c.Sock)
	if err != nil {
		return nil, err
	}
	env := c.Env
	if env == nil {
		env = c.defaultEnv
	}
	params := env(r)

	var resp *http.Response
	switch r.Method {
	case "GET":
		resp, err = fcgi.Get(params)
	case "POST":
		resp, err = fcgi.Post(params, params["CONTENT_TYPE"], r.Body, int(r.ContentLength) )
	}
	return resp, err
}
func (c *FCGI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := c.Do(r)
	if WriteError(w, err) {
		logs.ErrorLog(err, r.RequestURI, r)
		return
	}
	status, isStatus := resp.Header["Status"]
	location, isURL  := resp.Header["Location"]
	if isStatus && (status[0] == "302 Found") && isURL{
		http.Redirect (w, r, location[0], http.StatusTemporaryRedirect)
		return
	}

	defer resp.Body.Close()
	headers := w.Header()
	for key, val := range resp.Header {
		headers[key] = val
	}
	if _, err := io.Copy(w, resp.Body); WriteError(w, err) {

		logs.ErrorLog(err, r.RequestURI)
	}
}

func NewFPM(sock string) *FCGI {
	return &FCGI{Sock:sock}
}
func NewPHP(root string, sock string) *FCGI {
	return &FCGI{
		Sock:sock,
		Env: func(r *http.Request) map[string]string {

			ip, port   := r.RemoteAddr, ""
			if idx := strings.LastIndex(ip, ":"); idx > -1 {
				port = ip[idx+1:]
				ip   = ip[:idx]
			}
			pathInfo, docURI := "", r.URL.RequestURI()

			if idx := strings.Index(docURI, pathInfo); idx > -1 {
				docURI = docURI[len(pathInfo):]
			}
			// Some variables are unused but cleared explicitly to prevent
			// the parent environment from interfering.
			env := map[string]string{

				// Variables defined in CGI 1.1 spec
				"AUTH_TYPE":         "", // Not used
				"CONTENT_LENGTH":    r.Header.Get("Content-Length"),
				"CONTENT_TYPE":      r.Header.Get("Content-Type"),
				"GATEWAY_INTERFACE": "CGI/1.1",
				//"PATH_INFO":         pathInfo,
				"QUERY_STRING":      r.URL.RawQuery,
				"REMOTE_ADDR":       ip,
				"REMOTE_HOST":       ip, // For speed, remote host lookups disabled
				"REMOTE_PORT":       port,
				"REMOTE_IDENT":      "", // Not used
				"REMOTE_USER":       "", // Not used
				"REQUEST_METHOD":    r.Method,
				"SERVER_NAME":       r.Host,
				"SERVER_PORT":       ":80", //TODO
				"SERVER_PROTOCOL":   r.Proto,
				"SERVER_SOFTWARE":   "httpGo 0.01",

				// Other variables
				"DOCUMENT_ROOT":   root,
				"DOCUMENT_URI":    docURI,
				"HTTP_HOST":       r.Host, // added here, since not always part of headers
				"REQUEST_URI":     r.URL.RequestURI(),
				"SCRIPT_FILENAME": filepath.Join(root,"index.php"),
				"SCRIPT_NAME":     "/index.php",
			}
			// compliance with the CGI specification that PATH_TRANSLATED
			// should only exist if PATH_INFO is defined.
			// Info: https://www.ietf.org/rfc/rfc3875 Page 14
			//if env["PATH_INFO"] != "" {
			//	env["PATH_TRANSLATED"] = filepath.Join(pathToYii, pathInfo) // Info: http://www.oreilly.com/openbook/cgi/ch02_04.html
			//}

			// Some web apps rely on knowing HTTPS or not
			if r.TLS != nil {
				env["HTTPS"] = "on"
			}

			// Add all HTTP headers (except Caddy-Rewrite-Original-URI ) to env variables
			for field, val := range r.Header {
				if strings.ToLower(field) == strings.ToLower(internalRewriteFieldName) {
					continue
				}
				header := strings.ToUpper(field)
				header = headerNameReplacer.Replace(header)
				env["HTTP_"+header] = strings.Join(val, ", ")
			}
			return env
		},
	}
}
// не уверен, что это должно быть здесь - должен быть какой общий механизм для выдачи такого
func WriteError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return true
	}
	w.WriteHeader(http.StatusInternalServerError)
	logs.ErrorLog(err.(error))
	return true
}
