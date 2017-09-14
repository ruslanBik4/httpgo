package admin

import (
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/models/users"
	"net/http"
	"strconv"
)

var (
	routes = map[string]http.HandlerFunc{

		"/admin/":               handlerAdmin,
		"/admin/table/":         handlerAdminTable,
		"/admin/lists/":         handlerAdminLists,
		"/admin/row/new/":       handlerNewRecord,
		"/admin/row/edit/":      handlerEditRecord,
		"/admin/row/add/":       handlerAddRecord,
		"/admin/row/update/":    handlerUpdateRecord,
		"/admin/row/show/":      handlerShowRecord,
		"/admin/row/del/":       handlerDeleteRecord,
		"/admin/exec/":          HandlerExec,
		"/admin/schema/":        handlerSchema,
		"/admin/umutable/":      handlerUMUTables,
		"/admin/anothersignup/": HandlerSignUpAnotherUser,
	}
)

// RegisterRoutes link handlers in htttp.Handler
func RegisterRoutes(MyMux *http.ServeMux) {
	for route, fnc := range routes {
		CheckPermissions(route)
		MyMux.HandleFunc(route, system.WrapCatchHandler(fnc))
	}
}

// CheckPermissions is not complete yet
func CheckPermissions(route string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, _ := strconv.Atoi(users.IsLogin(r))

		if !GetUserPermissionForPageByUserId(userId, route, "View") {
			//			views.RenderNoPermissionPage(w, r)

		}
	})
}
