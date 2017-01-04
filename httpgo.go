package main
import (
	"fmt"
	"strings"
	"path/filepath"
	"net/http"
	"regexp"
	"io/ioutil"
	//"os/exec"
	"log"
	"time"
	"os"
	"github.com/ruslanBik4/httpgo/views/templates/layouts"
	"github.com/ruslanBik4/httpgo/views/templates/pages"
	"github.com/ruslanBik4/httpgo/models/users"
	"github.com/ruslanBik4/httpgo/models/db"
	"github.com/ruslanBik4/httpgo/models/admin"
	"github.com/ruslanBik4/httpgo/views"

	//"io"
	//"bytes"
	"bitbucket.org/PinIdea/fcgi_client"
	//"strconv"
	"path"
)
//go:generate /Users/rus/go/bin/qtc -dir=views/templates

const portHTTP = ":80"
const pathServer = "/home/travel/"
const pathToYii  = "/home/www/web/"
const php_fpmSCK = "/var/run/php5-fpm.sock"
const internalRewriteFieldName  = "travel"
var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
	// ErrIndexMissingSplit describes an index configuration error.
	//ErrIndexMissingSplit = errors.New("configured index file(s) must include split value")
	pathToHost string
	debug bool
	cssCache = map[string] []byte {}
	routes = map[string] func(w http.ResponseWriter, r *http.Request) {
		"/": handlerDefault,
		"/main/": handlerMainContent,
		"/query/": db.HandlerDBQuery,
		"/admin/": admin.HandlerAdmin,
		"/admin/table/": admin.HandlerAdminTable,
		"/admin/lists/": admin.HandlerAdminLists,
		"/admin/row/new/": admin.HandlerNewRecord,
		"/admin/row/edit/": admin.HandlerEditRecord,
		"/admin/row/add/": admin.HandlerAddRecord,
		"/admin/row/update/": admin.HandlerUpdateRecord,
		"/menu/" : handlerMenu,
		"/show/forms/": handlerForms,
		"/user/signup/": users.HandlerSignUp,
		"/user/signin/": users.HandlerSignIn,
		"/user/signout/": users.HandlerSignOut,
		"/user/active/" : users.HandlerActivateUser,
		"/user/profile/": users.HandlerProfile,
		"/user/oauth/":    users.HandlerQauth2,
		"/user/GoogleCallback/": users.HandleGoogleCallback,
		//"/admin/add": handlerAddPage,
		//"/store/nav/": handlerStoreNav,
		//"/admin/catalog/": handlerAddCatalog,
		//"/admin/psd/add" : handlerAddPSD,
		//"/admin/psd/" : handlerShowPSD,
	}

)
func routeRepeate(key string, w http.ResponseWriter, r *http.Request) error {

	if funcName, ok := routes[key]; ok {
		funcName(w, r)
	}

	return nil
}
func getRealPathFromHost( hostName string ) string {

	//if strings.HasSuffix(hostName, portHTTP) {
	//	hostName = hostName[:strings.Index(hostName, portHTTP)]
	//}
	//return pathServer + hostName + "/"
	return pathServer
}
func sockCatch() {
	err := recover()
	log.Println(err)
}
func doSocket(fcgi_params map[string]string, r *http.Request) (content []byte, err error){

	typeSckt := "unix" // or "unixgram" or "unixpacket"

	fcgi, err := fcgiclient.Dial(typeSckt, php_fpmSCK)
	if err != nil {
		return []byte("dial"), err
	}

	var resp *http.Response
	switch fcgi_params["REQUEST_METHOD"] {
	case "GET":
		resp, err = fcgi.Get(fcgi_params)
	case "POST":
		resp, err = fcgi.Post(fcgi_params, fcgi_params["CONTENT_TYPE"], r.Body, int(r.ContentLength) )
	}
	if err != nil {
		return nil, err
	}

	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}
