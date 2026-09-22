/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

"use strict";

// App is the page-lifecycle/content-loading singleton: one instance, built
// once at parse time, holding what used to be a dozen unrelated top-level
// functions and a bare `go_history` global. Everything below the IIFE is a
// same-named global wrapper (e.g. `function loadContent(url){ return
// App.loadContent(url); }`) so every existing caller - forms.js, observer.js,
// set_clicks.js, setFlatPickr.js, table-row.js, and any onclick="..."/
// hx-on="..." attribute already baked into rendered HTML - keeps working
// unchanged. New code should call App.* directly; the wrappers exist for
// everything this refactor can't also rewrite. over_click.js is gone -
// handleError()/getHeaders() moved into observer.js/user.js respectively.
//
// De-duplicated against htmx's own controls (see onPopState/setDocumentHash
// below for the history fix, and Auth.headers() in user.js for the auth/
// language fix - every request in the app, htmx or not, gets those two
// headers from that one place now): this file no longer tracks its own full
// copy of "what does document.title get set to" or "what headers go on this
// request" - those were previously computed twice by two different,
// occasionally-disagreeing implementations. loadContent()/loadCSS() below
// now issue their requests via htmx.ajax() rather than $.ajax(), so a 401
// during either one is handled by the same Auth.relogin(evt) path as every
// other request instead of its own separate case statement.
const App = (function () {

    let goHistory = 1;

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
        });
        setTimeout(function () {
            processAll('.fancybox-inner');
        }, 1000);
    }

    // replace special symbols
    function replMacros(url) {
        return url.replace(/{page}/, getPageLines() * 2)
    }

    // get lines for table according to windows height
    function getPageLines() {
        return Math.round((window.innerHeight - 60) / 22)
    }

    function loadStyles(id, styles) {
        let $head = $('head > style#' + id);
        if ($head.length === 0) {
            $('head').append('<style title="themes" id="' + id + '">' + styles + '</style>');
        } else {
            $head.html(styles);
        }
    }

    function loadCSSOnce(url) {
        if (document.querySelector(`link[href="${url}"]`)) return;

        const link = document.createElement("link");
        link.rel = "stylesheet";
        link.href = url;
        document.head.appendChild(link);
    }

    function addStyles(name, css) {
        $('head').append(`<style title="${name}">${css}</style>`);
    }

    function addScript(name, js) {
        $('head').append(`<script title="${name}">${js}</script>`);
    }

    // Эта функция отрабатывает при перемещении по истории просмотром (кнопки вперед-назад в браузере)
    //
    // <body hx-boost="true"> means htmx owns popstate for every boosted
    // click/submit already - it keeps its own history-cache entries and
    // restores them itself, with no help from this file. This handler only
    // exists for the few navigations that DON'T go through boost - the ones
    // App.loadContent() drives directly (changeLang, logOut, the post-login
    // redirect) - which push their own {url, data} state below. The
    // `event.state.data === undefined` guard is what keeps this from
    // misfiring on an htmx-owned history entry: those have a different
    // shape, so this correctly no-ops and lets htmx's own listener handle
    // it instead of calling setContent(undefined).
    function onPopState(event) {
        if (goHistory === 0 || event.state == null || event.state.data === undefined) {
            return true;
        }

        const title = setContent(event.state.data);
        if (title > "") {
            document.title = title;
        }
    }

    // смена адресной строки с предотвращением перезагрузки Содержимого
    //
    // Only pushes the history entry - title is set by putContent() (via
    // setContent()'s own, more complete `title, h2` extraction) right after
    // this is called, so duplicating that extraction here would just risk
    // setting document.title twice with two different results.
    function setDocumentHash(url, data) {
        window.history.pushState({'url': url, 'data': data}, '', url);
    }

    // run request & show content
    //
    // htmx.ajax(), not $.ajax(): this is now the SAME request pipeline every
    // hx-* element uses, so it gets Accept-Language/Authorization from
    // Auth.headers() via observer.js's htmx:configRequest, and a 401/404/etc.
    // goes through the shared htmx:responseError handler (which calls
    // Auth.relogin(evt) on 401) instead of its own separate error callback.
    // swap:'none' + a custom handler because the response isn't a single
    // innerHTML swap into one target - setContent()/putContent() below
    // distribute pieces of it across #content, the sidebar, breadcrumbs and
    // the topline, which nothing in htmx's own swap mechanism knows how to
    // do. That's also why this still drives its own history via
    // setDocumentHash()/onPopState() rather than htmx's push/hx-history-elt
    // snapshotting - htmx can only snapshot+restore a single element, not
    // this multi-region layout.
    function loadContent(url) {
        htmx.ajax('GET', url, {
            values: {html: true},
            swap: 'none',
            handler: (elt, info) => putContent(info.xhr.responseText, url)
        });
    }

    function putContent(data, url) {
        const title = setContent(data);
        const isChild = url && url.startsWith(document.location.href);

        setDocumentHash(url, data);
        if (title > "") {
            if (isChild) {
                $('ol.breadcrumb').append(`<li class="breadcrumb-item">${title}</li>`);
            } else {
                $('ol.breadcrumb li:last').text(title);
            }
            document.title = title;
        }
    }

    function setContent(data) {
        if (typeof data == 'string' && data.startsWith('<!DOCTYPE html>')) {
            $('html').html(data);
            return
        }
        const a = document.createElement('div');

        a.innerHTML = data;

        // sidebar work only for own page
        $('#catalog_pane .sidebar').remove();
        $('.sidebar', a).appendTo('#catalog_pane');
        findAndReplaceElem(a, 'breadcrumbs', 'breadcrumbs');
        findAndReplaceElem(a, '.sidebar-section', 'main .sidebar-section');
        findAndReplaceElem(a, 'header .topline', 'body > header .topline');
        findAndReplaceElem(a, 'header .topline-btns', 'body > header .topline-btns');
        const $content = $('#content');
        if (!findAndReplaceElem(a, '#content', '#content')) {
            $content.html(a.innerHTML).removeAttr('rel');
            setClickAll($content[0]);
        }
        return $('title, h2', a).text()
    }

    function findAndReplaceElem(src, selector, dst) {
        const elem = $(selector, src);
        if (elem.length > 0) {
            $(dst).html(elem.html()).removeAttr('rel');
            setClickAll($(dst)[0]);
            return true;
        }
        return false;
    }

    function loadJScript(url, asyncS, cacheS, successFunc, completeFunc) {
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

    // `cacheS` is kept only for call-signature compatibility (changeTheme()
    // in user.js still passes it) - htmx.ajax has no cache-busting option to
    // forward it to, and `true` (the only value ever passed) is the normal
    // browser-cache behavior anyway. Errors fall back to the shared
    // htmx:responseError handling (observer.js) instead of errorLoadResource
    // - htmx.ajax's `handler` option only runs on success, so a per-call
    // error callback isn't available here any more; see loadJScript() below
    // for the one remaining $.ajax call that still uses errorLoadResource
    // directly (loading a <script>, not data - out of scope for this
    // htmx conversion).
    function loadCSS(url, cacheS, successFunc) {
        htmx.ajax('GET', url, {
            swap: 'none',
            handler: (elt, info) => successFunc(info.xhr.responseText, info.xhr.statusText, info.xhr)
        });
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

    function setTableEvents() {
        if (window.setSortedClasses === undefined) {
            loadJScript("/js/table.js", false, true, setSortedClasses);
        }
    }

    // init runs once, on document ready - everything the old bare
    // `$(function () {...})` block at the bottom of app.js used to do.
    function init() {
        if (!window.onpopstate) {
            window.onpopstate = onPopState;
        }

        console.log("v1.2.199");
        window.addEventListener("beforeunload", evt => {
            evt = evt || window.event;

            if (evt) {
                var y = evt.pageY || evt.clientY;
                if (y === undefined) {
                    console.log(evt);
                }
                console.log(`beforeunload ${document.location} pageY:${y}`);
                evt.preventDefault();
                if (y < 0) {
                    return evt.returnValue = "Do you want to close this page?";
                }
            }
            return false;
        });

        Auth.ensureUser();

        processAll(document.body);
        createObserver();
    }

    // impotent set BEFORE load all srcipt
    cfgHTMX();

    return {
        init,
        fancyOpen,
        replMacros,
        GetPageLines: getPageLines,
        LoadStyles: loadStyles,
        loadCSSOnce,
        AddStyles: addStyles,
        AddScript: addScript,
        SetDocumentHash: setDocumentHash,
        loadContent,
        PutContent: putContent,
        SetContent: setContent,
        findAndReplaceElem,
        LoadJScript: loadJScript,
        LoadCSS: loadCSS,
        errorLoadResource,
        isScrollableY,
        isScrollableX,
        SetTableEvents: setTableEvents,
        get goHistory() {
            return goHistory;
        },
        set goHistory(v) {
            goHistory = v;
        },
    };
})();

