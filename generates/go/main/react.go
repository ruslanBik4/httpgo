/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Першій пріватний програміст.
 */

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/ruslanBik4/logs"
)

func ImportReact(dst string, args ...string) error {
	return nil
	cmd := exec.Command("git", args...)
	cmd.Dir = path.Join(dst, "static")
	err := MakeSrcDir(cmd.Dir)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	err = cmd.Run()
	if err != nil {
		logs.ErrorLog(errors.Wrap(err, "creator"))
	}

	tplDir := path.Join(cmd.Dir, "tpl")
	err = MakeSrcDir(tplDir)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	f, err := os.Open(path.Join(cmd.Dir, "react", "public", "index.html"))
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	defer func() {
		err := f.Close()
		if err != nil {
			logs.ErrorLog(err)
		}
	}()

	out, err := os.Create(path.Join(tplDir, "index.qtpl"))
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	defer func() {
		<-time.After(time.Second * 5)
		err := out.Close()
		if err != nil {
			logs.ErrorLog(err)
		}
	}()

	_, err = out.WriteString(`
	{% func Index(publicUrl string) %}
`)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	b = bytes.ReplaceAll(b, []byte("%PUBLIC_URL%"), []byte("{%s publicUrl %}"))
	_, err = out.Write(b)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	_, err = out.WriteString(`
	{% endfunc %}
`)
	if err != nil {
		return errors.Wrap(err, "creator")
	}

	return nil
}
