/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Observer watches the page for DOM changes htmx itself doesn't react to on
// its own: a *new* node (childList) needs the same processAll() pass any
// server-rendered content gets (SetDatesInputs/htmx.process/setClickAll),
// and now, an *existing* element whose attributes changed needs htmx.process()
// run on it again specifically when the changed attribute is one htmx reads
// (hx-get/hx-trigger/hx-target/... - e.g. a row highlighted/updated in place
// via plain JS that gains its hx-* wiring only after the fact, not at
// creation time). Converted from a bare createObserver() function into a
// singleton object, same shape as App/Auth - createObserver() below is kept
// as a thin global alias so app.js's existing `createObserver();` call in
// App.init() needs no change.
const Observer = (function () {
    let observer = null;

    // Attribute churn that's routine and never htmx-relevant - skipped
    // outright so a class toggle or inline style tweak doesn't even reach
    // the htmx-attribute check below. (Unchanged from the original list.)
    const IGNORED_ATTRIBUTES = new Set(["style", "class", "rel"]);

    // Only an attribute htmx itself reads is worth a re-process - matches
    // both the standard hx-* prefix and the data-hx-* alternate syntax htmx
    // also recognizes (hx-boost, hx-get/post/put/patch/delete, hx-trigger,
    // hx-target, hx-swap, hx-vals, hx-headers, hx-confirm, hx-include,
    // hx-select, hx-swap-oob, etc. - no need to enumerate them by name).
    function isHtmxAttribute(name) {
        return name != null && (name.startsWith("hx-") || name.startsWith("data-hx-"));
    }

    function handleMutations(mutationsList) {
        console.debug("mutations:", mutationsList);
        // A single batch can touch several hx-* attributes on the same
        // element (e.g. hx-get and hx-trigger set together) - htmx.process()
        // re-scans the whole element regardless, so track what's already
        // been (re)processed this batch instead of calling it once per
        // matching attribute.
        const reprocessed = new Set();

        for (const mutation of mutationsList) {
            switch (mutation.type) {
                case "childList":
                    mutation.addedNodes.forEach(node => processAll(node));
                    break;

                case "attributes": {
                    const name = mutation.attributeName;
                    // `continue` here, not the original code's `return` -
                    // `return` inside this for-of would have aborted the
                    // *entire* handler on the first ignored attribute,
                    // silently skipping every mutation still left in the
                    // batch (including real childList/attribute changes
                    // that came after it). `continue` only skips this one
                    // mutation, same as the switch's `break` does for the
                    // others.
                    if (IGNORED_ATTRIBUTES.has(name)) continue;

                console.log("Attributes changed:", mutation);
                    if (isHtmxAttribute(name) && !reprocessed.has(mutation.target)) {
                        reprocessed.add(mutation.target);
                        htmx.process(mutation.target);
                    }
                    break;
                }

                case "characterData":
                console.log("Text content changed:", mutation);
                    break;

                default:
                    console.error("Unhandled mutation type:", mutation);
            }
        }
    }

    function create() {
        if (observer) return observer; // idempotent - a second call is a no-op

        observer = new MutationObserver(handleMutations);
        observer.observe(document.body, {
        childList: true,      // Detect when children are added/removed
        attributes: true,     // Detect attribute changes
        subtree: true,        // Observe all descendants
        characterData: true   // Detect text content changes
        });
        console.debug('observer created');
        return observer;
    }

    // Not currently called anywhere - included so the object is a complete,
    // symmetric API rather than a one-way create(), in case a future page
    // (a fancybox-hosted subview, say) needs to stop watching.
    function disconnect() {
        observer?.disconnect();
        observer = null;
    }

    return {create, disconnect};
})();

function createObserver() {
    return Observer.create();
}

function processAll(parent) {
    console.debug('process');
     console.debug(parent);
    if (parent) {
        SetDatesInputs(parent);
        htmx.process(parent);
        setClickAll(parent);
        return;
    }
    console.error('is empty parent');
}

