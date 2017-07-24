package services

import (
    //"github.com/ruslanBik4/httpgo/models/db"
    //"models/permissions"
    "errors"
    //"log"
    //logtravel "models/log"
    //"time"
    "github.com/ruslanBik4/httpgo/models/db"
    "time"
    "database/sql"
    //"net/http"
    //"github.com/ruslanBik4/httpgo/views/templates/json"
    "github.com/ruslanBik4/httpgo/models/logs"
    "github.com/ruslanBik4/httpgo/views/templates/json"
)

type LogParam struct {
    module string;
    table  string;
    sqlBeg string;
    sqlEnd string;
    sql string;
}



//type LogParamsMap map[string]LogParams;
//type LogParams map[string]LogParam;

type DataLogSave struct {
    TypeData    string;    //Тип логирования - json для данных, text - для текстовой информации для одного параметра
    module      string;    //Раздел логирования
    template    string;    //Шаблон для для единовременной специальной настройки логирования
    table       string;    //Таблица с исходными данными
    id_in_table int;       //ID записи в таблице "table"
    event       string;    //Событие вызвавшее логирование
    id_users    int;       //ID юзера действие, которого нужно залогировать
    datetime    string;    //Дата логирования
    data           string; //Изменные данные (снимок текущего состояния)
    data_old       string; //Только изменившиеся данные старые значения
    data_new       string; //Только изменившиеся данные новые значения
    fields_changed string; //Список изменившихся полей
    sqlBeg         string; //Начало запроса на получение данных до "table". По умолчанию "SELECT * FROM "
    sqlEnd         string; //Конец запроса на получение данных до "table". По умолчанию " WHERE id=?"
    sql            string; //Полностью специализированный запрос. Тогда нет необходимости в "table", "sqlBeg", "sqlEnd"
    sqlLog         string; //Код sql строки логирования
    test           bool;   //Статус тестирования
    args        interface{};
}

type DataSaveMongo struct {
    TypeData       string; //Тип логирования - json для данных, text - для текстовой информации для одного параметра
    Module         string; //Раздел логирования
    Table         string;  //Таблица с исходными данными
    IdInTable     int;     //ID записи в таблице "table"
    Event         string;  //Событие вызвавшее логирование
    IdUsers       int;     //ID юзера действие, которого нужно залогировать
    DateTime      string;  //Дата логирования
    Data          string;  //Изменные данные (снимок текущего состояния)
    DataOld       string;  //Только изменившиеся данные старые значения
    DataNew       string;  //Только изменившиеся данные новые значения
    FieldsChanged string;  //Список изменившихся полей
}

//TODO: rebase test variables to _test file

//LogParamsMap Params for loging (keys - Module, table name
var LogParamsMap = map[string]LogParam{
    "booking_for_book": LogParam{module: "booking", table: "bookings_rooms", sqlBeg: "SELECT * FROM ", sqlEnd : " WHERE id_bookings=?",sql:"" },
    "bookings_rooms": LogParam{module: "booking", table: "bookings_rooms", sqlBeg: "SELECT * FROM ", sqlEnd : " WHERE id_bookings=?", sql:""},
    "test": LogParam{module: "Test", table: "test_data_log", sqlBeg: "SELECT * FROM ", sqlEnd : " WHERE id_test=?", sql:""},
    "test2": LogParam{module: "Test",  sqlBeg: "", sqlEnd : "", sql:`SELECT * FROM test_data_log WHERE id_test=?`},
}

//LogParamDef Default values params
var LogParamDef = LogParam{module: "", table: "", sqlBeg: "SELECT * FROM ", sqlEnd : " WHERE id=?",sql : "" }

//SQLLogTpl - SQL query for insert row in table log
var SQLLogTpl = `INSERT INTO log (module, type_data, table, id_in_table, events, id_users, datetime-sys,
fields_changed,data, data_old, data_new) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
//SQLLogTplTest  - test SQL query delete test rows
var SQLLogTplTestDel = "DELETE FROM log WHERE id=?"
//SQLLogTplTest  - test SQL query for insert row in table log
var SQLLogTplTest = `INSERT INTO log (id, module, type_data, table, id_in_table, events, id_users, datetime-sys,
    fields_changed,data, data_old, data_new)  VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

//TODO: DataLogSave variable name
type lService struct {
    name string
    DataLogSave
    status string
}

var logger *lService = &lService{name:"logger"}


//var loggingData *DataSave = &DataSave


