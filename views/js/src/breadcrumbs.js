/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Breadcrumbs: where the user is and the way back.
//
// The page holds <breadcrumbs id="breadcrumbs"><ol class="breadcrumb"></ol></breadcrumbs>
// (index.qtpl). "Home" is not a crumb: it is the site-icon button of the top
// menu (header.qtpl). What this file does:
//
//   * remembers every page that was loaded into #content by a GET - a link, a
//     table, a form to edit a row, a search result - as a trail
//     [{url, label, pushed}], per browser tab (sessionStorage);
//   * draws the trail as crumbs; a click on an earlier crumb returns to that
//     page;
//   * keeps a Back button with the name of the previous page - from an edit
//     form to the table it was opened from in ONE click.
//
// Server-made crumbs win: when the response brings its own
// <breadcrumbs id="breadcrumbs" hx-swap-oob="true"> with items, the trail is
// not drawn over them (the Back button still is).
//
// Returning prefers the browser history (history.go): htmx then restores the
// stored snapshot - the table as it was, with its loaded rows. When a page on
// the way was loaded without a history entry (hx-push-url="false"), the page is
// requested again with navigate() instead.
//
// A page loaded in a dialog (fancybox, hx-target="_modal") never reaches
// #content and so never enters the trail.
(function () {
    'use strict';
    if (window.Breadcrumbs) return;

    const KEY = 'httpgo.trail';
    const MAX = 10;
    const CONTENT = 'content';

    let trail = [];        // [{url, label, pushed}], the last one is the current page
    let lastLoc = '';      // location.pathname+search at the last sync
    let titleBefore = '';  // document.title when the request left
    let started = false;   // the first view of the page is already in the trail

    const here = () => location.pathname + location.search;

    function load() {
        try {
            const t = JSON.parse(sessionStorage.getItem(KEY));
            return Array.isArray(t) ? t.filter(e => e && typeof e.url === 'string' && typeof e.label === 'string') : [];
        } catch (e) {
            return [];
        }
    }

    function save() {
        try {
            sessionStorage.setItem(KEY, JSON.stringify(trail));
        } catch (e) { /* storage blocked - the trail just lives until reload */
        }
    }

    // a request path as "/path?query" (htmx gives it with or without origin)
    function pathOf(p) {
        try {
            const u = new URL(p, location.origin);
            // htmx's cache buster param is not part of the address
            u.searchParams.delete('org.htmx.cache-buster');
            return u.pathname + u.search;
        } catch (e) {
            return p;
        }
    }

    // the name of the page read from the screen: its own heading, else the new
    // <title>; '' when the screen does not say
    function labelOf() {
        const content = document.getElementById(CONTENT);
        const h = content?.querySelector('h1, h2, h3');
        const text = h?.textContent.replace(/\s+/g, ' ').trim();
        if (text) return text;

        const title = document.title.trim();
        return title && title !== titleBefore ? title : '';
    }

    // last resort name: the end of the address
    function nameOf(url) {
        const last = url.split('?')[0].split('/').filter(Boolean).pop();
        try {
            return last ? decodeURIComponent(last) : document.title.trim() || url;
        } catch (e) {
            return last || url;
        }
    }

    // ---- the trail ---------------------------------------------------------

    function sync(url, label, pushed) {
        const i = trail.findIndex(e => e.url === url);
        if (i >= 0) {
            // came back (or reloaded) - the later pages are left behind; a name
            // the screen does not give any more is the one already known
            trail.length = i + 1;
            if (label) trail[i].label = label;
        } else {
            trail.push({url, label: label || nameOf(url), pushed});
            if (trail.length > MAX) trail.shift();
        }
        lastLoc = here();
        save();
        render();
    }

    function onSettled(evt) {
        const d = evt.detail || {};
        const target = d.target;
        if (!target || target.id !== CONTENT || evt.target !== target) return;

        const cfg = d.requestConfig || {};
        if (String(cfg.verb || 'get').toLowerCase() !== 'get') return;
        if (d.xhr && d.xhr.status >= 400) return;

        const loc = here();
        const pushed = loc !== lastLoc;
        const req = pathOf(d.pathInfo?.finalRequestPath || d.pathInfo?.requestPath || loc);

        // the content #content loads by itself (hx-trigger="load once") is
        // the page the visitor opened, not a step away from it
        if (!started && cfg.triggeringEvent?.type === 'load') {
            started = true;
            trail = [{url: req, label: labelOf() || nameOf(req), pushed: false}];
            lastLoc = loc;
            save();
            render();
            return;
        }
        started = true;

        // when the address did not change the page is known by its request
        sync(pushed ? loc : req, labelOf(), pushed);
    }

    // browser Back/Forward (htmx restores the page from its snapshot)
    function onRestore() {
        const loc = here();
        const i = trail.findIndex(e => e.url === loc);
        if (i >= 0) {
            // the restored snapshot has no new heading/title to read - keep the stored name
            trail.length = i + 1;
        } else {
            trail = [{url: loc, label: labelOf() || nameOf(loc), pushed: true}];
        }
        lastLoc = loc;
        save();
        render();
    }

    // ---- going back --------------------------------------------------------

    function go(idx) {
        const k = trail.length - 1 - idx;
        if (idx < 0 || k <= 0) return;
        const entry = trail[idx];
        const contiguous = trail.slice(idx + 1).every(e => e.pushed);

        const request = () => {
            trail.length = idx + 1;
            save();
            render();
            if (typeof navigate === 'function') {
                navigate(entry.url);
            } else {
                location.href = entry.url;
            }
        };

        if (!contiguous) {
            request();
            return;
        }
        const before = here();
        history.go(-k);
        // history did not move (entries lost, other tab...) - ask for the page
        setTimeout(() => {
            if (here() === before && trail.length - 1 > idx) request();
        }, 700);
    }

    function back() {
        go(trail.length - 2);
    }

    // ---- drawing -----------------------------------------------------------

    function render() {
        const bc = document.getElementById('breadcrumbs') || document.querySelector('breadcrumbs');
        if (!bc) return;

        let btn = bc.querySelector(':scope > .bc-back');
        if (!btn) {
            btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'bc-back';
            btn.innerHTML = '<i class="bi bi-arrow-left" aria-hidden="true"></i><span></span>';
            bc.prepend(btn);
        }
        const prev = trail.length > 1 ? trail[trail.length - 2] : null;
        btn.hidden = !prev;
        if (prev) {
            btn.querySelector('span').textContent = prev.label;
            btn.title = btn.ariaLabel = 'Back to ' + prev.label;
        }

        let ol = bc.querySelector('ol.breadcrumb');
        if (!ol) {
            ol = document.createElement('ol');
            ol.className = 'breadcrumb';
            bc.append(ol);
        }
        ol.querySelectorAll(':scope > li[data-trail]').forEach(li => li.remove());
        if (ol.querySelector(':scope > li')) return; // the server drew its own crumbs

        trail.forEach((e, i) => {
            const li = document.createElement('li');
            li.dataset.trail = i;
            li.className = 'breadcrumb-item';
            if (i === trail.length - 1) {
                li.classList.add('active');
                li.setAttribute('aria-current', 'page');
                li.textContent = e.label;
                li.title = e.label;
            } else {
                const a = document.createElement('a');
                a.href = e.url;
                a.dataset.bcIdx = i;
                a.setAttribute('hx-boost', 'false');
                a.textContent = e.label;
                a.title = e.label;
                li.append(a);
            }
            ol.append(li);
        });
    }

    // ---- wiring ------------------------------------------------------------

    document.addEventListener('click', evt => {
        if (evt.defaultPrevented || evt.button !== 0 || evt.metaKey || evt.ctrlKey || evt.shiftKey || evt.altKey) return;
        const el = evt.target.closest?.('breadcrumbs .bc-back, breadcrumbs a[data-bc-idx]');
        if (!el) return;
        evt.preventDefault();
        evt.stopPropagation();
        if (el.classList.contains('bc-back')) back();
        else go(Number(el.dataset.bcIdx));
    }, true);

    document.addEventListener('htmx:beforeRequest', evt => {
        if (evt.detail?.target?.id === CONTENT) titleBefore = document.title.trim();
    });
    document.addEventListener('htmx:afterSettle', onSettled);
    document.addEventListener('htmx:historyRestore', onRestore);

    function init() {
        const loc = here();
        const saved = load();
        // a reload keeps the trail; arriving on another address starts a new one
        trail = saved.length && saved[saved.length - 1].url === loc ? saved : [];
        lastLoc = loc;
        if (trail.length) started = true;
        if (!document.getElementById(CONTENT)?.hasAttribute('hx-get')) {
            // the page itself is the first view
            started = true;
            if (!trail.length) trail = [{url: loc, label: labelOf() || nameOf(loc), pushed: false}];
        }
        titleBefore = document.title.trim();
        save();
        render();
    }

    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
    else init();

    window.Breadcrumbs = {back, go, trail: () => trail.map(e => ({...e})), render};
})();