// Declarative alternative to saveForm(this, successFn, errFn): every form is
// already an htmx request on its own, courtesy of <body hx-boost="true">
// (see html.qtpl) - no hx-post/onsubmit needed for that part at all. A form
// that also wants a custom success/error callback, without any inline JS,
// can opt into the same dispatch used below by naming two *global*
// functions instead of passing live references:
//
//   <form ... data-success="afterLogin" data-error="showLoginError">
function namedFormHandlers(form) {
    const successFunction = form.dataset.success && window[form.dataset.success];
    const errorFunction = form.dataset.error && window[form.dataset.error];
    return (successFunction || errorFunction) ? {successFunction, errorFunction} : undefined;
}

// Single delegated dirty-tracking listener for every form on the page.
// Replaces the oninput="return FormIsModified(event, this)"/onchange="..."
// attributes the template used to stamp onto every generated <form> -
// html.qtpl no longer emits them, this is the one place that logic lives now.
//
// evt.target.form (not .closest('form')) is deliberate: a field can be
// form-associated without being a DOM descendant of its <form> at all, via
// the form="someId" content attribute - .closest() only walks ancestors and
// would silently miss that field. .form is the native, spec-correct way
// every form-associable element (input/select/textarea/button/output/
// fieldset) resolves "which form do I belong to", so it's checked first;
// .closest('form') is only the fallback for something that isn't natively
// form-associable to begin with (a contenteditable block, a custom element)
// but still happens to live inside a <form>.
function markFormModified(evt) {
    const form = evt.target.form || evt.target.closest('form');
    if (form) FormIsModified(evt, form);
}

