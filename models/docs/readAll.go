package docs

import (
 "github.com/sajari/docconv"
        "github.com/ruslanBik4/httpgo/models/logs"
)

func GetPlainText(fileName string) string{
  if responce, err := docconv.ConvertPath(fileName); err != nil {
    logs.ErrorLog(err)
          return ""
  } else {
    return responce.Body
  }

}