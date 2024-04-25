/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
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
        PutContent(`<iframe src='${url}?embedded=true' allowtransparency seamless></iframe>`, 'iframe');
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
                    OverHijack($out, data);
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
            if (xhr.status === 401) {
                urlAfterLogin = url;
                $('#bLogin').trigger("click");
                return;
            }

            alert(`Code : "${xhr.status}", "${error}": "${xhr.responseText}"`);
            console.error(`Code : "${xhr.status}", "${error}": "${xhr.responseText}"`, xhr);
        }
    });
    return false;
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
                        div.prepend(`<h3> ${elem[y]}</h3>`);
                        break;
                    }
                    case "id":
                        div.attr('id', elem[y].id);
                        break;
                    default:
                        if (Array.isArray(elem[y])) {
                            div.append(`<h4>${y}</h4>`);
                            for (x in elem[y]) {
                                showJsonElem(elem[y][x]);
                                div.append("<br>")
                            }
                        } else if (elem[y] instanceof Object) {
                            div.append(`<h4>${y}</h4>`);
                            showJsonElem(elem[y]);
                        } else {
                            div.append(`<p><i>${y}:</i> <span>${elem[y]}</span></p>`);
                        }
                }
            }
        } else {
            divContent.append(`<span>${elem}</span>`);
        }

    }

    showJsonElem(data);
}
//replace special symbols
function replMacros(url) {
    return url.replace(/{page}/, GetPageLines())
}

function getHeaders(xhr) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + token);
    xhr.setRequestHeader('Accept-Language', lang);
}