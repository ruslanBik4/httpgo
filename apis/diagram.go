/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2exporter"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/ruslanBik4/logs"
)

const (
	tplEndpoint = `%[1]s: %s %s {
	shape: class
	`
	tplCol = `'%s': '%s' # %s
	`
	tplParamDesc = `  
+ **%s**: *%s*`
	tplEndpointEnd = `}
%s {
	shape: package
	'%s' {
		shape: document
	}
}
%[1]s.'%s' -> %s : %d
`
	tplComment = `comment%s: |md
%s  

%s
| {
	shape: page
}

%[1]s -> comment%[1]s {
  source-arrowhead: Desc
  target-arrowhead: * {
    shape: diamond
  }
}
`
)

func GetGraphSVG(ctx *fasthttp.RequestCtx, buf *bytes.Buffer, opts *d2svg.RenderOpts) (any, error) {
	graph, cfg, err := d2compiler.Compile("", buf, &d2compiler.CompileOptions{UTF16Pos: true})
	if err != nil {
		logs.ErrorLog(err)
		return nil, err
	}

	logs.StatusLog(cfg)
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, err
	}
	ruler.LineHeightFactor = .5
	err = graph.SetDimensions(nil, ruler, nil)
	if err != nil {
		return nil, err
	}
	err = d2dagrelayout.Layout(ctx, graph, nil)
	if err != nil {
		return nil, err
	}
	diagram, err := d2exporter.Export(ctx, graph, nil)
	if err != nil {
		return nil, err
	}

	return d2svg.Render(diagram, opts)
}
func getStringOfFnc(handler ApiRouteHandler) (string, string, string, int) {
	fnc := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())

	fName, line := fnc.FileLine(0)
	fncName := fnc.Name()
	shortName := strings.TrimSuffix(path.Base(fncName), "-fm")

	packName, _, _ := strings.Cut(shortName, ".")
	return fmt.Sprintf("'%s'.%s", path.Dir(fncName), packName), fmt.Sprintf(`'%s(ctx)': (any, error)`,
		strings.ReplaceAll(strings.ReplaceAll(shortName, ".(*", "#"), ")", ""),
	), path.Base(fName), line
}

func (a *Apis) getDiagram(ctx *fasthttp.RequestCtx) (any, error) {
	buf := bytes.NewBufferString("direction: right\n")
	for address, route := range a.routes {
		name := strcase.ToCamel(strings.ReplaceAll(address.path, "/", "_"))
		if name == "" {
			name = "index"
		}
		name += address.method.String()
		entry := address.method.String()
		_, _ = fmt.Fprintf(buf, tplEndpoint, name, entry, address.path)
		params := make([]string, len(route.Params))
		for i, param := range route.Params {
			desc := param.Desc
			_, _ = fmt.Fprintf(buf, tplCol, param.Name, param.Type.String(), desc)
			if desc == "" {
				desc = "(not description)"
			}
			params[i] = fmt.Sprintf(tplParamDesc, param.Name, desc)
		}
		packName, fnc, fileName, line := getStringOfFnc(route.Fnc)
		if route.DTO != nil {
			entry += "(body)"
		} else {
			entry += "()"
		}
		_, _ = fmt.Fprintf(buf, fnc)
		_, _ = fmt.Fprintf(buf, tplEndpointEnd, strings.TrimPrefix(packName, "."), fileName, name, line)
		if len(params) > 0 {
			_, _ = fmt.Fprintf(buf, tplComment, name, route.Desc, `
---
Params:  
`+strings.Join(params, "\n"))
		} else {
			_, _ = fmt.Fprintf(buf, tplComment, name, route.Desc, "")
		}
	}
	s := buf.String()
	opts := &d2svg.RenderOpts{
		//Pad:           &d2svg.DEFAULT_PADDING,
		//Sketch:        true,
		//Center:        true,
		ThemeID:     &d2themescatalog.NeutralGrey.ID,
		DarkThemeID: nil,
		Font:        "",
		//SetDimensions: true,
		MasterID: "",
	}
	res, err := GetGraphSVG(ctx, buf, opts)
	if err != nil {
		errMsg := err.Error()
		b, _, ok := strings.Cut(errMsg, ":")
		if ok {
			i, err := strconv.Atoi(b)
			if err != nil {
				return nil, err
			}
			parts := strings.Split(s, "\n")
			if i < len(parts) {
				return nil, errors.New(parts[i-1] + errMsg)
			}
		}
		return nil, err
	}

	return res, nil
}
