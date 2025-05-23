// Code generated by qtc from "over_click.js.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line ../over_click.js.qtpl:1
package js

//line ../over_click.js.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line ../over_click.js.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line ../over_click.js.qtpl:1
func StreamOverClick(qw422016 *qt422016.Writer) {
//line ../over_click.js.qtpl:1
	qw422016.N().S(`	`)
//line ../over_click.js.qtpl:2
	qw422016.N().S(`/*
 * Copyright (c) 2023-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

function OverClick() {
    var $out = $('#content');
    let url = replMacros(this.href);
    let target = this.target;
    if (target === "_iframe") {
        PutContent(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<iframe src='${url}?embedded=true' allowtransparency seamless></iframe>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`, 'iframe');
        return false;
    }

    $.ajax({
        url: url,
        data: {
            "lang": lang,
            "html": true
        },
        converters: {
            "* text": window.String,
            "text html": true,
            "text json": jQuery.parseJSON,
            "text xml": jQuery.parseXML,
            'arraybuffer': jQuery.arraybuffer,
            'binary blob': function (data, tyype) {
                console.log(tyype);
                console.log(typeof data);
                return data;
            },
            'blob binary': function (data, tyype) {
                console.log(tyype);
                console.log(typeof data);
                return data;
            },
            'text binary': function (data, tyype) {
                console.log(tyype);
                console.log(typeof data);
                return data;
            },
        },
        processData: false,
        contentType: false,
        dataFilter: function (data, type) {
            console.log(type);
            return data
        },
        beforeSend: getHeaders,
        xhr: function () {
            var xhr = new XMLHttpRequest();
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 2) {
                    if (xhr.status === 200) {
                        var disp = xhr.getResponseHeader('Content-Disposition');
                        if (disp && disp.startsWith('attachment')) {
                            xhr.responseType = "blob";
                        }
                    }
                }
            };
            return xhr;
        },
        // responseType: 'binary, blob, text, html, xml, json',
        success: function (data, status, xhr) {
            switch (xhr.status) {
                case 204: {
                    alert("no content!" + status)
                    return
                }
                case 206: {
                    readEvents($out, data);
                    return
                }
            }

            var disp = xhr.getResponseHeader('Content-Disposition');
            var typeCnt = xhr.getResponseHeader('Content-Type');
            if (disp && disp.startsWith('attachment')) {
                // todo: add last modify
                console.log(typeCnt);
                const fileName = disp.split("=")[1];
                const blob = new File([data], fileName, {
                    type: typeCnt,
                    lastModified: xhr.getResponseHeader('Last-Modified')
                });
                console.log(blob);
                const a = document.createElement('a');
                a.href = window.URL.createObjectURL(blob);
                a.download = fileName;
                a.rel = "tmp";
                document.body.appendChild(a);
                // blob.lastModified;
                a.click();
                setTimeout(() => {
                    document.body.removeChild(a);
                    window.URL.revokeObjectURL(url);
                }, 100);

                return;
            } else if (typeCnt.startsWith("text/css")) {
                const id = xhr.getResponseHeader('Section-Name');
                LoadStyles(id, data)
                return;
            } else if (typeCnt.startsWith("application/json")) {
                showJSON(data);
                return;
            }
            if (target === "_modal") {
                fancyOpen(data);
            } else {
                PutContent(data, url);
            }
        },
        error: function (xhr, status, error) {
            xhr.url = url;
            return handleError(xhr, status, error);
        },
    });
    return false;
}

function handleError(xhr, status, error) {
    switch (xhr.status) {
        case 401:
            urlAfterLogin = xhr.url;
            console.log(xhr.url);
            $('#bLogin').trigger("click");
            return;
        case 404:
            alert(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`Request page not found: ${xhr.url}`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
            return;
    }

    alert(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`Code : "${xhr.status}", "${error}": "${xhr.responseText}"`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
    console.error(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`Code : "${xhr.status}", "${error}": "${xhr.responseText}"`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`, xhr);
}

function showJSON(data) {
    if (!data) {
        alert('no results!')
        return false;
    }

    let divContent = $('#content').html('');

    function showJsonElem(elem) {
        let x, y;
        if (Array.isArray(elem)) {
            for (x in elem) {
                showJsonElem(elem[x]);
            }
        } else if (elem instanceof Object) {
            let div = divContent.append('<div>');
            for (y in elem) {
                switch (y) {
                    case "name":
                    case "full_name": {
                        div.prepend(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<h3> ${elem[y]}</h3>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
                        break;
                    }
                    case "id":
                        div.attr('id', elem[y].id);
                        break;
                    default:
                        if (Array.isArray(elem[y])) {
                            div.append(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<h4>${y}</h4>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
                            for (x in elem[y]) {
                                showJsonElem(elem[y][x]);
                                div.append("<br>")
                            }
                        } else if (elem[y] instanceof Object) {
                            div.append(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<h4>${y}</h4>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
                            showJsonElem(elem[y]);
                        } else {
                            div.append(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<p><i>${y}:</i> <span>${elem[y]}</span></p>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
                        }
                }
            }
        } else {
            divContent.append(`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`<span>${elem}</span>`)
//line ../over_click.js.qtpl:2
	qw422016.N().S("`")
//line ../over_click.js.qtpl:2
	qw422016.N().S(`);
        }

    }

    showJsonElem(data);
}

// //replace special symbols
// function replMacros(url) {
//     return url.replace(/{page}/, GetPageLines())
// }

function getHeaders(xhr) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + token);
    xhr.setRequestHeader('Accept-Language', lang);
    xhr.setRequestHeader('Content-Encoding', 'gzip, deflate');
}`)
//line ../over_click.js.qtpl:2
	qw422016.N().S(`
`)
//line ../over_click.js.qtpl:3
}

//line ../over_click.js.qtpl:3
func WriteOverClick(qq422016 qtio422016.Writer) {
//line ../over_click.js.qtpl:3
	qw422016 := qt422016.AcquireWriter(qq422016)
//line ../over_click.js.qtpl:3
	StreamOverClick(qw422016)
//line ../over_click.js.qtpl:3
	qt422016.ReleaseWriter(qw422016)
//line ../over_click.js.qtpl:3
}

//line ../over_click.js.qtpl:3
func OverClick() string {
//line ../over_click.js.qtpl:3
	qb422016 := qt422016.AcquireByteBuffer()
//line ../over_click.js.qtpl:3
	WriteOverClick(qb422016)
//line ../over_click.js.qtpl:3
	qs422016 := string(qb422016.B)
//line ../over_click.js.qtpl:3
	qt422016.ReleaseByteBuffer(qb422016)
//line ../over_click.js.qtpl:3
	return qs422016
//line ../over_click.js.qtpl:3
}
