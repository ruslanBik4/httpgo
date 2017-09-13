package admin

import (
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/models/users"
	"net/http"
	"strconv"
)

var (
	routes = map[string]http.HandlerFunc{

		"/admin/":               HandlerAdmin,
		"/admin/table/":         HandlerAdminTable,
		"/admin/lists/":         HandlerAdminLists,
		"/admin/row/new/":       HandlerNewRecord,
		"/admin/row/edit/":      HandlerEditRecord,
		"/admin/row/add/":       HandlerAddRecord,
		"/admin/row/update/":    HandlerUpdateRecord,
		"/admin/row/show/":      HandlerShowRecord,
		"/admin/row/del/":       HandlerDeleteRecord,
		"/admin/exec/":          HandlerExec,
		"/admin/schema/":        HandlerSchema,
		"/admin/umutable/":      HandlerUMUTables,
		"/admin/anothersignup/": HandlerSignUpAnotherUser,
	}
)

// RegisterRoutes link handlers in htttp.Handler
func RegisterRoutes() {
	for route, fnc := range routes {
		CheckPermissions(route)
		http.HandleFunc(route, system.WrapCatchHandler(fnc))
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
