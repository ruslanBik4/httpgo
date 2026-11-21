/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package httpGo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// files: name -> content
func makeSite(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		fName := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(fName), 0o755))
		require.NoError(t, os.WriteFile(fName, []byte(content), 0o644))
	}

	return dir
}

func TestFindStartPage(t *testing.T) {
	big := strings.Repeat("x", 5000)
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"index wins", map[string]string{
			"index.html": "a", "actual_Main Page (ENG).html": big, "Fyny.html": big,
		}, "index.html"},
		{"Index.HTM any case", map[string]string{"Index.HTM": "a", "Home.html": big}, "Index.HTM"},
		{"MainPage", map[string]string{"MainPage.html": "a", "About.html": big}, "MainPage.html"},
		{"main_page", map[string]string{"main_page.html": "a", "Frame_2.html": big}, "main_page.html"},
		{"Home Page", map[string]string{"Home Page.html": "a", "Frame_2.html": big}, "Home Page.html"},
		{"головна", map[string]string{"Головна.html": "a", "Контакти.html": big}, "Головна.html"},
		{"whole name before word in name", map[string]string{
			"homepage.html": "a", "actual_Main Page (ENG).html": big,
		}, "homepage.html"},
		{"StarBucks: word main, not the biggest Frame", map[string]string{
			"gomc main.html": "a", "Frame_2.html": big, "iPhone_13_Pro_Max___1.html": big,
		}, "gomc main.html"},
		{"domain is not main", map[string]string{
			"Domain_list.html": "a", "Contacts.html": "<a href='Domain_list.html'>",
		}, "Contacts.html"},
		{"Piza: desktop frame before iPhone", map[string]string{
			"Proekt__Pytstsa_.html": "a", "iPhone_13_Pro_Max___1.html": big,
		}, "Proekt__Pytstsa_.html"},
		{"navigation hub before bigger page", map[string]string{
			"Hub.html":    `<a href="Fyny.html"></a><a href="Pomory.html"></a>`,
			"Fyny.html":   big + `<a href="Hub.html">`,
			"Pomory.html": big,
		}, "Hub.html"},
		{"escaped link counts", map[string]string{
			"A.html":        `<a href="Page%20Two.html">`,
			"Page Two.html": big,
		}, "A.html"},
		{"single page", map[string]string{"Desktop.html": "a", "css/site.css": ""}, "Desktop.html"},
		{"compressed copies are skipped", map[string]string{
			"Desktop.html": "a", "Desktop.html.fasthttp.br": big, "index.html.fasthttp.br": big,
		}, "Desktop.html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findStartPage(makeSite(t, tt.files))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	_, err := findStartPage(makeSite(t, map[string]string{"css/site.css": ""}))
	assert.True(t, errors.Is(err, ErrNoStartPage))
}

func TestFindStartPage_cache(t *testing.T) {
	dir := makeSite(t, map[string]string{"Desktop.html": "a"})
	got, err := findStartPage(dir)
	require.NoError(t, err)
	assert.Equal(t, "Desktop.html", got)

	// a new file changes modification time of dir
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MainPage.html"), []byte("a"), 0o644))
	future := time.Now().Add(time.Second)
	require.NoError(t, os.Chtimes(dir, future, future))
	got, err = findStartPage(dir)
	require.NoError(t, err)
	assert.Equal(t, "MainPage.html", got)
}

func TestNameWords(t *testing.T) {
	assert.Equal(t, []string{"actual", "main", "page", "eng"}, nameWords("actual_Main Page (ENG)"))
	assert.Equal(t, []string{"main", "page"}, nameWords("MainPage"))
	assert.Equal(t, []string{"i", "phone", "13", "pro", "max", "1"}, nameWords("iPhone_13_Pro_Max___1"))
	assert.Equal(t, []string{"головна"}, nameWords("Головна"))
}

func TestSendDomainFile(t *testing.T) {
	dir := makeSite(t, map[string]string{
		"gomc main.html":   "<h1>main</h1>",
		"Frame_2.html":     "<h1>frame</h1>",
		"sub/Desktop.html": "<h1>sub</h1>",
		"css/site.css":     "body{}",
	})
	wd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	run := func(uri string) *fasthttp.RequestCtx {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.SetHost("site.gozaurus.com")
		ctx.Request.SetRequestURI(uri)
		sendDomainFile(ctx, "/")
		return ctx
	}

	ctx := run("/")
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, "<h1>main</h1>", string(ctx.Response.Body()))

	ctx = run("/sub/")
	assert.Equal(t, "<h1>sub</h1>", string(ctx.Response.Body()))

	ctx = run("/sub?lang=en")
	assert.Equal(t, fasthttp.StatusMovedPermanently, ctx.Response.StatusCode())
	assert.Equal(t, "http://site.gozaurus.com/sub/?lang=en", string(ctx.Response.Header.Peek("Location")))

	ctx = run("/css/site.css")
	assert.Equal(t, "body{}", string(ctx.Response.Body()))

	ctx = run("/css/")
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())

	ctx = run("/nope.html")
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}
