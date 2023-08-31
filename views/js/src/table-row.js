/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

function ClickPseudo(event) {
    var elem = event.target;
    var offset = event.clientX || event.originalEvent.clientX;
    console.log(event);
    console.log(`${offset} > ${elem.offsetLeft}, ${elem.getBoundingClientRect().left}`);
    // clear other sorted
    if (!event.altKey) {
        $('.sorted-desc').removeClass('sorted-desc');
        $('.sorted-asc').removeClass('sorted-asc');
    }

    if (offset > elem.getBoundingClientRect().left + 20) {
        $(elem).removeClass('sorted-desc').addClass('sorted-asc');
    } else {
        $(elem).removeClass('sorted-asc').addClass('sorted-desc');
    }

    loadTableWithOrder();
}

const reqOffset = /(order_by=)(\w+(%20desc)?)/;

function getOrderByStatus(url) {
    return reqOffset.exec(url)
}

// reorder data & get new table content
function loadTableWithOrder() {
    let orderBy = $('.usr-table__t-head .usr-table-col')
        .children('span.sorted-asc,span.sorted-desc')
        .map(function () {
            return $(this).attr('column') + (this.className === 'sorted-desc' ? '%20desc' : '');
        }).get().join(",");

    console.log(orderBy)

    var url = document.location.href;
    const parts = getOrderByStatus(url);

    if (!parts || parts.length <= 0) {
        url += (document.location.search > "" ? '&' : '?') + `order_by=${orderBy}`
    } else if (orderBy === parts[2]) {
        console.log(parts)
        return false;
    } else {
        url = url.replace(reqOffset, `$1${orderBy}`);
        console.log(url)
    }

    $.ajaxSetup({
        beforeSend: getHeaders,
    });
    // load only table rows content
    $('.usr-table-row-cont').load(url + ' .usr-table-row-cont');

    return setHashFromTable(url)
}

// set hash with all data of table
function setHashFromTable(url) {
    SetDocumentHash(url, $('.usr-table').html());
    return true;
}

// append data over limit into table content
function appendTable() {
    var elem = $('.usr-table-row-cont');
    var lines = elem.children('div:visible').length;
    let url = document.location.href;
    if (url.indexOf('offset') === -1) {
        url += (document.location.search > "" ? '&' : '?') + `offset=${lines}`
    } else {
        const reqOffset = /(offset=)(\d+)/.exec(url);
        if (lines == integer(reqOffset[1])) {
            return false;
        }
        url = url.replace(reqOffset, `$1${lines}`);
    }
    $('div.filt-arrow > input, div.filt-arrow > select').each(
        (i, elem) => {
            if (elem.value) {
                let value = elem.value;
                if (elem.type === 'checkbox') {
                    if (!elem.checked) {
                        return
                    }
                    value = true
                }
                if (url.indexOf(`${elem.dataset.name}=`) === -1) {
                    url += `&${elem.dataset.name}=${value}`;
                } else {
                    var r = new RegExp(`(${elem.dataset.name}=)(\[^&]+)`);
                    url = url.replace(r, `$1${value}`);
                }
            }
        });
    console.log(url);
    $.ajax({
        url: url,
        data: {
            "html": true
        },
        processData: false,
        contentType: false,
        beforeSend: getHeaders,
        success: function (data, status, xhr) {
            if (xhr.status === 204) {
                return false;
            }
            elem.append($('<div />').html(data).find('.usr-table-row-cont').html());
            setHashFromTable(url);
        },
        error: function (xhr, status, error) {
            if (xhr.status === 401) {
                urlAfterLogin = url;
                $('#bLogin').trigger("click");
                return;
            }

            alert("Code : " + xhr.status + ", " + error + ": " + xhr.responseText);
            console.log(xhr);
        }
    });
    return true;
}

function filterTableData(value, className) {
    if (value.trim() === "") {
        $('.usr-table-row-cont .usr-table-row').show();
        return true;
    }

    let i = 0;
    var items = document.getElementsByClassName(className);
    // filter rows according to input text/value
    Array.prototype.slice.call(items).forEach(
        (el, ind, arr) => {
            if (el.textContent.includes(value.trim())
                || el.parentElement.className.includes("usr-table__t-head")
                || el.parentElement.className.includes("usr-table__filter")) {
                el.parentElement.style = "";
                i++
                return true;
            }
            el.parentElement.style = "display:none";
            return true;
        });
    // append data if we filter has counter of lines less than page
    if (i < GetPageLines()) {
        appendTable();
    }

//  if (elem.length > 0) {
////todo- chg on each
//    elem[0].scrollIntoView({block: "center", behavior: "smooth"});
//    elem[0].focus();
//    elem[0].animate([
//      {color: 'blue'},
//      {color: 'red'}
//    ], {
//        duration: 3000,
//        iterations: 100
//    });
//  }
}

function ScrollToElem(selector) {
    var list = $(selector);
    if (list.length > 0) {
        list[0].scrollIntoView(100);
    } else {
        alert(selector + ' not found!');
    }
    return true;
}

function SetTableEvents() {
    var throttleTimer;
    const throttle = (callback, time) => {
        if (throttleTimer) return;
        throttleTimer = true;
        callback();
        setTimeout(() => {
            throttleTimer = false;
        }, time);

    };
    $('.usr-table__t-head .usr-table-col span').click(ClickPseudo);
    let tableCnt = $('.usr-table-content');
    tableCnt.off('mousewheel');
    tableCnt.on('mousewheel', function (event, delta) {
        var elem = event.target;
        console.log(event)
        if (elem.clientHeight < elem.scrollWidth) {
            console.log(`${elem.clientHeight} < ${elem.scrollWidth}`);
            return true;
        }

        if ((event.deltaY < 0) && tableCnt.scrollTop() + tableCnt.height() > Math.ceil(tableCnt[0].scrollHeight / 2)) {
            console.log(elem);
            console.log(`${tableCnt.scrollTop() + tableCnt.height()} ${Math.ceil(tableCnt[0].scrollHeight / 2)}`);
            throttle(appendTable, 300);
            return true;
        }
        return true;
    })

    const reqOffset = getOrderByStatus(document.location.href);
    if (reqOffset && reqOffset.length > 1) {
        const colName = reqOffset[2].split('%20');
        console.log(colName);
        $(`.usr-table__t-head .usr-table-col:nth-child(n+2)[column=${colName[1]}]`).addClass('sorted-asc');
    }
}


function handleFileCSVSelect(evt) {
    var files = evt.files || evt.target.files; // FileList object
    if (files.length < 1)
        return false;

    let $progress = $('#progress').show(),
        reader = new FileReader(),
        f = files[0];

    reader.onload = (function (theFile) {
        return function (e) {
            let csv = e.target.result.csvToArray({head: true, rSep: "\n"});
            let fText = '';
            csv.forEach(function (elem) {
                let row = '';
                elem.forEach(function (cell, i) {
                    row += `<div  class="usr-table-col  table-col-${i}">${cell}</div>`;
                });
                console.log(row);
                fText += `<div  class="usr-table-row">${row}</div>`;
            });
            $('.usr-table-row-cont').html(fText);
        };
    })(f);

    // Read in the image file as a data URL.
    reader.readAsText(f);
}