func fpmRun(w http.ResponseWriter, r *http.Request) {
	env := map[string]string{
		"QUERY_STRING":      r.URL.RawQuery,
		"REQUEST_METHOD":    r.Method,
		"SCRIPT_FILENAME":   r.URL.Path,
		"SCRIPT_NAME":       r.URL.Path,
	}
	if out,err := doSocket(env, r); err == nil {
		w.Write( out )
	} else {
		log.Printf("%s : %v", out, err)
	}
}
func runPHP(filename string, w http.ResponseWriter, r *http.Request) (out []byte, err error){

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
		"SERVER_PORT":       portHTTP,
		"SERVER_PROTOCOL":   r.Proto,
		"SERVER_SOFTWARE":   "httpGo 0.01",

		// Other variables
		"DOCUMENT_ROOT":   pathToYii,
		"DOCUMENT_URI":    docURI,
		"HTTP_HOST":       r.Host, // added here, since not always part of headers
		"REQUEST_URI":     r.URL.RequestURI(),
		"SCRIPT_FILENAME": pathToYii + filename,
		"SCRIPT_NAME":     filename,
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

	defer Catch(w)
	if out,err = doSocket(env, r); err == nil {
		return out, nil
	} else {
		log.Printf("%s : %v", out, err)
	}
	return  nil, err
}
func readFile(filename string, w http.ResponseWriter, r *http.Request) ([]byte, error){

	keyName := path.Base(filename)
	if out, ok := cssCache[keyName]; ok {

		return out, nil
	}
	body, err := ioutil.ReadFile(pathToHost + filename);
	if err != nil {

		if os.IsNotExist(err) {
			if body, err = ioutil.ReadFile(pathToYii + filename); err != nil {
				log.Println(err)
				return nil, err
			}

		}
	}

	cssCache[keyName] = body
	return body, err

}
func handlerDefault(w http.ResponseWriter, r *http.Request) {
	var staticValidator = regexp.MustCompile("^([\\w]+/?)*.(svg)|(css)|(js)|(map)|(ttf)$")
	var htmlValidator = regexp.MustCompile("^([\\wА-Яа-я-_]+/?)*.(html?)|(css)|(js)$")
	var imageValidator = regexp.MustCompile("^([a-zA-Z0-9-_]+/?)*.(jpe?g)|(png)|(ico)|(gif)$")
	var phpValidator = regexp.MustCompile("^([a-zA-Z0-9-_]+/?)*.php(\\?\\w*)?$")
	var fpmValidator = regexp.MustCompile("^(status|ping|pong)$")
	//var dirValidator = regexp.MustCompile("^([a-zA-Z0-9-_/.]+/?)*/$")
	// 	var body []byte

	pathToHost = getRealPathFromHost(r.Host)

	filename := r.URL.Path[1:]

	if staticValidator.MatchString(filename){
		body, err := readFile(filename, w, r)
		if err != nil {
			log.Println("Error during reading file ", filename, err )
		}
		if filename[len(filename)-4:] == ".css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		w.Write(body)
	} else if htmlValidator.MatchString(filename) {
		body, _ := ioutil.ReadFile(pathToHost + filename)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(body)
	} else if (r.URL.Path == "/") {
		views.RenderTemplate(w, r, "index", &pages.IndexPageBody{Title : "Главная страница"} )
	} else if imageValidator.MatchString(filename) {
		body, _ := ioutil.ReadFile(pathToHost + filename)
		w.Write( body)
	} else if phpValidator.MatchString(filename) {
		if out, err := runPHP(filename, w, r); err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Error during execute %s, %v (%s)", filename, err, out)
		} else {
			w.Write(out)
		}
	} else if fpmValidator.MatchString(filename) {
		fpmRun(w, r)

	//} else if dirValidator.MatchString(filename) {
	//
	//   dirName := filename[:len(filename)-1]
	//   if info, err := os.Stat(pathToHost + dirName); err == nil && info.IsDir() {
	//	   if body, err := ioutil.ReadFile(pathToHost + filename + "index.html"); err != nil {
	//		   log.Println(err)
	//	   } else {
	//		   w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//		   w.Write(body)
	//	   }
	//   }
		// php later
	} else{
		//http.Redirect(w, r, r.Host + ":8080" + r.RequestURI, http.StatusFound)
		//return
		if out, err := runPHP("/index.php", w, r); err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Error during execute %s, %v (%s)", filename, err, out)
		} else {
			w.Write( out )
		}

	}

}
func handlerMainContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "text index page %s %v", "filename", users.GetSession(r, "user") )
}
func handlerForms(w http.ResponseWriter, r *http.Request){
	views.RenderTemplate(w, r, r.FormValue("name") + "Form", &pages.IndexPageBody{Title : r.FormValue("email") } )
}
func isAJAXRequest(r *http.Request) bool {
	return len(r.Header["X-Requested-With"]) > 0
}
func handlerMenu(w http.ResponseWriter, r *http.Request) {
	var menu db.MenuItems

	idMenu := r.RequestURI[6:len(r.URL.Path)-1]
	//log.Println(idMenu)


	var catalog, content string
	// отрисовка меню страницы
	if menu.GetMenu(idMenu) > 0 {

		p := &layouts.MenuOwnerBody{ Title: idMenu, TopMenu: make(map[string] *layouts.ItemMenu, 0)}

		for _, item := range menu.Items {
			p.TopMenu[item.Title] = &layouts.ItemMenu{ Link: "/menu/" + item.Name + "/" }

		}

		// return into parent menu if he occurent
		if menu.Self.ParentID > 0 {
			p.TopMenu["< на уровень выше"] = &layouts.ItemMenu{ Link: fmt.Sprintf("/menu/%d/", menu.Self.ParentID ) }
		}
		catalog = p.MenuOwner()
	}
	//для отрисовки контента страницы
	if menu.Self.Link > ""  {
		content = fmt.Sprintf("<div class='autoload' data-href='%s'></div>", menu.Self.Link)
	}
	if isAJAXRequest(r) {
		fmt.Fprintf(w, "%s %s", catalog, content)
	} else {
		pIndex := &pages.IndexPageBody{
			Title: menu.Self.Title,
			Content: content,
			Route: "/menu/" + idMenu + "/",
			Catalog : []string {catalog} }
		views.RenderTemplate(w, r, "index", pIndex)
	}

}
func Catch(w http.ResponseWriter) {
	err := recover()
	if err != nil {
		log.Print("panic runtime! ", err)
	}
}
// считываю счасти из папки
func WalkReadCSS(root string )  filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {

		if info.IsDir() || (filepath.Ext(info.Name()) == ".php") {
			return nil
		}

		keyName := filepath.Base(path)
		if _, ok := cssCache[keyName]; !ok {

			body, err := ioutil.ReadFile(path);
			if err != nil {
				log.Println(err)
				return err
			}
			cssCache[keyName] = body
			log.Println(keyName)
		}
		return  nil
	}
}
func cachingCSSAndJS() {
	filepath.Walk( pathToYii + "assets", WalkReadCSS("") )
	filepath.Walk( pathToYii + "css", WalkReadCSS("") )
	//filepath.Walk( pathToYii + "js", WalkReadCSS("") )
}

func main() {
	go cachingCSSAndJS()

	for route, handler := range routes {

		http.HandleFunc(route, handler)
	}

	log.Println("Server starting in " + time.Now().String() )
	log.Fatal( http.ListenAndServe(portHTTP, nil) )

}