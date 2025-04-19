/*
 * Copyright (c) 2023-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

function ClickPseudo(elem) {
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

}

const selTablesRows = '.usr-table-row-cont';

// reorder data and get new table content
function loadTableWithOrder() {
    let orderBy = $('.usr-table__t-head .usr-table-col')
        .children('span.sorted-asc,span.sorted-desc')
        .map(function () {
            return $(this).attr('column') + (this.className === 'sorted-desc' ? ' desc' : '');
        }).get().join(",");

    let url = new URL(window.location.href);
    let params = new URLSearchParams(url.search);

    if (orderBy === params.get("order_by")) {
        return false;
    }

    params.set("order_by", orderBy);

    $.ajaxSetup({
        beforeSend: getHeaders,
    });
    let newURL = url.origin + url.pathname + "?" + params.toString();
    // load only table rows content
    $(selTablesRows).load(newURL + ' .usr-table-row-cont');

    return setHashFromTable(newURL)
}

// set hash with all data of table
function setHashFromTable(url) {
    SetDocumentHash(url, $('.usr-table').html());
    return true;
}

function reqParams() {
    let url = new URL(document.location.href);
    let params = setParams(url);
    params.set('limit', GetPageLines() * 2);
    console.log(params.toString());
    return Object.fromEntries(params.entries());
}

function setParams(url) {
    var lines = $(selTablesRows).children('div:not(.'+unfiltered+')').length;
    let params = new URLSearchParams(url.search);

    params.set("offset", `${lines}`);
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
                params.set(elem.dataset.name, value);
            }
        });

    let orderBy = $('.usr-table__t-head .usr-table-col')
        .children('span.sorted-asc,span.sorted-desc')
        .map(function () {
            return $(this).attr('column') + (this.className === 'sorted-desc' ? ' desc' : '');
        }).get().join(",");

    // when order_by changed request first data
    if (params.get("order_by") !== orderBy) {
        params.set("order_by", orderBy);
        params.set("offset", 0);
    }

    return params;
}

function chkConditions(href) {
    let url = new URL(href);
    let params = setParams(url);

    return url.origin + url.pathname + "?" + params.toString();
}

// append data over limit into table content
function appendTable() {
    let tableCnt = $('.usr-table-content');
    var tableRows = $(selTablesRows);
    let newURL = chkConditions(window.location.href);
    $.ajax({
        url: newURL,
        data: {
            "html": true
        },
        processData: false,
        contentType: false,
        beforeSend: getHeaders,
        success: function (data, status, xhr) {
            if (xhr.status === 204) {
                tableCnt.off('mousewheel');
                tableRows.html(data);
                return false;
            }
            tableRows.append($('<div />').html(data).find(selTablesRows).html());
            setHashFromTable(newURL);
        },
        error: function (xhr, status, error) {
            if (xhr.status === 401) {
                urlAfterLogin = newURL;
                $('#bLogin').trigger("click");
                return;
            }

            alert("Code : " + xhr.status + ", " + error + ": " + xhr.responseText);
            console.log(xhr);
        }
    });
    return true;
}
const unfiltered = 'unfiltered';

function filterTableData(value, className) {
    let val = value.trim();
    if (val === "") {
        $('.usr-table-row-cont .usr-table-row').show();
        return true;
    }

    let dateRanges = val.match(/\[(\d+-\d+-\d+),(\d+-\d+-\d+)\]/);
    if (!dateRanges) {
        var numberRanges = val.match(/[[)](\d+.?\d*),(\d+.?\d*)[\])]/);
    }
    if (!dateRanges && !numberRanges) {
        var strSlices = val.match(/(\w),(\w+)/);
        console.log(strSlices);
    }
    let i = 0;
    var items = document.getElementsByClassName(className);
    // filter rows according to input text/value
    Array.prototype.slice.call(items).forEach(
        (el, ind, arr) => {
            let textContent = el.textContent;
            if (textContent.includes(val)
                || (dateRanges && (dateRanges.length > 1) && textContent >= dateRanges[1] && textContent <= dateRanges[2])
                || (numberRanges && (numberRanges.length > 1) && textContent >= numberRanges[1] && textContent <= numberRanges[2])
                || (strSlices && (strSlices.length > 1) && textContent >= strSlices[1] && textContent <= strSlices[2])
                || el.parentElement.className.includes("usr-table__t-head")
                || el.parentElement.className.includes("usr-table__filter")) {
                $(el).removeClass(unfiltered);
                i++
                return true;
            }
            $(el).addClass(unfiltered);
            return true;
        });

    $('.usr-table-row').each((ind, elem) => {
        if ($(elem).children('.usr-table-col.'+unfiltered).length > 0) {
            $(elem).hide();
        } else {
            $(elem).show();
        }
    })
    // append data if we filter has counter of lines less than page
    if (i < GetPageLines()) {
        appendTable();
    }
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
    // $('.usr-table__t-head .usr-table-col span').click(ClickPseudo);
    setSortedClasses();

}

function setSortedClasses() {
    let url = new URL(window.location.href);
    let params = new URLSearchParams(url.search).get("order_by");

    console.log(params);
    if (params) {
        for (let name of params.split(',')) {
            console.log(name);
            let sortedClass = 'sorted-asc';
            if (name.toString().endsWith('desc')) {
                sortedClass = 'sorted-desc';
                name = name.toString().slice(0,-5)
            }
            $(`.usr-table__t-head .usr-table-col span[column="${name}"]`).addClass(sortedClass);
        }
    }
}

function HideColumn(num, chk) {
    if (chk) {
        $(`.table-col-${num}`).show();
    } else {
        $(`.table-col-${num}`).hide();
    }
}

function HideAllColumn(elem) {
    const chkColumns = 'input[type="checkbox"][data-role="chk_column"]';
    $(chkColumns).click(); //elem.checked);
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
            let fText = [];
            for (let elem of csv) {
                let row = '<div  class="usr-table-col  table-col-0">$</div>';
                if (elem.length === 1) {
                    elem = elem[0].split(';');
                }
                elem.forEach(function (cell, i) {
                    row += `<div  class="usr-table-col  table-col-${i}">${cell}</div>`;
                });
                fText.push(`<div  class="usr-table-row">${row}</div>`);
                if (fText.length >= GetPageLines()) {
                    console.log(row);
                    break;
                }
            }

            $(selTablesRows).html(fText.join('\n'));
        };
    })(f);

    // Read in the image file as a data URL.
    reader.readAsText(f);
}
