// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/services"
	"github.com/ruslanBik4/httpgo/views"
)

// HandleAddPhoto добавление фото к обьекту (по имени таблицы и Айди)
//@/api/v1/photos/add/
func HandleAddPhoto(w http.ResponseWriter, r *http.Request) {
	const _24K = (1 << 10) * 24
	r.ParseMultipartForm(_24K)

	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName == "") || (id == "") {
		views.RenderNotParamsInPOST(w, "table", "id")
		return
	}

	if r.MultipartForm == nil {
		views.RenderNotParamsInPOST(w, "file", "empty list")
		return

	}
	w.Write([]byte("{"))
	comma := ""

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			//var err interface{}
			inFile, _ := header.Open()
			err := services.Send("photos", tableName, id, inFile, header.Filename)
			if err != nil {
				views.RenderInternalError(w, err)
				return
			}

			w.Write([]byte(comma + `"` + header.Filename + `": "succesfull"`))
			comma = ","
		}
	}
	w.Write([]byte("}"))

}

// HandlePhotos возвращает ссылки на фото обьекта либо файл
// @/api/v1/photos/
func HandlePhotos(w http.ResponseWriter, r *http.Request) {

	tableName := r.FormValue("table")
	id := r.FormValue("id")
	num := r.FormValue("num")

	if (tableName == "") || (id == "") {
		views.RenderNotParamsInPOST(w, "table", "id")
		return
	}

	number, err := strconv.Atoi(num)
	if err != nil {
		number = -1
	}
	result, err := services.Get("photos", tableName, id, number)
	if err == nil {

		switch ioReader := result.(type) {
		case []string:
			views.RenderStringSliceJSON(w, ioReader)
		case *os.File:
			//Get the file size
			FileStat, _ := ioReader.Stat() //Get info from file
			fileName, fileExt := filepath.Base(FileStat.Name()), filepath.Ext(FileStat.Name())
			FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
			//Send the headers
			w.Header().Set("Content-Description", "File Transfer")
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
			w.Header().Set("Content-Transfer-Encoding", "binary")
			w.Header().Set("Cache-Control", "must-revalidate")

			if fileExt == ".jpg" {

				var str string
				var img image.Image
				img, str, err = image.Decode(ioReader)

				ioReader.Name()
				if err == nil {
					err = jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
					if err == nil {
						logs.DebugLog(str, FileSize, img.Bounds())
					}
				}
			} else {
				w.Header().Set("Content-Length", FileSize)
				io.Copy(w, ioReader)
			}

		}
	}

	if err != nil {
		views.RenderInternalError(w, err)
		return
	}

}
