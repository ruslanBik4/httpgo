// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package services

import (
	"fmt"
	"go/types"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/ruslanBik4/dbEngine/dbEngine"
	"github.com/ruslanBik4/logs"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ruslanBik4/httpgo/apis"
	"github.com/ruslanBik4/httpgo/views"
)

// parameters of main requests
var (
	paramsForLogs = []apis.InParam{
		{
			Name:     paramsSystemctlUnit,
			Type:     apis.NewTypeInParam(types.String),
			DefValue: "httpgo",
		},
		{
			Name: paramDate,
			Type: apis.NewTypeInParam(types.String),
		},
		{
			Name: paramTime,
			Type: apis.NewTypeInParam(types.String),
		},
		{
			Desc: "show only last {ago} hours",
			Name: paramAgo,
			Type: apis.NewTypeInParam(types.Int8),
		},
		{
			Desc: "pattern for log filter",
			Name: paramPatter,
			Type: apis.NewTypeInParam(types.String),
		},
	}
)
var systemRoutes = apis.ApiRoutes{
	ShowStatus: {
		Fnc:    HandleStatusServer,
		Desc:   "view status server",
		Params: paramsForLogs,
	},
	ShowDBStatus: {
		Fnc:    HandleStatusDB,
		Desc:   "view status server",
		Params: paramsForLogs,
	},
	ShowPsqlLog: {
		Fnc:    HandleShowPostgresLog,
		Desc:   "view log Postgres 11",
		Params: paramsForLogs,
	},
	ShowStatusServices: {
		Fnc:    HandleStatusServices,
		Desc:   "view status services",
		Params: paramsForLogs,
	},
	ShowLog: {
		Fnc:    HandleShowLogServer,
		Desc:   "view full log on server ",
		Params: paramsForLogs,
	},
	ShowDebugLog: {
		Fnc:    HandleShowDebugServer,
		Desc:   "view debug log on server ",
		Params: paramsForLogs,
	},
	ShowErrorsLog: {
		Fnc:    HandleShowErrorsServer,
		Desc:   "view error log on server ",
		Params: paramsForLogs,
	},
	ShowInfoLog: {
		Fnc:    HandleShowStatusServer,
		Desc:   "view info log on server",
		Params: paramsForLogs,
	},
	// ShowFEUpdateLog: {
	// 	Fnc:  HandleShowErrorsFrontUpdate,
	// 	Desc: "view error log on server ",
	// },
	// "/api/system/stat_conn/": {
	// 	Fnc: HandleStatConn,
	// 	Params: []apis.InParam{
	// 		{
	// 			Name: "cmd",
	// 			Desc: "command for pgx",
	// 			Req:  false,
	// 			Type: apis.NewTypeInParam(types.String),
	// 		},
	// 	},
	// },
}

// HandleLogServer show status httpgo
// @/api/status/
func HandleStatusServer(ctx *fasthttp.RequestCtx) (interface{}, error) {
	unitName := ctx.UserValue(paramsSystemctlUnit).(string)
	cmd := exec.Command("systemctl", "status", unitName, "-l")

	stdoutStderr, err := cmd.CombinedOutput()

	return nil, views.RenderOutput(ctx, stdoutStderr, err)
}

// HandleLogServer show status services
// @/api/status/services
func HandleStatusServices(ctx *fasthttp.RequestCtx) (interface{}, error) {
	return Status("all"), nil
}

// HandleLogServer show status httpgo
// @/api/status/
func HandleStatusDB(ctx *fasthttp.RequestCtx) (interface{}, error) {

	DB, ok := ctx.UserValue("DB").(*dbEngine.DB)
	if !ok {
		return nil, dbEngine.ErrDBNotFound
	}

	return DB.Conn.SelectToMaps(ctx, `SELECT pid, age(query_start, clock_timestamp()), usename, query, state, 									backend_type, wait_event_type, wait_event  FROM pg_stat_activity
									WHERE query NOT ILIKE '%pg_stat_activity%'
									ORDER BY query_start desc;`)
}

// HandleShowPostgresLog show services errors
// @/api/status/psql
func HandleShowPostgresLog(ctx *fasthttp.RequestCtx) (interface{}, error) {

	stdout, err := RunPostgresqlLog("ERROR.*")
	return nil, views.RenderOutput(ctx, stdout, err)
}

// HandleShowErrorsServer show services errors
// @/api/log/errors/
func HandleShowErrorsServer(ctx *fasthttp.RequestCtx) (interface{}, error) {

	return getLogOutput(ctx, getParamLog("ERROR"))
}

