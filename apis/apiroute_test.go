package apis

import (
	//"bufio"
	//"go/types"
	//"net"
	//"sync"
	"encoding/json"
	"testing"

	//"github.com/json-iterator/go"

	"github.com/stretchr/testify/assert"
)

type commCase string

type PRCommandParams struct {
	Command   commCase `json:"command"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	Account   int32    `json:"account"`
	LastQuery commCase `json:"last_query"`
}

// Implementing RouteDTO interface
func (prParams *PRCommandParams) GetValue() interface{} {
	return prParams
}

func (prParams *PRCommandParams) NewValue() interface{} {

	newVal := PRCommandParams{}

	return newVal

}

const jsonText = `{"account":7060246,"command":"adjustments","end_date":"2020-01-25","start_date":"2020-01-01"}`

var (
	route = &ApiRoute{
			Desc:      "test route",
			Method:    POST,
			DTO: &PRCommandParams{},
			}
)

func TestCheckAndRun(t *testing.T) {

	dto := route.DTO.NewValue()
	val := &dto
	//err := jsoniter.UnmarshalFromString(json, &val)
	
	//assert.Nil(t, err)
	
	//t.Logf("%+v", dto)
	
	err := json.Unmarshal([]byte(jsonText), &val)
	
	assert.Nil(t, err)
	
	t.Logf("%+v", dto)
}