/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */

"use strict";

function fancyOpen(data) {
    $.fancybox.open({
        'autoScale': true,
        'transitionIn': 'elastic',
        'transitionOut': 'elastic',
        'speedIn': 500,
        'speedOut': 300,
        'type': 'html',
        'autoDimensions': true,
        'centerOnScroll': true,
        'content': data
    })
}

//replace special symbols
function replMacros(url) {
    return url.replace(/{page}/, GetPageLines())
}

// get lines for table according to windows height
function GetPageLines() {
    return Math.round((window.innerHeight - 60) / 22)
}

function LoadStyles(id, styles) {
    let $head = $('head > style#' + id);
    if ($head.length === 0) {
        $('head').append('<style title="themes" id="' + id + '">' + styles + '</style>');
    } else {
        $head.html(styles);
    }
}

function AddStyles(name, css) {
    $('head').append(`<style title="${name}">${css}</style>`);
}

function AddScript(name, js) {
    $('head').append(`<script title="${name}">${js}</script>`);
}

var go_history = 1;

// Эта функция отрабатывает при перемещении по истории просмотром (кнопки вперед-назад в браузере)
function MyPopState(event) {
    if ((go_history === 0) || (event.state == null))
        return true;
    console.log(event);
    document.title = event.title;
    SetContent(event.state.data);
}

// смена адресной строки с предотвращением перезагрузки Содержимого
function SetDocumentHash(url, data) {
    // let root_page = "/";
    // let default_page = "index.html";
    document.title = (typeof data == 'string' && data.search('<title') > -1 && $(data).filter('title').text()) || GetShortURL(url);

    // var origin = document.location.origin + (str_path[0] === '/' ? '' : "/")
    //     + ((str_path !== root_page) && (str_path !== default_page) ? str_path : '');
    console.log(url)
    window.history.pushState({'url': url, 'data': data}, document.title, url);
}

function GetShortURL(url) {
    if (url === "") {
        return '/';
    }

    if (url.startsWith(document.location.origin)) {
        return url.slice(document.location.origin.length + 1);
    }
    return url
}


$(function () {
    if (!window.onpopstate) {
        window.onpopstate = MyPopState;
    }

    window.addEventListener("beforeunload", function (evt) {
        evt = evt || window.event;

        if (evt) {
            var y = evt.pageY || evt.clientY;
            if (y === undefined) {
                console.log(evt)
            }
            console.log(`beforeunload ${document.location} pageY:${y}`);
            evt.preventDefault();
            if (y < 0) {
                return evt.returnValue = "Do you want to close this page?"
            }

            if (document.location.pathname > "/") {
                let url = document.location
                document.location.href = document.location.origin;
//					loadContent(url.toString());
//					url.pathname = "/";
                console.log(`reload ${url}`)
                evt.target.URL = url.origin;
                evt.srcElement.URL = evt.target.URL;
                console.log(evt)
            }
        }
        return false
    })

    if (!userStruct) {
        userStruct = getUser();
    }

    const rel = $('#content').attr('rel');
    if (rel > "") {
        loadContent(rel)
    }
    setClickAll();
    $('body').on('DOMSubtreeModified', setClickAll);


}) // $(document).ready

// run request & show content
function loadContent(url) {

    $.ajax({
        url: url,
        data: {
            "lang": lang,
            "html": true
        },
        beforeSend: getHeaders,
        success: function (data, status) {
            PutContent(data, url);
        },
        error: function (xhr, status, error) {
            switch (xhr.status) {
                case 401:
                    urlAfterLogin = url;
                    $('#bLogin').trigger("click");
                    return;
                case 404:
                    alert(`address '${url}' not found!`)
                    return;
                case 0:
                    console.log(xhr);
            }

            alert("Code : " + xhr.status + " error :" + error);
            console.log(`${url} ${status} ${error}`);
        }
    });
}

function PutContent(data, url) {
    SetContent(data);
    SetDocumentHash(url, data);
}

function SetContent(data) {
    if (typeof data == 'string' && data.startsWith('<!DOCTYPE html>')) {
        $('html').html(data);
        return
    }
    const a = document.createElement('div');

    a.innerHTML = data;

    // sidebar work only for own page
    $('#catalog_pane .sidebar').remove();
    $('.sidebar', a).appendTo('#catalog_pane');
    findAndReplaceElem(a, '.sidebar-section', 'main .sidebar-section')
    findAndReplaceElem(a, 'header .topline', 'body > header .topline');
    findAndReplaceElem(a, 'header .topline-btns', 'body > header .topline-btns');
    if (!findAndReplaceElem(a, '#content', '#content')) {
        $('#content').html(a.innerHTML);
    }

}

function findAndReplaceElem(src, selector, dst) {
    const elem = $(selector, src);
    if (elem.length > 0) {
        $(dst).html(elem.html());
        return true;
    }
    return false;
}

function LoadJScript(url, asyncS, cacheS, successFunc, completeFunc) {
    $.ajax({
        type: "GET",
        async: asyncS,
        cache: cacheS,
        url: url,
        global: false,
        dataType: "script",
        success: successFunc,
        complete: completeFunc,
        error: errorLoadResource
    });
}

function LoadCSS(url, cacheS, successFunc) {
    $.ajax({
        type: "GET",
        cache: cacheS,
        url: url,
        beforeSend: getHeaders,
        global: false,
        dataType: "text",
        success: successFunc,
        error: errorLoadResource
    })
}

function errorLoadResource(xhr, textStatus, errorThrown) {
    if (errorThrown !== undefined) {
        console.error(`%s from '${xhr}'! (${textStatus}). Pls, reload page!`, errorThrown);
    } else {
        console.error(`Can't load resource from '${xhr}'! (${textStatus}). Pls, reload page! %s`, textStatus);
    }
}

function isScrollableY(node) {
    const overflowY = window.getComputedStyle(node)['overflow-y'];
    return (overflowY === 'scroll' || overflowY === 'auto');
}

function isScrollableX(node) {
    const overflowX = window.getComputedStyle(node)['overflow-x'];
    return (overflowX === 'scroll' || overflowX === 'auto');
}