// HandleShowStatusServer show status message
// @/api/log/info/
func HandleShowStatusServer(ctx *fasthttp.RequestCtx) (interface{}, error) {

	return getLogOutput(ctx, getParamLog("INFO"))
}

// HandleShowDebugServer show debug messages
// @/api/log/debug/
func HandleShowDebugServer(ctx *fasthttp.RequestCtx) (interface{}, error) {

	return getLogOutput(ctx, getParamLog("DEBUG"))
}

// HandleShowLogServer show logs messages
// @/api/log/
func HandleShowLogServer(ctx *fasthttp.RequestCtx) (interface{}, error) {

	return getLogOutput(ctx, "")
}

func getLogOutput(ctx *fasthttp.RequestCtx, params string) (interface{}, error) {
	var cmd *exec.Cmd

	showLogCmd, ok := ctx.UserValue(SHOW_LOG_CMD).(string)
	if ok {
		cmd = exec.Command(showLogCmd)
	} else {
		cmd = runJournal(ctx)
	}

	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(stdout) == 0 {
		return "log is empty", nil
	}

	p, ok := ctx.UserValue(paramPatter).(string)
	if ok {
		params += ".*" + p + ".*"
	}

	if params == "" {
		return nil, views.RenderOutput(ctx, stdout, err)
	}

	buf, err := RunGrep(stdout, "--color=never", "-oP", params)
	if err != nil {
		if strings.Contains(err.Error(), "exit status") {
			return "log is empty - " + err.Error(), nil
		}

		return nil, views.RenderOutput(ctx, buf, err)
	}

	return nil, views.RenderOutput(ctx, buf, err)
}

// logs token for journalctl
const (
	startGrepParam = `\[\[`
	endGrepParam   = `(\x1B\[0m)?\]\]\K(.*)`
)

func getParamLog(flagLog string) string {

	logParam := flagLog

	return logParam + endGrepParam
}

func RunGrep(stdout []byte, params ...string) ([]byte, error) {
	cmd := exec.Command("grep", params...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		_, err = stdin.Write(stdout)
	}()

	return cmd.CombinedOutput()
}

func runJournal(ctx *fasthttp.RequestCtx) *exec.Cmd {
	var since string
	h, ok := ctx.UserValue(paramAgo).(int8)
	if ok {
		since = fmt.Sprintf("%d hour ago", h)
	} else {
		since, ok = ctx.UserValue(paramDate).(string)
		if !ok {
			since = time.Now().Format("2006-01-02")
		}

		if timeSince, ok := ctx.UserValue(paramTime).(string); ok {
			since = since + ` ` + timeSince
		}
	}

	unitName := ctx.UserValue(paramsSystemctlUnit).(string)

	return exec.Command("sudo", "journalctl", "-u", unitName, "-o", "cat", "--since", since)
}

func RunPostgresqlLog(params string) ([]byte, error) {
	dayWeek := time.Now().Weekday().String()
	fileName := "/var/lib/pgsql/11/data/log/postgresql-" + dayWeek[:3] + ".log"
	logs.DebugLog(fileName)
	cmd := exec.Command("sudo", "cat", fileName)

	stdout, err := cmd.Output()
	if err != nil {
		return stdout, err
	}

	return RunGrep(stdout, params)
}

type ShowLogsEngine struct {
	name, status string
}

func (s ShowLogsEngine) Init(ctx context.Context) error {
	mapR, ok := ctx.Value("mapRouting").(apis.MapRoutes)
	if ok {
		badRoutings := mapR.AddRoutes(systemRoutes)
		if len(badRoutings) > 0 {
			logs.ErrorLog(apis.ErrRouteForbidden, badRoutings)
			s.status = apis.ErrRouteForbidden.Error()
			return errors.Wrap(apis.ErrRouteForbidden, strings.Join(badRoutings, ","))
		}
		s.status = "ready"
	}

	return nil
}

func (s ShowLogsEngine) Send(ctx context.Context, messages ...interface{}) error {
	panic("implement me")
}

func (s ShowLogsEngine) Get(ctx context.Context, messages ...interface{}) (response interface{}, err error) {
	panic("implement me")
}

func (s ShowLogsEngine) Connect(in <-chan interface{}) (out chan interface{}, err error) {
	panic("implement me")
}

func (s ShowLogsEngine) Close(out chan<- interface{}) error {
	panic("implement me")
}

func (s ShowLogsEngine) Status() string {
	return s.status
}

var showLogServ = ShowLogsEngine{name: "showLogs", status: "starting"}

func init() {
	AddService(showLogServ.name, showLogServ)
}
