// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"github.com/ruslanBik4/httpgo/models/db/schema"
)
type FormStructure struct {
	Action          string
	ClassCSS        string
	IdCSS           string
	Name            string
	Events          map[string] string
}

func (thisForm *FormStructure) setFormDefaults(ns *schema.FieldsTable) {

	if thisForm.Action == "" {
		thisForm.Action = "/admin/exec/"
	}


	if thisForm.Events == nil {

		thisForm.Events = make(map[string]string, 0)
		thisForm.Events["successSaveForm"] = "afterSaveAnyForm"
	}

	if _, ok := thisForm.Events["successSaveForm"]; !ok {

		if ns.SaveFormEvents != nil {
			if val, ok1 := ns.SaveFormEvents["successSaveForm"]; ok1 {
				thisForm.Events["successSaveForm"] = val
			}
		}
	}
	if ns == nil {
		return
	}

	if thisForm.IdCSS == "" {
		thisForm.IdCSS = "f" + ns.Name
	}
	if thisForm.Name == "" {
		thisForm.Name = ns.Name
	}
	if str, ok := ns.DataJSOM["onload"]; ok {
		thisForm.Events["onload"] = str.(string)
	}
}