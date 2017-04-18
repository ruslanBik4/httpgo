package admin

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/models/system"
	"github.com/ruslanBik4/httpgo/models/users"
//	"github.com/ruslanBik4/httpgo/views"
	"strconv"
)

var (
	routes = map[string] http.HandlerFunc {

		"/admin/": HandlerAdmin,
		"/admin/table/": HandlerAdminTable,
		"/admin/lists/": HandlerAdminLists,
		"/admin/row/new/": HandlerNewRecord,
		"/admin/row/edit/": HandlerEditRecord,
		"/admin/row/add/": HandlerAddRecord,
		"/admin/row/update/": HandlerUpdateRecord,
		"/admin/row/show/": HandlerShowRecord,
		"/admin/row/del/" : HandlerDeleteRecord,
		"/admin/exec/": HandlerExec,
		"/admin/schema/": HandlerSchema,
		"/admin/umutable/": HandlerUMUTables,
		"/admin/anothersignup/": HandlerSignUpAnotherUser,
	}
)

func RegisterRoutes() {
	for route, fnc := range routes {
		CheckPermissions(fnc, route)
		http.HandleFunc(route, system.WrapCatchHandler(fnc))
	}
}

func CheckPermissions(fnc http.HandlerFunc, route string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user_id,_ := strconv.Atoi(users.IsLogin(r))

		if !GetUserPermissionForPageByUserId(user_id, route, "View") {
//			views.RenderNoPermissionPage(w, r)

		}
	})
}