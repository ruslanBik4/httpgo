package services

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	mongo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/url"
)

var (
	moderation *mService = &mService{name: "moderation"}
)

type Record struct {
	Config map[string]string
	Data   []url.Values
}

type Struct struct {
	Key  string
	Data string
}

type mService struct {
	name    string
	connect *mongo.Session
	status  string
}

func (moderation *mService) Init() error {

	session, err := mongo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}

	session.SetMode(mongo.Monotonic, true)

	moderation.connect = session
	moderation.status = "ready"
	return nil
}

func (moderation *mService) Connect(in <-chan interface{}) (out chan interface{}, err error) {
	out = make(chan interface{})

	go func() {
		out <- "open"
		for {
			select {
			case v := <-in:
				if v.(string) == "close" {
					moderation.Close(out)
				} else {
					out <- v
				}
			}
		}
	}()
	return out, nil
}

func (moderation *mService) Close(out chan<- interface{}) error {
	close(out)
	return nil
}

func (moderation *mService) Status() string {
	return moderation.status
}

func (moderation *mService) Send(messages ...interface{}) error {

	setData := Record{
		Config: make(map[string]string, 0),
		Data:   make([]url.Values, 0),
	}

	for _, message := range messages {
		switch mess := message.(type) {
		case map[string]string:
			setData.Config["table"] = mess["table"]
			setData.Config["key"] = mess["key"]
			setData.Config["action"] = mess["action"]
		case []url.Values:
			setData.Data = mess
		default:

			return &ErrServiceNotCorrectParamType{
				Name: moderation.name,
			}

		}
	}

	if setData.Config["table"] == "" || setData.Config["key"] == "" ||
		(setData.Config["action"] != "insert" && setData.Config["action"] != "delete") {

		return &ErrServiceNotCorrectParamType{
			Name: moderation.name,
		}
	}

	cConnect := GetMongoCollectionConnect(setData.Config["table"])

	if setData.Config["action"] == "delete" {
		//err := cConnect.Remove(bson.M{"key": setData.Config["key"]})
		err := Send("mongod", setData.Config["table"], "Remove", bson.M{"key": setData.Config["key"]})
		if err != nil {
			return err
		}

		return nil
	}

	checkRow := Struct{}
	err := cConnect.Find(bson.M{"key": setData.Config["key"]}).One(&checkRow)

	if checkRow.Data != "" {
		return &ErrServiceNotCorrectParamType{
			Name: moderation.name,
		}
	}

	data := ToGOB64(setData.Data)

	err = cConnect.Insert(&Struct{setData.Config["key"], data})

	if err != nil {
		return err
	}

	return nil
}

func (moderation *mService) Get(messages ...interface{}) (interface{}, error) {

	getData := Record{
		Config: make(map[string]string),
		Data:   make([]url.Values, 0),
	}

	for _, message := range messages {
		switch mess := message.(type) {
		case map[string]string:
			getData.Config["table"] = mess["table"]
			getData.Config["key"] = mess["key"]
		}
	}

	//cConnect := moderation.connect.DB("newDB").C(getData.Config["table"])
	cConnect := GetMongoCollectionConnect(getData.Config["table"])

	responce := Struct{}

	err := cConnect.Find(bson.M{"key": getData.Config["key"]}).One(&responce)

	if err != nil {
		return nil, err
	}

	data := FromGOB64(responce.Data)

	return data, nil
}

func init() {
	AddService(moderation.name, moderation)
}

func GetMongoConnection() *mongo.Session {
	return moderation.connect
}

func ToGOB64(m []url.Values) string {

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(m); err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func FromGOB64(str string) []url.Values {

	var m []url.Values
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println(`failed base64 Decode`, err)
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	if err := d.Decode(&m); err != nil {
		fmt.Println(`failed gob Decode`, err)
	}
	return m
}
