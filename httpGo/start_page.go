/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"bytes"
	"cmp"
	"io"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

// ErrNoStartPage means the directory has no html file to show as its start page
var ErrNoStartPage = errors.New("no html file for start page")

// startPageNames are whole names of a start page (letters & digits of name, lowercased):
// "MainPage.html", "main_page.html", "Home Page.html", "Головна.html" ...
var startPageNames = map[string]bool{
	"mainpage": true, "homepage": true, "main": true, "home": true,
	"start": true, "startpage": true, "default": true, "welcome": true,
	"landing": true, "landingpage": true, "frontpage": true,
	"головна": true, "головнасторінка": true, "главная": true, "главнаястраница": true,
	"golovna": true, "holovna": true, "glavnaya": true,
}

// startPageWords are words that mark a start page inside a longer name:
// "actual_Main Page (ENG).html", "gomc main.html", "NewHomePage_v2.html"
var startPageWords = []string{"main", "home", "landing", "головна", "главная"}

// desktop & mobile words of names of Figma frames
var (
	desktopWords = []string{"desktop", "macbook", "web", "1920", "1440", "1366", "1280"}
	mobileWords  = []string{"iphone", "android", "mobile", "phone", "390", "375", "360"}
)

// maxLinksScan limits bytes of html file read for counting its links
const maxLinksScan = 1 << 20

type startPage struct {
	modTime time.Time
	name    string
}

// startPages caches start page of directory until directory changes
// (adding, removing or renaming of a file changes modification time of directory)
var startPages sync.Map // dir -> startPage

// findStartPage returns name of the file to show for "/" of dir:
//  1. index.html / index.htm (any case);
//  2. a file whose whole name is a start page name: MainPage.html, home_page.html ...;
//  3. a file whose name has a start page word: "actual_Main Page (ENG).html";
//  4. any html file.
//
// Several files of one rank are ordered by: the number of other html files of dir
// it links to (start page is the navigation hub of site), desktop frame before
// mobile one (Figma: "Desktop", "MacBook" before "iPhone"), larger size, name.
func findStartPage(dir string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.Errorf("'%s' is not directory", dir)
	}

	if v, ok := startPages.Load(dir); ok {
		if sp := v.(startPage); sp.modTime.Equal(info.ModTime()) {
			return sp.name, nil
		}
	}

	name, err := chooseStartPage(dir)
	if err != nil {
		return "", err
	}
	startPages.Store(dir, startPage{modTime: info.ModTime(), name: name})

	return name, nil
}

type pageCandidate struct {
	name  string
	size  int64
	rank  int
	links int
	form  int // 1 desktop, 0 unknown, -1 mobile
}

func chooseStartPage(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	pages := make([]*pageCandidate, 0, len(entries))
	for _, e := range entries {
		ext := strings.ToLower(path.Ext(e.Name()))
		if e.IsDir() || ext != ".html" && ext != ".htm" {
			continue // also skips compressed copies: "index.html.fasthttp.br"
		}
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		base := strings.TrimSuffix(e.Name(), path.Ext(e.Name()))
		words := nameWords(base)
		joined := strings.Join(words, "")
		c := &pageCandidate{name: e.Name(), size: info.Size(), rank: 4, form: frameForm(words)}
		switch {
		case joined == "index":
			// explicit start page, no more checks needed
			return e.Name(), nil
		case startPageNames[joined]:
			c.rank = 2
		case hasAnyWord(words, startPageWords) || strings.Contains(joined, "mainpage") || strings.Contains(joined, "homepage"):
			c.rank = 3
		}
		pages = append(pages, c)
	}

	if len(pages) == 0 {
		return "", errors.Wrapf(ErrNoStartPage, "'%s'", dir)
	}

	best := slices.MinFunc(pages, func(a, b *pageCandidate) int { return a.rank - b.rank }).rank
	pages = slices.DeleteFunc(pages, func(c *pageCandidate) bool { return c.rank != best })
	if len(pages) == 1 {
		return pages[0].name, nil
	}

	for _, c := range pages {
		c.links = countLinks(path.Join(dir, c.name), c.name, pages)
	}
	slices.SortFunc(pages, func(a, b *pageCandidate) int {
		return cmp.Or(
			cmp.Compare(b.links, a.links),
			cmp.Compare(b.form, a.form),
			cmp.Compare(b.size, a.size),
			strings.Compare(a.name, b.name),
		)
	})

	return pages[0].name, nil
}

// countLinks counts other pages (as written in href: plain or url-escaped) mentioned in file
func countLinks(fileName, self string, pages []*pageCandidate) int {
	f, err := os.Open(fileName)
	if err != nil {
		return 0
	}
	defer f.Close()

	content, err := io.ReadAll(io.LimitReader(f, maxLinksScan))
	if err != nil {
		return 0
	}

	n := 0
	for _, c := range pages {
		if c.name == self {
			continue
		}
		if bytes.Contains(content, []byte(c.name)) ||
			bytes.Contains(content, []byte(url.PathEscape(c.name))) {
			n++
		}
	}

	return n
}

// nameWords splits name into lowercased words by non-letters and camelCase:
// "actual_Main Page (ENG)" -> actual main page eng, "MainPage" -> main page
func nameWords(name string) []string {
	var (
		words []string
		cur   []rune
		prev  rune
	)
	flush := func() {
		if len(cur) > 0 {
			words = append(words, strings.ToLower(string(cur)))
			cur = cur[:0]
		}
	}
	for _, r := range name {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			flush()
		case unicode.IsUpper(r) && unicode.IsLower(prev),
			unicode.IsDigit(r) != unicode.IsDigit(prev) && len(cur) > 0:
			flush()
			cur = append(cur, r)
		default:
			cur = append(cur, r)
		}
		prev = r
	}
	flush()

	return words
}

func hasAnyWord(words, marks []string) bool {
	for _, w := range words {
		if slices.Contains(marks, w) {
			return true
		}
	}

	return false
}

func frameForm(words []string) int {
	switch {
	case hasAnyWord(words, desktopWords):
		return 1
	case hasAnyWord(words, mobileWords):
		return -1
	default:
		return 0
	}
}