function cfgHTMX() {
    document.body.addEventListener('input', markFormModified);
    document.body.addEventListener('change', markFormModified);

    document.body.addEventListener('htmx:onLoadError', function (evt) {
        console.log(evt);
        handleError(evt);
    });

    // Holds any *new* htmx request fired while a relogin (Auth.relogin(),
    // user.js) is in flight, instead of letting it go out and fail with its
    // own 401 too - the documented htmx async-auth pattern
    // (htmx.org/examples/async-auth/): preventDefault() the request here,
    // then call the SAME event's issueRequest(true) once the login
    // resolves. The `true` skips this same htmx:confirm on the reissue, so
    // it can't loop. The sign-in request itself is excluded so the login
    // form's own submit (and anything else that talks to it) is never held
    // up by the very reauth it's trying to complete - see isSignInRequest()
    // below for why that's keyed off the request's actual target/action
    // rather than a guessed element id.
    document.body.addEventListener('htmx:confirm', function (evt) {
        if (!Auth.isReauthInProgress()) return;
        if (isSignInRequest(evt)) return;

        evt.preventDefault();
        Auth.onceReady().then(() => evt.detail.issueRequest(true));
    });

    document.body.addEventListener('htmx:configRequest', function (evt) {
        // Single source of truth for auth/language headers - see
        // Auth.headers() in user.js. Every request in the app gets them from
        // here now: htmx requests via this listener, and forms.js/sendFile's
        // one remaining raw XHR (which can't go through htmx) by calling
        // Auth.headers() directly - over_click.js's getHeaders() shim for
        // raw jQuery $.ajax calls is gone along with the last of those calls.
        Object.assign(evt.detail.headers, Auth.headers());
        evt.detail.headers['X-Requested-With'] = 'XMLHttpRequest';
        evt.detail.headers['HX-Request'] = 'html';
        const pathInfo = evt.detail.path;
        if (pathInfo && pathInfo.finalRequestPath) {
            console.log(pathInfo);
            pathInfo.responsePath = replMacros(pathInfo.finalRequestPath);
        } else {
            console.log(evt);
        }

        const form = evt.detail.elt.closest?.('form');
        if (form) {
            const parameters = evt.detail.parameters;
            const isNewRecord = !form.querySelector('input[name="id"]');
            form.querySelectorAll('[name]').forEach(field => {
                const hiddenUnusedField = field.closest('figure:not([validated])') && !field.closest('figure').offsetParent;
                const remove = hiddenUnusedField || field.readOnly
                    || (!field.value && (isNewRecord || field.type === 'select-one' || field.type === 'file'));
                if (remove) delete parameters[field.name];
            });
            form.querySelectorAll('input[type="checkbox"][checked]:not(:checked)').forEach(field => {
                parameters[field.name] = '0';
            });
            parameters.is_get_form_actions = 'true';
        }
    });

    document.body.addEventListener('htmx:beforeRequest', evt => {
        const form = evt.detail.elt.closest?.('form');
        if (!form) return;

        const title = form.querySelector('h2')?.textContent
            || form.querySelector('figcaption')?.textContent
            || form.name || form.id;
        if (!validateFields(form) || (form.noValidate && !confirm(`Do you sure to send form "${title}"?`))) {
            evt.preventDefault();
            return;
        }
        form.querySelectorAll('figure:not([validated]) .input-label > input, figure:not([validated]) .input-label > select')
            .forEach(field => {
                if (!field.offsetParent) field.disabled = true;
            });
        form.querySelector('output')?.replaceChildren(document.createTextNode('Start sending...'));
        // Clear every trace of a *previous* failed submit - not just hide it -
        // so a field that was invalid last time but is fine now doesn't keep
        // showing a stale error label/red border through a fresh attempt.
        form.querySelectorAll('.errorLabel').forEach(label => {
            label.hidden = true;
            label.textContent = '';
        });
        form.querySelectorAll('.error-field').forEach(field => field.classList.remove('error-field'));
        form.querySelectorAll('progress, .loading').forEach(element => {
            element.hidden = false;
        });
    });

    document.body.addEventListener('htmx:xhr:progress', evt => {
        const form = evt.detail.elt.closest?.('form');
        if (!form || !evt.detail.total) return;
        const percent = Math.round((evt.detail.loaded / evt.detail.total) * 100);
        const output = form.querySelector('output');
        if (output) output.textContent = `Progress - ${percent}%`;
        form.querySelectorAll('progress').forEach(progress => {
            progress.value = percent;
        });
    });

    // === Before Swap ===
    document.body.addEventListener('htmx:beforeSwap', evt => {
        const elt = evt.detail.elt;
        const xhr = evt.detail.xhr;

        // showJSON()'s successor: OverClick() (over_click.js, now removed)
        // used to special-case a raw `application/json` response the same
        // way, in its own hand-rolled $.ajax success handler - restored
        // here as the shared path every plain (non-form) htmx request goes
        // through instead. Scoped to non-form requests only: a form's JSON
        // response (a create/update result, e.g.) already has its own,
        // more specific handling in htmx:afterRequest below and must not be
        // replaced by the generic dump. Also skipped when the element opted
        // into htmx-json.js's declarative template binding
        // (hx-swap="json") - that's a different, complementary tool for
        // JSON with a pre-authored template; showJSON is the schema-less
        // fallback for JSON with no template at all.
        if (!elt.closest?.('form')) {
            const contentType = xhr.getResponseHeader('Content-Type') || '';
            if (xhr.status === 200 && contentType.startsWith('application/json') && effectiveSwapStyle(elt) !== 'json') {
                evt.detail.shouldSwap = false;
                showJSON(parseHTMXResponse(xhr), evt.detail.target);
                return;
            }
        }

        if (!elt.attributes.target) {
            if (xhr.status === 204) {
                showMessage(elt, 'empty response!');
                evt.preventDefault();
            }
            return;
        }
        let responseText = xhr.responseText;
        switch (elt.attributes.target.nodeValue) {
            case "_modal":
                fancyOpen(responseText);
                evt.preventDefault();
                return false;

            case "_blank":
                var uri = "data:text/html," + encodeURIComponent(responseText);
                var newWindow = window.open('localhost', "Preview");
                newWindow.document.write(responseText);
                newWindow.focus();
                setTimeout(function () {
                    newWindow.setClickAll();
                }, 1000);

                evt.preventDefault();
                return false;
        }
    });

    // === After Request - Main handler ===
    // This is the ONLY place readEvents()/success/error handling for htmx
    // requests happens. A form branches to its own <output> element; every
    // other request falls through to the #content branch below. Nothing
    // else in the codebase should call readEvents() for an htmx-driven
    // request - see forms.js/saveForm() for why a second call path used to
    // exist and cause it to run twice per submit.
    document.body.addEventListener('htmx:afterRequest', evt => {
        const xhr = evt.detail.xhr;
        const form = evt.detail.elt.closest?.('form');
        if (form) {
            form.querySelectorAll('progress, .loading').forEach(element => {
                element.hidden = true;
            });
            if (!evt.detail.successful) return;

            const output = form.querySelector('output');
            const data = parseHTMXResponse(xhr);
            if (xhr.status === 206) {
                readEvents($(output), data);
                return;
            }
            if (output) output.textContent = xhr.statusText || 'Success';
            const handlers = htmxFormHandlers.get(form) || namedFormHandlers(form);
            if (handlers?.successFunction) {
                handlers.successFunction(data, form);
            } else if (isSignInForm(form)) {
                // No data-success="afterLogin" needed on the template's own
                // <form> - matched by the real endpoint (same rationale as
                // isSignInRequest() above), so a successful /user/signin/
                // response is always treated as login data (token/user -
                // afterLogin() does the localStorage write + saveUser()'s
                // full cascade) instead of falling into afterSaveAnyForm()'s
                // generic content_url/formActions/message handling, which
                // doesn't know what to do with a token payload and used to
                // leave the user looking at the raw response with no
                // session actually established.
                afterLogin(data, form);
            } else {
                afterSaveAnyForm(data || {}, xhr.statusText, form);
            }
            $.fancybox.close();
            return;
        }

        const target = evt.detail.target;
        const data = parseHTMXResponse(xhr);

        switch (xhr.status) {
            case 204:
                evt.preventDefault();
                showMessage(evt.detail.elt, 'empty response!');
                return false;

            case 206:
                readEvents($('#content'), data);
                return false;

            case 201: {
                // xhr.response is whatever the server sent - usually a JSON
                // object ({id, message, ...}), not a bare id, so pull a real
                // message out of it instead of stringifying the raw
                // response. Land it in a nearby <output> when the triggering
                // element has one (mirrors the form convention above);
                // otherwise a non-blocking toast, instead of alert()'s
                // modal, so adding one row after another doesn't force a
                // click through a dialog every time.
                const message = data?.message
                    || (data?.id !== undefined ? `Successful add record #${data.id}` : 'Successful add record');
                showMessage(evt.detail.elt, message);
                // fallthrough intentional: a fresh record still needs the
                // normal 200 handling (fancybox/data-fancybox check, target
                // refresh, processAll) run right after showing the message.
            }
            case 200:
                if (evt.target.matches("[data-fancybox]")) {
                    fancyOpen(xhr.response);
                    return;
                }
                htmx.trigger(target, 'on_change');
                processAll(target);
        }
    });

    // === Response Error ===
    // This is the direct htmx equivalent of the old jQuery Form Plugin
    // error callback ($(thisForm).ajaxSubmit({ error: ... })): a custom
    // errorFunction (registered via saveForm(this, ok, err) or a
    // data-error="fnName" attribute) fully overrides the default handling,
    // exactly like before; otherwise 401 re-triggers login, 400 populates
    // per-field error labels, and anything else falls back to the same
    // alert(xhr.responseText) the legacy code used for an unclassified error.
    document.body.addEventListener('htmx:responseError', evt => {
        const xhr = evt.detail.xhr;
        const form = evt.detail.elt.closest?.('form');

        if (form) {
            form.querySelectorAll('progress, .loading').forEach(element => {
                element.hidden = true;
            });
            const output = form.querySelector('output');
            if (output) output.textContent = xhr.responseText;

            const handlers = htmxFormHandlers.get(form) || namedFormHandlers(form);
            if (handlers?.errorFunction) {
                handlers.errorFunction(xhr.statusText, form);
                return;
            }

            evt.preventDefault();
        switch (xhr.status) {
            case 401:
                return Auth.relogin(evt);
            case 400:
                showErrors(parseHTMXResponse(xhr)?.formErrors, form);
                return;
            default:
                showMessage(form, xhr.responseText);
                return;
        }
        }

        // Non-form htmx requests (hx-get links, table filters, etc.) - no
        // <output>/errorFunction to report through, so this mirrors the form
        // branch above as closely as it can: 401 re-authenticates (and, via
        // Auth.relogin(evt)'s retry queue, replays this exact request once
        // logged back in), 400 populates field errors when there's a form-
        // shaped target for them, and anything else falls back to the same
        // alert(xhr.responseText) the form branch's default case uses.
        evt.preventDefault();
        switch (xhr.status) {
            case 401:
                return Auth.relogin(evt);
            case 400:
                showErrors(parseHTMXResponse(xhr)?.formErrors, evt.detail.elt);
                return;
            default:
                showMessage(evt.detail.elt, xhr.responseText);
                return;
        }
    });

    // Passkey-first login, without needing a per-form
    // onsubmit="return Auth.login(event, this)" attribute in the template:
    // any submit of the real sign-in form (matched by action, same as
    // isSignInRequest()/the afterRequest routing above - not a guessed id
    // the markup may not carry) is handed to Auth.login(), which tries a
    // passkey first and, only if that's unsupported/declined/fails, lets
    // this exact submit continue as a normal password POST. Bound with
    // { capture: true } on document.body so it runs *before* htmx's own
    // submit listener (bound directly on the form itself, by
    // htmx.process()) sees the event: Auth.login()'s event.preventDefault()
    // (called synchronously, before its first await, when it decides to
    // attempt a passkey) has to land before htmx would otherwise go ahead
    // and issue its own boosted request for the same submit.
    document.body.addEventListener('submit', evt => {
        const form = evt.target;
        if (form.tagName === 'FORM' && isSignInForm(form)) {
            Auth.login(evt, form);
        }
    }, true);
}

