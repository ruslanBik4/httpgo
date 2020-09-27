// Copyright 2018 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"regexp"
	// "github.com/ruslanBik4/logs"
)

const regHTMLArticle = `\<meta\s+property\="og:title"\s+content\="([^"]+)"\/\>[\s\w\S]+\<article\>(?:[\s\w\S]+\<time\s+datetime\="[\d-T:+\s]+"\>([\d\w\s,:]+)\<\/time\>)?([\s\S]+)\<\/article\>`

var (
	regParse      = regexp.MustCompile(regHTMLArticle)
	regRemoveTags = regexp.MustCompile("<[^>]+>")
	regRuss       = regexp.MustCompile("[А-Яа-я]")
	regEng        = regexp.MustCompile(`^[A-Za-z\d\s:;"']+$`)
)

func GetTags(b []byte) [][][]byte {
	return regParse.FindAllSubmatch(b, 1)
}

func IsRusText(b string) bool {
	return regRuss.MatchString(b)
}

func IsEngText(b string) bool {
	return regEng.MatchString(b)
}

func DecodeHtml() {
}