$(App.init);

// --- Backward-compatible global aliases ------------------------------------
// Same names, same call signatures as the pre-singleton app.js - every file
// that still calls one of these bare (forms.js, observer.js, set_clicks.js,
// setFlatPickr.js, table-row.js, generated HTML) keeps working with zero
// changes. New code should prefer App.* directly. (relogin() moved to
// user.js - it's Auth.relogin(evt) now, not App.relogin(url).)
function fancyOpen(data) {
    return App.fancyOpen(data);
}

function replMacros(url) {
    return App.replMacros(url);
}

function GetPageLines() {
    return App.GetPageLines();
}

function LoadStyles(id, styles) {
    return App.LoadStyles(id, styles);
}

function loadCSSOnce(url) {
    return App.loadCSSOnce(url);
}

function AddStyles(name, css) {
    return App.AddStyles(name, css);
}

function AddScript(name, js) {
    return App.AddScript(name, js);
}

function SetDocumentHash(url, data) {
    return App.SetDocumentHash(url, data);
}

function loadContent(url) {
    return App.loadContent(url);
}

function PutContent(data, url) {
    return App.PutContent(data, url);
}

function SetContent(data) {
    return App.SetContent(data);
}

function findAndReplaceElem(src, selector, dst) {
    return App.findAndReplaceElem(src, selector, dst);
}

function LoadJScript(url, asyncS, cacheS, successFunc, completeFunc) {
    return App.LoadJScript(url, asyncS, cacheS, successFunc, completeFunc);
}

function LoadCSS(url, cacheS, successFunc) {
    return App.LoadCSS(url, cacheS, successFunc);
}

function errorLoadResource(xhr, textStatus, errorThrown) {
    return App.errorLoadResource(xhr, textStatus, errorThrown);
}

function isScrollableY(node) {
    return App.isScrollableY(node);
}

function isScrollableX(node) {
    return App.isScrollableX(node);
}

function SetTableEvents() {
    return App.SetTableEvents();
}

// go_history wasn't read anywhere outside app.js in the files reviewed, but
// it's a bare global that some unreviewed page script could still toggle
// (e.g. `go_history = 0` to suppress one popstate) - this keeps that working
// by aliasing it straight onto App.goHistory instead of silently dropping it.
Object.defineProperty(window, 'go_history', {
    get: () => App.goHistory,
    set: (v) => {
        App.goHistory = v;
    },
});