// The real sign-in endpoint (routes.qtpl: "/user/signin/") - the one stable
// thing every piece of sign-in-specific handling below can key off, instead
// of an id (#fLogin/#bLogin) the actual template may or may not carry.
const SIGNIN_ACTION = '/user/signin/';

function isSignInForm(form) {
    return form?.getAttribute('action') === SIGNIN_ACTION;
}

// Is this htmx:confirm event about to fire the actual sign-in request (or
// something else the login UI itself needs)? The original check was
// `evt.target.closest('#fLogin, #bLogin')` - it assumed the app's real
// markup stamps those ids on the login form/its trigger button. This app's
// actual sign-in form (POST /user/signin/, per routes.qtpl) carries neither
// id, so that check never matched it: once Auth.isReauthInProgress() got
// stuck true (a background request 401'd and nothing ever completed the
// reauth - see beginReauth()'s escape valve in user.js for why that can no
// longer last forever), the reauth gate above would preventDefault() the
// login form's own submit and await a promise that never resolved -
// completely silently, with no console error, no network request, nothing:
// exactly "the form doesn't submit". Matching on the actual request target
// (the form's action, or an hx-get/hx-post element's own path) instead of a
// guessed id fixes that regardless of what ids (if any) the real markup
// uses; #fLogin/#bLogin are still checked too, for any part of the app that
// *does* use that convention (e.g. a modal login trigger).
function isSignInRequest(evt) {
    const elt = evt.target;
    if (elt.closest?.('#fLogin, #bLogin')) return true;

    const form = elt.closest?.('form');
    if (isSignInForm(form)) return true;

    const action = elt.getAttribute?.('hx-post') || elt.getAttribute?.('hx-get');
    return action === SIGNIN_ACTION;
}

