package services

//import "time"
import (
    "testing"
    "time"
    "github.com/ruslanBik4/httpgo/models/server"
    "flag"
    //"github.com/ruslanBik4/httpgo/models/db"
    "gopkg.in/mgo.v2/bson"
    "github.com/ruslanBik4/httpgo/models/logs"
    //httpgoService "github.com/ruslanBik4/httpgo/models/services"
    //"views/templates/layouts/extranet/core/objects/services"
)

var NameServ = "logger"
var tableTest = "test_data_log"
var TestMode = false
var TestIdUser = 1
var (
    f_static  = flag.String("path", "/home/serg/work/gocode/src/github.com/ruslanBik4/httpgo", "path to static files")
    f_web     = flag.String("web", "/home/serg/work/ustudio/travel/web", "path to web files")
    f_session = flag.String("sessionPath", "/home/serg/work/session", "path to store sessions data")
)


func TestLoggerSendStruct(t *testing.T) {
    ServerConfig := server.GetServerConfig()
    if err := ServerConfig.Init(f_static, f_web, f_session); err != nil {
        logs.ErrorLog(err)
    }
    
    MongoConfig := server.GetMongodConfig()
    if err := MongoConfig.Init(f_static, f_web, f_session); err != nil {
        logs.ErrorLog(err)
    }
    
    InitServices()
    
    status := Status("logger")
    
    i := 0
    for (status != "ready") && (i < 2500) {
        time.Sleep(5)
        i++
        status = Status("logger")
    }

    
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:         "Test",
        TypeData:       "text",
        table:          "Boooking message",
        id_in_table:    58,
        event:          "test_stuct",
        id_users:       1,
        datetime:       curdatetime,
        fields_changed: "",
        data:           "{ 'a': [ 1, 2, 3 ] }", sqlLog: SQLLogTpl}
    
    err := Send(NameServ, dataTest)
    
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}

func TestLoggerSendNoUser(t *testing.T) {
    
    
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "text", table: tableTest, id_in_table: 1, event: "test_log_text",
        datetime: curdatetime, fields_changed: "", data: "Test Data text Log", sqlLog: SQLLogTpl}
    
    err := Send(NameServ, dataTest)
    logs.DebugLog("err", err)
    if err != nil {
        
        t.Skipped()
    } else {
        t.Error(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}

func TestLoggerSendTable(t *testing.T) {
    
    
    var args interface{}
    
    args = 1
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "json", table: tableTest, event: "test_json_table_args",id_users: TestIdUser,
        datetime: curdatetime, fields_changed: "", data: "Test Data text Log", sqlLog: SQLLogTpl, args: args}
    
    err := Send(NameServ, dataTest)
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}

func TestLoggerSendTable1(t *testing.T) {
    
  
    var args interface{}
    
    args = 1
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "json", table: "bookings", event: "TestLoggerSendTable1",
        id_users: TestIdUser, datetime: curdatetime, fields_changed: "",  sqlLog: SQLLogTpl, args: args}
    
    err := Send(NameServ, dataTest)
    logs.DebugLog("err", err)
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}


func TestLoggerSendTableTemplate(t *testing.T) {
    
  
    var args interface{}
    
    args = 1
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "json", template: "booking_for_book", event: "TestLoggerSendTableTemplate",id_users: TestIdUser,
        datetime: curdatetime, fields_changed: "",  sqlLog: SQLLogTpl, args: args}
    
    err := Send(NameServ, dataTest)
    
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}
func TestLoggerSendTemplate(t *testing.T) {
    
       var args interface{}
    
    
    args = 9
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "json", template: "test", event: "test_json_Template",id_users: TestIdUser,
        datetime: curdatetime, fields_changed: "", data: "Test Data text Log", sqlLog: SQLLogTpl, args: args}
    
    err := Send(NameServ, dataTest)
    
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}

func TestLoggerSendText(t *testing.T) {
    
    
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "text", table: "table", id_in_table: 1, event: "test_log_text", id_users: TestIdUser,
        datetime: curdatetime, fields_changed: "", data: "Test Data text Log", sqlLog: SQLLogTpl}
    
    err := Send(NameServ, dataTest)
    logs.DebugLog("err", err)
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}


func TestLoggerGetText(t *testing.T) {
    
    
    curTime := time.Now()
    curdatetime := curTime.Format(time.RFC3339)
    dataTest := DataLogSave{
        module:   "Test", TypeData: "text", table: "table", id_in_table: 1, event: "test_log_text", id_users: TestIdUser,
        datetime: curdatetime, fields_changed: "", data: "Test Data text Log", sqlLog: SQLLogTpl}
    
    err := Send(NameServ, dataTest)
    logs.DebugLog("err", err)
    if err != nil {
        t.Error(err)
        logs.ErrorLog(err)
    }
    //_ = db.DoQuery(SQLLogTplTestDel, 1)
}

