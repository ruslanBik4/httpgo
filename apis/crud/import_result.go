/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package crud

// ImportResult is the final message of a bulk import (csv or Excel) written
// to the SSE log of the import: the browser shows Msg and follows FormActions
// (redirect to the browse page of the table).
type ImportResult struct {
	FormActions []FormActions `json:"formActions"`
	Count       int64         `json:"count,omitempty"`
	Msg         string        `json:"message"`
	// Excel holds the details of an Excel import (*excel.Result of
	// github.com/ruslanBik4/dbEngine/dbEngine/excel: matched / skipped
	// columns, parent records found or created, pictures, warnings);
	// nil for a csv import.
	Excel any `json:"excel,omitempty"`
}

// NewImportResult creates the result of an import that redirects to redirectURL.
func NewImportResult(redirectURL string, count int64, msg string) *ImportResult {
	return &ImportResult{
		FormActions: []FormActions{
			{
				Typ: "redirect",
				Url: redirectURL,
			},
		},
		Count: count,
		Msg:   msg,
	}
}
