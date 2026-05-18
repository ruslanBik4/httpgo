/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

package apis

import (
	"fmt"
	"go/types"
	"net/url"
	"regexp"
	"strings"
)

type CodeSample struct {
	Lang   string `json:"lang"`
	Label  string `json:"label,omitempty"`
	Source string `json:"source"`
}

func GenerateFrontendExamples(route *ApiRoute, path string) []CodeSample {
	method := strings.ToUpper(route.Method.String())

	query := url.Values{}
	form := url.Values{}

	var bodyFields []string

	for _, p := range route.Params {
		val := p.TestValue

		if val == "" {
			if p.DefValue != nil {
				val = fmt.Sprint(p.DefValue)
			} else {
				val = getExampleValue(p.Type)
			}
		}

		switch p.Type.(type) {
		case HeaderInParam:
			continue

		default:
			if route.Multipart {
				form.Set(p.Name, val)
			} else {
				query.Set(p.Name, val)
			}

			bodyFields = append(bodyFields,
				fmt.Sprintf("%q: %q", p.Name, val))
		}
	}

	fullURL := path

	if !route.Multipart && len(query) > 0 &&
		(method == "GET" || method == "DELETE") {
		fullURL += "?" + query.Encode()
	}

	var samples []CodeSample

	// FETCH
	{
		headers := []string{
			`"Content-Type": "application/json"`,
		}

		if route.IsAJAXRequest {
			headers = append(headers,
				`"X-Requested-With": "XMLHttpRequest"`)
		}

		var body string

		if method != "GET" && method != "DELETE" {
			body = fmt.Sprintf(`
	body: JSON.stringify({
		%s
	}),`, strings.Join(bodyFields, ",\n\t\t"))
		}

		code := fmt.Sprintf(`fetch("%s", {
	method: "%s",
	headers: {
		%s
	},%s
})
.then(r => r.json())
.then(console.log)`,
			fullURL,
			method,
			strings.Join(headers, ",\n\t\t"),
			body,
		)

		samples = append(samples, CodeSample{
			Lang:   "JavaScript",
			Label:  "fetch",
			Source: code,
		})
	}

	// HTMX
	{
		var attrs []string

		switch method {
		case "GET":
			attrs = append(attrs,
				fmt.Sprintf(`hx-get="%s"`, fullURL))
		case "POST":
			attrs = append(attrs,
				fmt.Sprintf(`hx-post="%s"`, path))
		case "PUT":
			attrs = append(attrs,
				fmt.Sprintf(`hx-put="%s"`, path))
		case "DELETE":
			attrs = append(attrs,
				fmt.Sprintf(`hx-delete="%s"`, path))
		}

		attrs = append(attrs,
			`hx-trigger="click"`)

		if route.IsAJAXRequest {
			attrs = append(attrs,
				`hx-headers='{"X-Requested-With":"XMLHttpRequest"}'`)
		}

		code := fmt.Sprintf(
			`<button %s>
	Load
</button>`,
			strings.Join(attrs, " "),
		)

		samples = append(samples, CodeSample{
			Lang:   "HTML",
			Label:  "HTMX",
			Source: code,
		})
	}

	// REACT
	{
		var body string

		if method != "GET" && method != "DELETE" {
			body = fmt.Sprintf(`
				body: JSON.stringify({
					%s
				}),`,
				strings.Join(bodyFields, ",\n\t\t\t\t\t"))
		}

		headers := []string{
			`"Content-Type": "application/json"`,
		}

		if route.IsAJAXRequest {
			headers = append(headers,
				`"X-Requested-With": "XMLHttpRequest"`)
		}

		componentName := "ApiExample"

		if route.Desc != "" {
			componentName = toReactComponentName(route.Desc)
		}

		trigger := ""

		if method == "GET" {
			trigger = fmt.Sprintf(`
	useEffect(() => {
		load();
	}, []);
`)
		}

		code := fmt.Sprintf(`import { useEffect, useState } from "react";

export default function %s() {
	const [data, setData] = useState(null);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState(null);

	async function load() {
		try {
			setLoading(true);

			const response = await fetch("%s", {
				method: "%s",
				headers: {
					%s
				},%s
			});

			if (!response.ok) {
				throw new Error("Request failed");
			}

			const json = await response.json();

			setData(json);
		} catch (e) {
			setError(e.message);
		} finally {
			setLoading(false);
		}
	}
%s
	if (loading) {
		return <div>Loading...</div>;
	}

	if (error) {
		return <div>Error: {error}</div>;
	}

	return (
		<div>
			<button onClick={load}>
				Load data
			</button>

			<pre>
				{JSON.stringify(data, null, 2)}
			</pre>
		</div>
	);
}`,
			componentName,
			fullURL,
			method,
			strings.Join(headers, ",\n\t\t\t\t\t"),
			body,
			trigger,
		)

		samples = append(samples, CodeSample{
			Lang:   "React",
			Label:  "React",
			Source: code,
		})
	}
	// EventSource
	if route.IsServerEvents {
		code := fmt.Sprintf(`const evt = new EventSource("%s");

evt.onmessage = (e) => {
	console.log(e.data);
};

evt.onerror = (e) => {
	console.error(e);
};`, fullURL)

		samples = append(samples, CodeSample{
			Lang:   "JavaScript",
			Label:  "EventSource",
			Source: code,
		})
	}

	// CURL
	{
		curl := fmt.Sprintf(`curl -X %s "%s"`,
			method,
			fullURL)

		samples = append(samples, CodeSample{
			Lang:   "bash",
			Label:  "curl",
			Source: curl,
		})
	}

	return samples
}

func getExampleValue(t any) string {
	switch v := t.(type) {

	case TypeInParam:
		switch v.BasicKind {
		case types.Int,
			types.Int8,
			types.Int16,
			types.Int32,
			types.Int64:
			return "123"

		case types.Float32,
			types.Float64:
			return "123.45"

		case types.Bool:
			return "true"

		default:
			return "example"
		}
	}

	return "example"
}

func toReactComponentName(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return "ApiExample"
	}

	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	parts := reg.Split(s, -1)

	var out string

	for _, p := range parts {
		if p == "" {
			continue
		}

		out += strings.ToUpper(p[:1]) +
			strings.ToLower(p[1:])
	}

	if out == "" {
		return "ApiExample"
	}

	return out
}
