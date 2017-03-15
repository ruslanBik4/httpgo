// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
	"github.com/ruslanBik4/httpgo/models/docs"
	"fmt"
)
func TestReadGoogleSheets(t *testing.T) {
	var sheet docs.SheetsGoogleDocs
	fileName := ""

	if err := sheet.Init(); err != nil {
		t.Errorf("error initialization: filename=%s, error=%q", fileName, err)
	}

	fmt.Printf("before read")
	if resp, err := sheet.Read(); err != nil {
		t.Errorf("Error during reading sheet %v", err)
	} else {
		fmt.Printf("%v", resp)
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[4])
		}

	}
}
