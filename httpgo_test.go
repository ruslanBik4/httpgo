// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
	"github.com/ruslanBik4/httpgo/models/docs"
	"fmt"
)
const spreadsheetId = "1EvNM788L-CC7N1kYIieQZEuinpmI7yVzu_mV75DF3cM"

func TestReadGoogleSheets(t *testing.T) {
	var sheet docs.SheetsGoogleDocs
	fileName := ""

	if err := sheet.Init(); err != nil {
		t.Errorf("error initialization: filename=%s, error=%q", fileName, err)
	}

	fmt.Printf("before read")
	readRange := "Шаблон!"
	if resp, err := sheet.Read(spreadsheetId, readRange); err != nil {
		t.Errorf("Error during reading sheet %v", err)
	} else {
		fmt.Printf("%v", resp)
		for idx,  row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			for idx := range row {
				fmt.Printf("%s | ", row[idx])
			}
			fmt.Printf(", %s\n", idx)
		}

	}
}

