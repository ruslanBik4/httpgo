/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

"use strict";

// App is the page-lifecycle/content-loading singleton: one instance, built
// once at parse time, holding what used to be a dozen unrelated top-level
// functions. Everything below the IIFE is a same-named global wrapper (e.g.
// `function loadContent(url){ return App.loadContent(url); }`) so every
// existing caller - forms.js, observer.js, set_clicks.js, setFlatPickr.js,
// table-row.js, and any onclick="..."/hx-on="..." attribute already baked into
// rendered HTML - keeps working unchanged. New code should call App.* directly.
//
// This file no longer touches page regions by hand. Content, sidebar,
// breadcrumbs and <title> are swapped by htmx from the response (swap target +
// out-of-band elements), history/back/forward are htmx's own (hx-push-url),
// and every request - hx-* element, boosted link/form, or script - gets its
// auth/language headers from Auth.headers() via observer.js's
// htmx:configRequest. See navigate()/loadContent() below.
const App = (function () {

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

    // Script-initiated navigation, done the htmx way - the single replacement
    // for the old setContent()/putContent()/findAndReplaceElem()/
    // setDocumentHash()/onPopState() set, which fetched a page with a custom
    // handler and then copied pieces of it (#content, breadcrumbs, sidebar,
    // topline, <title>) into the DOM by hand and kept their own history.
    // Now htmx does all of that from the response itself:
    //   * the main body is swapped into `target` (innerHTML);
    //   * every other region the page owns arrives as an out-of-band element
    //     in the same response (<div id="catalog_sidebar" hx-swap-oob="true">,
    //     <breadcrumbs id="breadcrumbs" hx-swap-oob="true">, ...) - see
    //     clearStaleSidebar() in observer.js for the "no sidebar" case;
    //   * <title> in the response sets document.title;
    //   * history, back/forward and the page snapshot are htmx's own
    //     (hx-push-url) - no popstate handler of ours.
    // htmx reads hx-push-url from the request's source element, so each call
    // gets a throw-away hidden one (a shared element would make htmx drop a
    // second request while the first is still in flight). Being a normal
    // htmx request it picks up Auth.headers() in htmx:configRequest, and a
    // 401 goes through htmx:responseError -> Auth.relogin(evt).
    function navigate(url, {target = '#content', select, push = true} = {}) {
        const source = document.createElement('span');
        source.hidden = true;
        if (push) source.setAttribute('hx-push-url', 'true');
        document.body.append(source);

        const done = () => setTimeout(() => source.remove(), 0);
        return htmx.ajax('GET', url, {source, target, select, swap: 'innerHTML'}).then(done, done);
    }

    function loadContent(url) {
        return navigate(url);
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
        console.log("v1.3.0");
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
        loadContent,
        navigate,
        LoadJScript: loadJScript,
        LoadCSS: loadCSS,
        errorLoadResource,
        isScrollableY,
        isScrollableX,
        SetTableEvents: setTableEvents,
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

function loadContent(url) {
    return App.loadContent(url);
}

function navigate(url, opts) {
    return App.navigate(url, opts);
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