/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package layouts

import (
	"bytes"
	"strings"
	"testing"
)

func TestHeadDir(t *testing.T) {
	for lang, want := range map[string]string{"he": "rtl", "he-IL": "rtl", "ar,en;q=0.8": "rtl", "AR": "rtl", "en": "ltr", "uk": "ltr", "": "ltr"} {
		if got := (&HeadHTMLPage{Language: lang}).Dir(); got != want {
			t.Errorf("%q: %s, want %s", lang, got, want)
		}
	}
	var b bytes.Buffer
	(&HeadHTMLPage{Language: "he", Charset: "utf-8"}).WriteHeadHTML(&b)
	if !strings.Contains(b.String(), `<html lang="he" dir="rtl">`) {
		t.Errorf("%.200s", b.String())
	}
}
