package docx

import (
 "github.com/sajari/docconv"
 "log"
)

func getPlainText(fileName string) string{
  if responce, err := docconv.ConvertPath(fileName); err != nil {
    log.Println(err)
  } else {
    return responce.Body
  }
}