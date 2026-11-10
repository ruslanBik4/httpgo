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
//   * remembers the pages that were loaded into #content by a GET - a link, a
//     table, a form to edit a row, a search result - as a trail
//     [{url, label, pushed, state}], per browser tab (sessionStorage);
//   * a page opened from a MENU (top, footer, left, the Home icon) is a new
//     start, not a step: the trail begins with it again. The same address
//     (the language, the paging parameters do not count) is never listed
//     twice: coming back to a page already in the trail cuts the trail there;
//   * draws the trail as crumbs; a click on an earlier crumb returns to that
//     page;
//   * keeps a Back button with the name of the previous page - from an edit
//     form to the table it was opened from in ONE click;
//   * keeps the STATE of a page while the user is away (the filters and the
//     scroll of a table, see registerState()) and puts it back on return;
//   * after a record is saved (afterSave(), called from afterSaveAnyForm())
//     returns to the page the form was opened from and loads it again, so the
//     table shows the new data - instead of leaving the form or an old copy.
//
// Server-made crumbs win: when the response brings its own
// <breadcrumbs id="breadcrumbs" hx-swap-oob="true"> with items, the trail is
// not drawn over them (the Back button still is).
//
// Returning prefers the browser history (history.go): htmx then restores the
// stored snapshot - the table as it was, with its loaded rows. When a page on
// the way was loaded without a history entry (hx-push-url="false"), the page is
// requested again instead.
//
// A page loaded in a dialog (fancybox, hx-target="_modal") or with
// hx-swap="none" never reaches #content and so never enters the trail.
(function () {
    'use strict';
    if (window.Breadcrumbs) return;

    const KEY = 'httpgo.trail';
    const MAX = 10;
    const CONTENT = 'content';
    // a page reached from one of these starts a new trail (data-bc-root marks any other)
    const MENU_SEL = '.topline-navbar, .footer-mnu, .home-link, .left-mnu-list, [data-bc-root]';
    // requests that open a dialog, not a page
    const DIALOG_SEL = '[data-fancybox], [target="_modal"], #bLogin';
    const DIALOG_PATHS = ['/show/forms/sign', '/user/sign'];
    // parameters that do not make another page
    const VOLATILE = new Set(['lang', 'offset', 'limit', 'counter', '_', 'org.htmx.cache-buster']);

    let trail = [];          // [{url, label, pushed, state}], the last one is the current page
    let lastLoc = '';        // location.pathname+search at the last sync
    let titleBefore = '';    // document.title when the request left
    let started = false;     // the first view of the page is already in the trail
    let shown = null;        // the entry whose page is on the screen (the trail may already be cut)
    let returning = null;    // the entry a request is bringing the user back to
    let refreshing = null;   // {entry, source}: the current page is being loaded again in place
    let pendingRefresh = false; // history.go() is under way, load the page again when it is there
    const providers = new Map(); // name -> {capture(), apply(state, fresh)}

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

    function pathnameOf(p) {
        try {
            return new URL(p, location.origin).pathname.replace(/\/+$/, '') || '/';
        } catch (e) {
            return p;
        }
    }

    // what makes a page "the same page": the path and the parameters that
    // choose the content (not the language, not the paging)
    function keyOf(p) {
        try {
            const u = new URL(p, location.origin);
            const params = [...u.searchParams].filter(([k]) => !VOLATILE.has(k))
                .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
                .map(([k, v]) => k + '=' + v);
            return pathnameOf(p) + (params.length ? '?' + params.join('&') : '');
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

    // is this request a page for #content (and not a dialog, a message, a partial load)?
    function isViewRequest(d) {
        const cfg = d.requestConfig || {};
        if (String(cfg.verb || 'get').toLowerCase() !== 'get') return false;
        const elt = cfg.elt;
        if (elt?.matches?.(DIALOG_SEL)) return false;
        const swap = elt && typeof effectiveSwapStyle === 'function' ? effectiveSwapStyle(elt) : '';
        if (swap === 'none') return false;
        const path = pathnameOf(d.pathInfo?.finalRequestPath || d.pathInfo?.requestPath || '/');
        return !DIALOG_PATHS.some(p => path.startsWith(p));
    }

    // ---- the state of a page ------------------------------------------------

    // A page part (a table...) tells how to keep and put back what the user did
    // on it: capture() -> a plain object or null, apply(object, fresh) - fresh
    // is true when the page was just loaded again (rows are new), false when
    // it comes from the history snapshot (rows are as they were).
    function registerState(name, provider) {
        providers.set(name, provider);
    }

    // keep the state of the page that is being left
    function capture() {
        const top = shown;
        if (!top) return;

        const state = {scroll: window.scrollY || 0, parts: {}};
        providers.forEach((p, name) => {
            try {
                const v = p.capture();
                if (v) state.parts[name] = v;
            } catch (e) {
                console.error('breadcrumbs: state of', name, e);
            }
        });
        // the address as the user left it (filters, order...) - only while the
        // browser still shows this page
        if (keyOf(here()) === keyOf(top.url)) state.href = here();
        top.state = state;
        save();
    }

    function applyState(entry, fresh) {
        const state = entry?.state;
        if (!state) return;

        providers.forEach((p, name) => {
            const v = state.parts?.[name];
            if (!v) return;
            try {
                p.apply(v, fresh);
            } catch (e) {
                console.error('breadcrumbs: state of', name, e);
            }
        });
        if (state.scroll) {
            try {
                window.scrollTo(0, state.scroll);
            } catch (e) { /* no scrolling here */
            }
        }
    }

    // the address a page is asked from again: where the user left it, without the paging
    function reloadUrl(entry) {
        try {
            const u = new URL(entry.state?.href || entry.url, location.origin);
            u.searchParams.delete('offset');
            u.searchParams.delete('limit');
            return u.pathname + u.search;
        } catch (e) {
            return entry.url;
        }
    }

    // ---- loading -------------------------------------------------------------

    // GET url into #content, never from the browser's cache. pushTo / replaceTo
    // are the address shown after it (the request itself may carry more).
    function fetchView(url, {pushTo, replaceTo} = {}) {
        const source = document.createElement('span');
        source.hidden = true;
        source.dataset.bcFetch = '1'; // our own request: the state was kept before it
        if (pushTo) source.setAttribute('hx-push-url', pushTo);
        else if (replaceTo) source.setAttribute('hx-replace-url', replaceTo);
        else source.setAttribute('hx-push-url', 'false');
        document.body.append(source);

        const done = () => setTimeout(() => source.remove(), 0);
        const request = htmx.ajax('GET', url, {
            source, target: '#' + CONTENT, swap: 'innerHTML',
            headers: {'Cache-Control': 'no-cache'},
        });
        request.then(done, done);
        return {source, request};
    }

    // load the current page again in place (same address, fresh data)
    function refreshCurrent() {
        const entry = trail[trail.length - 1];
        if (!entry) return;
        const {source} = fetchView(reloadUrl(entry), {replaceTo: entry.url});
        refreshing = {entry, source};
    }

    // open a page (a redirect after a save) fresh and as a step of the trail
    function open(url) {
        capture();
        fetchView(url, {pushTo: 'true'});
    }

    // ---- the trail ---------------------------------------------------------

    function sync(url, label, pushed, root) {
        if (root) {
            // a menu item is a new start
            trail = [{url, label: label || nameOf(url), pushed}];
            returning = null;
        } else {
            const key = keyOf(url);
            let i = trail.findIndex(e => keyOf(e.url) === key);
            const top = trail[trail.length - 1];
            if (i < 0 && top && label && label === top.label && pathnameOf(top.url) === pathnameOf(url)) {
                i = trail.length - 1; // the same page under another query
            }
            if (i >= 0) {
                // came back (or reloaded) - the later pages are left behind; a
                // name the screen does not give any more is the one already known
                trail.length = i + 1;
                const entry = trail[i];
                entry.url = url;
                entry.pushed = pushed || entry.pushed;
                if (label) entry.label = label;
                if (returning === entry) applyState(entry, true);
            } else {
                trail.push({url, label: label || nameOf(url), pushed});
                if (trail.length > MAX) trail.shift();
            }
            returning = null;
        }
        shown = trail[trail.length - 1];
        lastLoc = here();
        save();
        render();
    }

    function onSettled(evt) {
        const d = evt.detail || {};
        const target = d.target;
        if (!target || target.id !== CONTENT || evt.target !== target) return;
        if (!isViewRequest(d)) return;
        if (d.xhr && d.xhr.status >= 400) return;

        const cfg = d.requestConfig || {};

        // the current page was loaded again in place: same page, new data
        if (refreshing && cfg.elt === refreshing.source) {
            const entry = refreshing.entry;
            refreshing = null;
            shown = entry;
            const label = labelOf();
            if (label) entry.label = label;
            lastLoc = here();
            save();
            render();
            applyState(entry, true);
            return;
        }

        const loc = here();
        const pushed = loc !== lastLoc;
        const req = pathOf(d.pathInfo?.finalRequestPath || d.pathInfo?.requestPath || loc);

        // the content #content loads by itself (hx-trigger="load once") is
        // the page the visitor opened, not a step away from it
        if (!started && cfg.triggeringEvent?.type === 'load') {
            started = true;
            trail = [{url: req, label: labelOf() || nameOf(req), pushed: false}];
            shown = trail[0];
            lastLoc = loc;
            save();
            render();
            return;
        }
        started = true;

        const url = pushed ? loc : req;
        sync(url, labelOf(), pushed, !!cfg.elt?.closest?.(MENU_SEL));
    }

    // browser Back/Forward (htmx restores the page from its snapshot)
    function onRestore() {
        const loc = here();
        const i = trail.findIndex(e => keyOf(e.url) === keyOf(loc));
        if (i >= 0) {
            // the restored snapshot has no new heading/title to read - keep the stored name
            trail.length = i + 1;
            applyState(trail[i], false);
        } else {
            trail = [{url: loc, label: labelOf() || nameOf(loc), pushed: true}];
        }
        shown = trail[trail.length - 1];
        lastLoc = loc;
        save();
        render();

        if (pendingRefresh) {
            pendingRefresh = false;
            refreshCurrent();
        }
    }

    // ---- going back --------------------------------------------------------

    // Return to trail[idx]. fresh: load the page again so it shows new data.
    function go(idx, {fresh = false} = {}) {
        const k = trail.length - 1 - idx;
        if (idx < 0 || k < 0) return;

        capture();
        if (k === 0) {
            if (fresh) refreshCurrent();
            return;
        }

        const entry = trail[idx];
        const contiguous = trail.slice(idx + 1).every(e => e.pushed);

        const request = () => {
            trail.length = idx + 1;
            returning = entry;
            save();
            render();
            fetchView(reloadUrl(entry), {pushTo: entry.url});
        };

        if (!contiguous) {
            request();
            return;
        }
        pendingRefresh = fresh;
        const before = here();
        history.go(-k);
        // history did not move (entries lost, other tab...) - ask for the page
        setTimeout(() => {
            if (here() === before && trail.length - 1 > idx) {
                pendingRefresh = false;
                request();
            }
        }, 700);
    }

    function back() {
        go(trail.length - 2);
    }

    // After a record was saved: show the page the form was opened from, loaded
    // again. Returns true when it did (the caller must not open anything else).
    // `data` is the server's answer: its redirect (formActions / content_url)
    // wins when it points to another page than the one the user came from.
    function afterSave(form, data) {
        if (!trail.length) return false;

        const redirect = data?.formActions?.[0]?.url || data?.content_url;
        // a form in the page replaced it: the page before is the one to return to;
        // a form in a dialog lies over the current page
        const idx = form?.closest?.('#' + CONTENT) ? trail.length - 2 : trail.length - 1;
        const entry = trail[idx];
        if (!entry) return false;
        if (redirect && pathnameOf(entry.url) !== pathnameOf(redirect)) return false;

        go(idx, {fresh: true});
        return true;
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
        const d = evt.detail || {};
        if (d.target?.id !== CONTENT || !isViewRequest(d)) return;
        titleBefore = document.title.trim();
        // the page is about to be left: keep what the user did on it (not when
        // this is our own load - the screen then still shows the page we left)
        if (!d.requestConfig?.elt?.dataset?.bcFetch) capture();
    });
    // the user leaves by the browser's Back/Forward - the snapshot is taken now
    document.addEventListener('htmx:beforeHistorySave', capture);
    document.addEventListener('htmx:afterSettle', onSettled);
    document.addEventListener('htmx:historyRestore', onRestore);

    function init() {
        const loc = here();
        const saved = load();
        // a reload keeps the trail; arriving on another address starts a new one
        trail = saved.length && keyOf(saved[saved.length - 1].url) === keyOf(loc) ? saved : [];
        lastLoc = loc;
        if (trail.length) started = true;
        if (!document.getElementById(CONTENT)?.hasAttribute('hx-get')) {
            // the page itself is the first view
            started = true;
            if (!trail.length) trail = [{url: loc, label: labelOf() || nameOf(loc), pushed: false}];
        }
        titleBefore = document.title.trim();
        shown = trail[trail.length - 1] || null;
        save();
        render();
    }

    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
    else init();

    window.Breadcrumbs = {
        back, go, afterSave, open, registerState,
        refresh: refreshCurrent,
        trail: () => trail.map(e => ({...e})),
        render,
    };
})();
