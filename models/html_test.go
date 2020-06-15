// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/ruslanBik4/httpgo/logs"
)

func TestGetTags(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"/Users/ruslan/Downloads/DataClusteringSample2703/20200427", //01",//9024187325606168587.html",
			args{},
		},
	}
	for _, tt := range tests {
		t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			ff, err := filepath.Glob(tt.name + "/*/*.html")
			if err != nil {
				t.Error(errors.Wrap(err, ""))
			}

			i, j, e := 0, 0, 0
			w := sync.WaitGroup{}
			for _, name := range ff {

				b, err := ioutil.ReadFile(name)
				if err != nil {
					t.Error(err, tt.name)
					continue
				}

				w.Add(1)
				go func(b []byte, name string) {
					defer w.Done()
					req := GetTags(b)
					if len(req) == 0 || len(req[0]) < 2 {
						t.Error(errors.New("wrong text"), name)

						return
					}

					title := html.UnescapeString(string(req[0][1]))
					if IsRusText(title) {
						logs.StatusLog("Заголовок: '%s' Дата: %s\n русский %25.25s", title, req[0][2], regRemoveTags.ReplaceAll(req[0][3], nil))
						j++
					} else if IsEngText(title) {
						logs.StatusLog("Title: '%s' Data: %s \n eng: %25.25s", title, req[0][2],
							bytes.TrimLeft(regRemoveTags.ReplaceAll(req[0][3], nil), " "))
						e++
					}
					// logs.StatusLog("Text: %s", regRemoveTags.ReplaceAll( req[0][3], nil) )

				}(b, name)
				i++
			}
			w.Wait()
			logs.StatusLog("%d files, %d rus article, %d eng articles", i, j, e)
		})
	}
}
