/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
function OverClick() {
    var $out = $('#content');
    $.ajax({
        url: url,
        data: {
            "lang": lang,
            "html": true
        },
        processData: false,
        contentType: false,
        beforeSend: function (xhr) {
            xhr.setRequestHeader('Authorization', 'Bearer ' + token);
            xhr.setRequestHeader('Accept-Language', lang);
        },
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
            if (disp && disp.search('attachment') !== -1) {

                var blob = new Blob([data], {type: typeCnt});
                var url = window.URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = 'file.pdf';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                window.URL.revokeObjectURL(url);

                data = `<img src='data:${data}' alt='${url}'/>`;
            } else if (typeCnt.startsWith("text/css")) {
                var id = xhr.getResponseHeader('Section-Name');
                LoadStyles(id, data)
                return;
            } else if (typeCnt.startsWith("application/json")) {
                $('#content').html(JSON.stringify(data))
                return;
            }
            if (target !== "_modal") {
                PutContent(data, url);
            } else {
                fancyOpen(data);
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