// Show a short confirmation near whatever triggered the request. Forms have
// their own <output> (see htmx:afterRequest above); a non-form trigger
// (a hx-post button, an inline row action) may sit next to one too - use it
// when present, otherwise fall back to a small auto-dismissing toast rather
// than a blocking alert(), which gets old fast on repeated "record added"
// actions. Toast styling (.htmx-toast) needs a couple of CSS rules in the
// site's stylesheet - this only creates/removes the element.
function showMessage(elt, message) {
    if (!message) return;
    const output = elt?.closest('[data-with-output]')?.querySelector('output')
        || (elt?.parentElement?.querySelector(':scope > output'));
    if (output) {
        output.textContent = message;
        return;
    }
    showToast(message);
}

function showToast(message, timeout = 4000) {
    const toast = document.createElement('div');
    toast.className = 'htmx-toast';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), timeout);
}

function showErrors(formErrors, thisForm) {
    if (formErrors) {
        let out = $('output', thisForm);
        for (let key in formErrors) {
            let elem = $(`[name=${key}]`, thisForm);
            if (elem.length > 0) {
                elem.nextAll('.errorLabel').show().text(formErrors[key]);
                elem.addClass('error-field').focus();
                if (elem.nextAll('.errorLabel').length === 0) {
                    elem.after(`<h6 class="errorLabel" style="display:flex; font-size: smaller"> ${formErrors[key]} </h6>`);
                }
                elem.parents('details').attr('open', true);
            } else {
                out.text(formErrors[key]);
            }
        }
    }
}

