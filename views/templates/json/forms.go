// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import "github.com/ruslanBik4/httpgo/views/templates/forms"

type FormStructure struct {
	Action          string
	ClassCSS        string
	IdCSS           string
	Name            string
	Events          map[string] string
}

func (thisForm *FormStructure) setFormDefaults(ns *forms.FieldsTable) {

	if thisForm.Action == "" {
		thisForm.Action = "/admin/exec/"
	}
	if thisForm.IdCSS == "" {
		thisForm.IdCSS = "f" + ns.Name
	}
	if thisForm.Name == "" {
		thisForm.Name = ns.Name
	}

	if thisForm.Events != nil {

	if _, ok := thisForm.Events["successSaveForm"]; !ok {

		if ns.SaveFormEvents != nil {
			if val, ok1 := ns.SaveFormEvents["successSaveForm"]; ok1 {
			thisForm.Events["successSaveForm"] = val
			}
		} else {
			thisForm.Events["successSaveForm"] = "afterSaveAnyForm"
		}
	}
	} else {

		thisForm.Events = make(map[string]string, 0)
		if str, ok := ns.DataJSOM["onload"]; ok {
			thisForm.Events["onload"] = str.(string)
		}
		thisForm.Events["onsubmit"] = "saveForm"
		thisForm.Events["oninput"] = "formInput(this);"
		thisForm.Events["onreset"] = "formReset(this);"
		if ns.SaveFormEvents != nil {
			if val, ok := ns.SaveFormEvents["successSaveForm"]; ok {
			thisForm.Events["successSaveForm"] = val
			}
		} else {
			thisForm.Events["successSaveForm"] = "afterSaveAnyForm"
		}
	}
}