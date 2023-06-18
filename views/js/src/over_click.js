/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
function OverClick() {
    var $out = $('#content');
    let url = replMacros(this.href);
    let target = this.target;
    $.ajax({
        url: url,
        data: {
            "lang": lang,
            "html": true
        },
        processData: false,
        contentType: false,
        beforeSend: getHeaders,
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

                var blob = new Blob([data], {type: typeCnt});
                var a = document.createElement('a');
                a.href = window.URL.createObjectURL(blob);
                a.download = disp.split("=")[1];
                a.rel = "tmp";
                console.log(a);
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);

                return;
            } else if (typeCnt.startsWith("text/css")) {
                var id = xhr.getResponseHeader('Section-Name');
                LoadStyles(id, data)
                return;
            } else if (typeCnt.startsWith("application/json")) {
                $('#content').html(JSON.stringify(data))
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
            console.log(xhr);
        }
    });
    return false;
}

//replace special symbols
function replMacros(url) {
    return url.replace(/{page}/, GetPageLines())
}

function getHeaders(xhr) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + token);
    xhr.setRequestHeader('Accept-Language', lang);
}