function parseHTMXResponse(xhr) {
    if (!xhr.responseText) return null;
    try {
        return JSON.parse(xhr.responseText);
    } catch (_) {
        return null;
    }
}

// Moved in from over_click.js (now removed): the shared handler for
// htmx:onLoadError (a request that failed before it even got a response -
// network error, CORS, etc.) and, before this refactor, every raw $.ajax
// call's own error callback. 401 goes through the exact same Auth.relogin(evt)
// every other error path uses now, so a network hiccup during a stale
// session recovers the same way a normal 401 does.
function handleError(evt) {
    const xhr = evt.detail.xhr;
    switch (xhr.status) {
        case 401:
            return Auth.relogin(evt);
        case 404:
            showMessage(evt.detail?.elt, `Request page not found: ${xhr.responseURL || ''}`);
            return;
    }

    showMessage(evt.detail?.elt, `Code : "${xhr.status}": "${xhr.responseText}"`);
    console.error(`Code : "${xhr.status}": "${xhr.responseText}"`, xhr);
}

// What hx-swap style would actually apply to a request from this element -
// read directly off the DOM (htmx's own attribute-inheritance convention:
// the nearest ancestor - or self - carrying hx-swap), not off an htmx event
// internal whose exact shape isn't documented. Used by the
// htmx:beforeSwap listener above to tell a plain JSON response (render it
// with showJSON() below) apart from one a hx-swap="json" element is already
// handling declaratively via htmx-json.js's json-swap extension.
function effectiveSwapStyle(elt) {
    return elt.closest?.('[hx-swap]')?.getAttribute('hx-swap');
}

// showJSON() is the schema-less counterpart to htmx-json.js's json-swap
// extension: json-swap needs a <template json-each>/json-if/... already
// authored in the page (opted into per-element via hx-swap="json") to bind
// JSON into; showJSON just recursively dumps whatever JSON it's given as
// generic HTML, with no template required. Restored from over_click.js's
// OverClick() (removed - see that file's old history) essentially unchanged,
// wired to the shared htmx:beforeSwap listener above instead of its own
// hand-rolled $.ajax success handler. One bug fixed in the restore: the
// original did `let div = divContent.append('<div>')` - jQuery's .append()
// returns the ORIGINAL collection, not the element it just created, so
// every "new" div was actually the same top-level container; nested
// object/array properties ended up flattened into it instead of grouped
// under their own div. `$('<div>').appendTo(target)` returns the new
// element itself, which is what the per-object grouping below actually
// needs.
function showJSON(data, target) {
    if (!data) {
        showMessage(null, 'no results!');
        return false;
    }

    const $target = $(target || '#content').html('');

    function showJsonElem(elem, $into) {
        let x, y;
        if (Array.isArray(elem)) {
            for (x in elem) {
                showJsonElem(elem[x], $into);
            }
        } else if (elem instanceof Object) {
            let div = $('<div>').appendTo($into);
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
                                showJsonElem(elem[y][x], div);
                                div.append("<br>");
                            }
                        } else if (elem[y] instanceof Object) {
                            div.append(`<h4>${y}</h4>`);
                            showJsonElem(elem[y], div);
                        } else {
                            div.append(`<p><i>${y}:</i> <span>${elem[y]}</span></p>`);
                        }
                }
            }
        } else {
            $into.append(`<span>${elem}</span>`);
        }
    }

    showJsonElem(data, $target);
    return true;
}

function isSelfRequest(ctx, event) {
    const trigger = event.detail?.requestConfig?.elt;
    return trigger === ctx;
}