func (logger *lService) Send(messages ...interface{}) error {

    //var err error, err = nil => err := nil
    var err error
    var id int
    curLogParam :=  LogParamDef

    for _, message := range messages {
            switch mess := message.(type) {
            case DataLogSave:
                if (mess.TypeData == "") {
                    mess.TypeData = "json"
                }

                if mess.sqlLog == "" {
                    mess.sqlLog = SQLLogTpl
                }
    
                if mess.TypeData == "json" {
                    if (mess.template !=""){
                        curLogParam = GetParam(mess.template)

                    }
                    if (curLogParam.table !=""){
                        mess.table = curLogParam.table
                    }
                    if (curLogParam.sqlBeg !=""){
                        mess.sqlBeg  = curLogParam.sqlBeg
                    }
                    if (curLogParam.sqlEnd !=""){
                        mess.sqlEnd = curLogParam.sqlEnd
                    }

                    if (curLogParam.sql !=""){
                        mess.sql = curLogParam.sql
                    }

                    paramsGetData :=db.SqlCustom{Table:mess.table,
                        SqlBeg:mess.sqlBeg,
                        SqlEnd:mess.sqlEnd,
                        Sql:mess.sql}
                    
                    rows, row, rowField, columns, colTypes, Err := db.GetDataCustom(paramsGetData, mess.args)
                    if (Err != nil){
                        logs.ErrorLog(Err)
                    }
                    err = mess.LogProcessRows(rows, row, rowField, columns, colTypes)
                    if (err != nil){
                        logs.ErrorLog(err)
                    }
                } else {
                    curTime := time.Now()
                    curdatetime := curTime.Format(time.RFC3339)



                    if (mess.sqlLog ==""){ьуы
                        mess.sqlLog = SQLLogTpl
                    }
                    dataRow := DataLogSave{module: mess.module,
                        TypeData:                  mess.TypeData,
                        table:                     mess.table,
                        id_in_table:               mess.id_in_table,
                        event:                     mess.event,
                        id_users:                  mess.id_users,
                        datetime:                  curdatetime,
                        fields_changed:            "",
                        data:                      mess.data,
                        data_old:                  mess.data_old,
                        data_new:                  mess.data_new}

                    _, err = dataRow.LogSaveRow()
                }

                if mess.id_users == 0 {
                    err = errors.New("Error. Not requred data id_users")
                    logs.ErrorLog(err," id=", id, " err mess.id_users =" )
                }


            default:
                err = errors.New("Error. Not valid structure DataLogSave")
            }
    }

    return err

}

func (logger *lService) Get(messages ... interface{}) (responce interface{}, err error) {

    return nil, nil
}

func (logger *lService) Connect(in <- chan interface{}) (out chan interface{}, err error) {

    return nil, nil
}

func (logger *lService)  Close(out chan <- interface{}) error {
    close(out)
    return nil
}

func (logger *lService)  Status() string {
    return logger.status
}

func (logger *lService) Init() error {
    var err error =nil
    
    status := Status("mongod")
    
    i := 0
    for (status != "ready") && (i < 2500) {
        time.Sleep(5)
        i++
        status = Status("mongod")
        
    }
    if status != "ready"{
        
        err = errors.New("Not Init service mongo")
    } else{
        logger.status = "ready"
    }
    return err
    
}

func init() {
    AddService(logger.name, logger)
}

//LogSaveRow save current log data in database
func (dataSave DataLogSave) LogSaveRow() (id int, err error) {
    //logs.DebugLog("sql =", dataSave.sqlLog)
    if (dataSave.sqlLog == "") {
        dataSave.sqlLog = SQLLogTpl;
    }
    
    //currentTime := int(time.Now().UTC().Unix())
    
    record := &DataSaveMongo{}
    record = &DataSaveMongo{

        TypeData:  dataSave.TypeData,
        Module:    dataSave.module,
        Table:     dataSave.table,
        IdInTable: dataSave.id_in_table,
        IdUsers:   dataSave.id_users,
        DateTime:  dataSave.datetime,
        Data:      dataSave.data,
        DataOld:   dataSave.data_old,
        DataNew:   dataSave.data_new,
        //FieldsChanged:dataSave.FieldsChanged,
    }
    logs.DebugLog(Status("mongod"))

    
    err = Send("mongod", "logger", "Insert", record)
    //logs.DebugLog("sql =", dataSave)
    if err != nil {
        logs.ErrorLog(err)
        return 0, err
    }
    
    //return db.DoInsert(dataSave.sqlLog, dataSave.Module, dataSave.TypeData, dataSave.table, dataSave.id_in_table,
    //    dataSave.event, dataSave.id_users, dataSave.datetime, dataSave.FieldsChanged, dataSave.data, dataSave.data_old,
    //    dataSave.data_new);
    return 0, nil
}


//GetParam function for get data from LogParamsMap or LogParamDef
func GetParam(template  string) (params LogParam) {
    params = LogParamDef;
    for idx, val := range LogParamsMap {
        if (idx == template) {
            params = val
        }
    }

    return params
}

//LogProcessRows - function for processing multirows data
func (dataSave DataLogSave) LogProcessRows(rows *sql.Rows, row [] interface {}, rowField map[string] *sql.NullString,
columns [] string, colTypes [] *sql.ColumnType) ( err error ) {
    var arrJSON =json.MultiDimension{}

    var curDateTime string
    var Err error
    var data string
    //status = "loging ok"
    id := 0;

    data = ""
    for rows.Next() {

        if err := rows.Scan(row... ); err != nil {
            ////PrintErr(err)
            continue
        }

        //PrintTest(w, "\n row =")
        //PrintTest(w, row)
        id, arrJSON ,  Err = db.ConvertPrepareRowToJson(rowField, columns, colTypes)
        data =  json.AnyJSON(arrJSON)
        curTime := time.Now()

        curDateTime = curTime.Format(time.RFC3339)
        dataRow := DataLogSave{module: dataSave.module, TypeData: dataSave.TypeData, table: dataSave.table, id_in_table: id,
            event:                     dataSave.event, id_users: dataSave.id_users, datetime: curDateTime, fields_changed: "",
            data:                      data, data_old: "", data_new: "", sqlLog: dataSave.sqlLog}


        _, Err = dataRow.LogSaveRow()
    }

    return Err
}
