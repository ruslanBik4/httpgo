// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/ruslanBik4/httpgo/models/logs"
	"github.com/ruslanBik4/httpgo/models/services"
	"github.com/ruslanBik4/httpgo/views"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func HandleAddPhoto(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName == "") || (id == "") {
		views.RenderBadRequest(w)
		return
	}

	const _24K = (1 << 10) * 24
	r.ParseMultipartForm(_24K)
	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			//var err interface{}
			inFile, _ := header.Open()

			path := filepath.Join(tableName, id, header.Filename)
			err := services.Send("photos", "save", path, inFile)
			if err != nil {
				switch err.(type) {
				case services.ErrServiceNotCorrectOperation:

					logs.ErrorLog(err.(error))
				}
				views.RenderInternalError(w, err)

			} else {
				w.Write([]byte("Succesfull - " + header.Filename))
			}
		}
	}
	w.Write([]byte("\nDone"))

}
func HandlePhotos(w http.ResponseWriter, r *http.Request) {

	tableName := r.FormValue("table")
	id := r.FormValue("id")
	num := r.FormValue("num")

	if (tableName == "") || (id == "") {
		views.RenderBadRequest(w)
		return
	}

	number, err := strconv.Atoi(num)
	if err != nil {
		number = -1
	}
	result, err := services.Get("photos", tableName, id, number)
	if err != nil {
		views.RenderInternalError(w, err)
	}

	switch ioReader := result.(type) {
	case []string:
		views.RenderStringSliceJSON(w, ioReader)
	case *os.File:
		//Get the file size
		FileStat, _ := ioReader.Stat()                     //Get info from file
		FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string
		//Send the headers
		w.Header().Set("Content-Description", "File Transfer")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+tableName+id+".jpg")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Cache-Control", "must-revalidate")
		//w.Header().Set("Content-Length", FileSize)

		img, str, err := image.Decode(ioReader)
		if err != nil {
			logs.ErrorLog(err)
			views.RenderInternalError(w, err)

		}

		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: 90}); err != nil {
			views.RenderInternalError(w, err)
		} else {
			logs.DebugLog(str, FileSize, img.Bounds())
		}
	}

}
