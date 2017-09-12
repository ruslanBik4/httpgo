package docs

import (
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/sajari/docconv"
)

func GetPlainText(fileName string) string {
	if response, err := docconv.ConvertPath(fileName); err != nil {
		logs.ErrorLog(err)
		return ""
	} else {
		return response.Body
	}

}
