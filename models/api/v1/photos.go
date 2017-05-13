// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"github.com/ruslanBik4/httpgo/views"
	"github.com/ruslanBik4/httpgo/models/services"
	"io"
	"log"
	"strconv"
	"os"
)

func HandlePhotos(w http.ResponseWriter, r *http.Request) {

	tableName := r.FormValue("table")
	id := r.FormValue("id")

	if (tableName == "") || (id == "id") {
		views.RenderBadRequest(w)
		return
	}

	result, err := services.Get("photos", tableName, id)
	if err != nil {
		views.RenderInternalError(w, err)
	}

	switch ioReader := result.(type) {

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
		w.Header().Set("Content-Length", FileSize)

		if num, err := io.Copy(w, ioReader); err != nil {
			log.Println(err)
		} else if num < FileStat.Size() {
			log.Println(num, FileStat.Size())
		}
	}

}
