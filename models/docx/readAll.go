package docx

import (
 "github.com/sajari/docconv"
 "log"
)

func GetPlainText(fileName string) string{
  if responce, err := docconv.ConvertPath(fileName); err != nil {
    log.Println(err)
          return ""
  } else {
    return responce.Body
  }
}