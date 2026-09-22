/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

const Auth = (function () {

    let userStruct;
    let urlAfterLogin = '';
    let pendingRequests = [];
    let reauthPromise = null;
    let reauthResolve = null;
    let reauthTimeout = null;

    // --- session bootstrap / persistence -----------------------------------

    // localStorage("USER") is read exactly once, here, at closure-init time -
    // both readStoredToken() (below) and getUser() need it, and reading it
    // twice (once per function) was pointless duplicate work for a value
    // that can't change between the two reads happening a few lines apart
    // at page load.
    function readStoredUser() {
        const user = localStorage.getItem("USER");
        if (user > '') {
            try {
                return JSON.parse(user);
            } catch (e) {
                console.log(e);
            }
        }
    }

    const storedUser = readStoredUser();
    // Read once, here, and reused by both readStoredToken() (immediately
    // below) and getUser() (further down) - same reasoning as storedUser
    // above, just for the "TOKEN" key instead of "USER".
    const storedTokenValue = localStorage.getItem("TOKEN");

    // TOKEN is the normal home for a session token, but some login responses
    // only ever get stored as part of the "USER" JSON blob (property "token"
    // or "auth_token") - this is the one place that fallback is resolved, so
    // every other reader (headers(), registerPasskey()) just sees a correct
    // `token` closure variable regardless of which of the two it came from.
    function readStoredToken() {
        return storedTokenValue || (storedUser && (storedUser.token || storedUser.auth_token));
    }

    let token = readStoredToken();
    let lang = document.documentElement.lang.split(",")[0] || 'en';

    // Fixed: this used to bail out (`if (token) return;`) whenever a token
    // was already present - which is the NORMAL case on every page reload
    // once TOKEN is set, not just the first visit - so ensureUser() below
    // set `userStruct` to undefined on every reload despite a perfectly
    // valid session existing, and `Auth.userStruct` never reflected who was
    // actually logged in. Now it always returns the already-known user (the
    // one cached `storedUser`, no second localStorage read) once there is
    // one. saveUser()'s full re-render/theme/lang cascade still only runs
    // for the genuine self-heal case - TOKEN was missing and this came
    // solely from the USER blob's own fallback property - not on every
    // ordinary reload where TOKEN was already there and the page
    // (server-rendered) already reflects it.
    function getUser() {
        if (userStruct) return userStruct;
        if (!storedUser) return;

        if (!storedTokenValue) {
            saveUser(storedUser);
        }
        return storedUser;
    }

    // Only called from App.init() (via `Auth.ensureUser()`, wired through
    // jQuery's document-ready) - NOT eagerly here at closure-construction
    // time any more. `Auth`'s IIFE runs the instant the browser executes
    // user.js's <script> tag, synchronously, mid-parse - if htmx.min.js
    // happens to be a later <script> tag on the page, `window.htmx` doesn't
    // exist yet at that point. That's fine as long as nothing here needs it
    // - but the self-heal branch of getUser() below calls saveUser(), which
    // calls changeLang(), which now calls htmx.ajax() (it used to be
    // $.ajax(), and jQuery is reliably loaded first) - so an eager call
    // here threw "Can't find variable: htmx" whenever htmx.min.js's
    // <script> tag came after user.js's/app.js's in the page. App.init()
    // itself only runs on document-ready, which fires after every
    // synchronous <script> tag on the page has executed regardless of
    // their relative order, so calling it from there instead removes the
    // ordering dependency entirely - no need to move or reorder any
    // <script> tags.
    function ensureUser() {
        if (!userStruct) userStruct = getUser();
        return userStruct;
    }

// handling response login form, save users data & render some properties
function afterLogin(userData, thisForm) {
    if (!userData) {
        showMessage(thisForm, "Need users data!");
        console.log(thisForm)
        return false;
    }

    localStorage.setItem("USER", JSON.stringify(userData));
    saveUser(userData);
    userStruct = userData;
    $('input[autofocus]:last').focus();

    return true;
}

    // headers() is the single source of truth for the two headers every
    // request needs: observer.js's htmx:configRequest handler calls this for
    // every htmx request, and forms.js's sendFile() (the one remaining raw
    // XHR upload) calls it directly - instead of each keeping their own
    // copy of "read token/lang, build these two headers", as three separate
    // places (this, over_click.js's now-removed getHeaders(), and a couple
    // of $.ajaxSetup() calls) used to.
    function headers() {
        const h = {'Accept-Language': lang};
        if (token) {
            h['Authorization'] = 'Bearer ' + token;
        }
        return h;
    }

function changeLang(newLang) {
    if (lang.includes(newLang)) {
        return false
    }

    lang = newLang;
    // htmx.ajax(), not $.ajax() - picks up Accept-Language (the whole
    // point of this call) from Auth.headers() via observer.js's
    // htmx:configRequest, same as every other request in the app now.
    htmx.ajax('GET', '/top_menu', {
        swap: 'none',
        handler: (elt, info) => $('.topline-navbar').html(info.xhr.responseText)
    });
    htmx.ajax('GET', '/foot_menu', {
        swap: 'none',
        handler: (elt, info) => $('.footer-mnu').html(info.xhr.responseText)
    });
    loadContent(document.location.href.replace(/lang=\d+/, ``));
    return false
}

    // htmx:afterRequest data-success handler for the login form, e.g.
    // <form ... data-success="Auth.SaveUser"> (or the global SaveUser alias)
    function saveUserFromEvent(evt) {
        let userData = JSON.parse(evt.detail.xhr.responseText);
    return saveUser(userData)
}

function saveUser(userData) {
    var userSuffix = userData.lang ? `(${userData.lang})` : '';
    token = userData.token || userData.access_token || userData.bearer_token || userData.auth_token;
    localStorage.setItem("TOKEN", token);
    // Fixed vs. the pre-refactor version, which passed userData here
    // directly: localStorage.setItem coerces a plain object to the
    // string "[object Object]", not JSON - so a page reload's
    // getUser() -> JSON.parse(...) would throw on the very next visit.
    localStorage.setItem("USER", JSON.stringify(userData));

    $('#sUser').text(userData.name + userSuffix);
    $('body').attr('auth', true);
    changeLang(userData.lang);
    // No $.ajaxSetup({beforeSend: getHeaders}) here anymore - every
    // request in the app now either goes through htmx.ajax()/hx-* (and
    // gets these headers via htmx:configRequest -> Auth.headers()) or,
    // for the one raw XHR left (forms.js's sendFile upload), calls
    // Auth.headers() directly. A global ajaxSetup default was a third,
    // redundant way of doing the same thing for call sites that no
    // longer exist as raw $.ajax calls at all.

    if (userData.theme) {
        // call custom theme changer
        if (typeof CustomChangeTheme !== "undefined") {
            CustomChangeTheme(userData.theme);
        } else {
            // use default
            changeTheme(userData.theme);
        }
    }

    // Fixed vs. the pre-refactor version: `urlAfterLogin.onsubmit` was
    // always undefined - urlAfterLogin is a string, never a form/element
    // reference - so that branch could never run. What's left is exactly
    // the one live behavior it guarded: fall back to the server's own
    // suggested landing page when nothing else already set one.
    if (urlAfterLogin === '' && userData.formActions !== undefined) {
            urlAfterLogin = userData.formActions[0].url;
        }
    if (urlAfterLogin > '') {
        loadContent(urlAfterLogin);
        urlAfterLogin = '';
    }
    // No-op when this login wasn't a relogin() - completeReauth() only
    // does anything when beginReauth() actually ran.
    completeReauth();

    return false
}

    function changeTheme(id_themes) {
    LoadCSS(`/themes/?id=${id_themes}`, true,
        function (data, status, xhr) {
            switch (xhr.status) {
                case 201:
                case 204:
                case 200: {
                    let theme = xhr.getResponseHeader('Section-Name');
                    LoadStyles(theme, data);
                    $('body').attr('theme', theme);
                    return
                }
                default:
                    console.log(`themes load fail: ${status}, ${data}`)
            }
        });

    return false
}

function logOut(elem) {
    if (!confirm(`Do you sure to logout?`)) {
        return false;
    }
    token = null;
    $('#sUser').text('');
    localStorage.removeItem("USER");
    localStorage.removeItem("TOKEN");
    $('body').removeAttr('auth');
    loadContent(elem.href);

    return false;
}

    // --- Re-authentication: a generic, htmx-driven replacement for the old
    // hard-coded relogin(url)/App.relogin ------------------------------------
    //
    // beginReauth()/completeReauth() gate everything on a single shared
    // promise: beginReauth() opens it (idempotent - a second 401 arriving
    // before the user finishes logging in just shares the same promise),
    // completeReauth() (called from saveUser() above) resolves it and
    // replays whatever was queued while it was open. onceReady() is what
    // observer.js's htmx:confirm listener awaits to hold a *new* request
    // until that resolution happens, instead of letting it fail too.
    function isReauthInProgress() {
        return reauthPromise !== null;
    }

    function beginReauth() {
        if (!reauthPromise) {
            reauthPromise = new Promise(resolve => {
                reauthResolve = resolve;
            });
            // Escape valve: without this, an abandoned/unreachable relogin
            // (the login UI never actually shown, or the user just never
            // finishes it) leaves reauthPromise pending forever - and with
            // it, observer.js's htmx:confirm gate silently holds *every*
            // future request on the page, including the login form's own
            // submit, awaiting a promise nothing will ever resolve. 30s is
            // generous for someone actually mid-login but short enough that
            // a wedged/abandoned attempt can't brick the app until a hard
            // refresh. Giving up here does NOT replay the queued requests
            // (still no valid session to run them with - that would just
            // re-401 and re-queue), it only releases everything that was
            // being held so the page is usable again.
            reauthTimeout = setTimeout(() => {
                console.warn('Auth: relogin was never completed - releasing requests held for it without replaying them.');
                // reauthResolve, not the executor's own `resolve` param -
                // that name only lives inside the Promise executor above,
                // out of scope here. reauthResolve is the same function,
                // just reached through the closure variable it was stashed
                // in a few lines up.
                const give_up = reauthResolve;
                reauthPromise = null;
                reauthResolve = null;
                pendingRequests = [];
                give_up();
            }, 30000);
        }
        return reauthPromise;
    }

    function onceReady() {
        return reauthPromise || Promise.resolve();
    }

    function completeReauth() {
        if (!reauthResolve) return;
        clearTimeout(reauthTimeout);
        const resolve = reauthResolve;
        reauthPromise = null;
        reauthResolve = null;
        resolve();
        replayPendingRequests();
    }

    // Read back whatever hx-get/hx-post/.../href a failed element actually
    // carries - a fallback source for verb/path, used only when the event
    // itself didn't already give us one (see queuePendingRequest() below).
    // This is the same information htmx itself would use to issue the
    // request, just read straight from the DOM instead of an htmx internal.
    function resolveRequest(elt) {
        for (const verb of ['get', 'post', 'put', 'patch', 'delete']) {
            const path = elt.getAttribute && elt.getAttribute(`hx-${verb}`);
            if (path) return {verb: verb.toUpperCase(), path};
        }
        if (elt.tagName === 'A' && elt.getAttribute('href')) {
            return {verb: 'GET', path: elt.getAttribute('href')};
        }
        return null;
    }

    // Every failed request goes on this queue, not just the ones with a
    // specific triggering element - previously, anything without a usable
    // `elt` (loadContent()/appendTable()/loadTableWithOrder(), all of which
    // call htmx.ajax() with no `source`, so `elt` defaults to document.body)
    // fell back to overwriting a single `urlAfterLogin` string. That loses
    // every failure but the last one when more than one such request fails
    // around the same time (e.g. a page load and a background table poll
    // both hitting an expired session together) - this queue keeps all of
    // them instead of just the most recent.
    function queuePendingRequest(evt) {
        if (!evt || !evt.detail) return;
        const elt = evt.detail.elt;
        if (elt && (elt.id === 'fLogin' || elt.closest?.('#fLogin'))) {
            return; // never requeue the login form's own submit
        }

        const xhr = evt.detail.xhr;
        // requestConfig is a documented part of htmx:responseError/
        // htmx:onLoadError's event detail, but its own internal field names
        // aren't part of htmx's public docs - read defensively, in the
        // shape observed on this app's own htmx version (evt.detail.path/
        // pathInfo.requestPath/verb are already read the same way in
        // htmx:configRequest above), and fall back to the DOM/xhr when a
        // field isn't there.
        const requestConfig = evt.detail.requestConfig || {};
        const verb = (requestConfig.verb
            || (elt && elt !== document.body && resolveRequest(elt)?.verb)
            || 'GET').toUpperCase();
        const path = requestConfig.path
            || requestConfig.pathInfo?.requestPath
            || (elt && elt !== document.body && resolveRequest(elt)?.path)
            || (xhr && xhr.responseURL);

        if (!path) return; // nothing usable to replay

        pendingRequests.push({
            elt: (elt && elt !== document.body) ? elt : null,
            target: evt.detail.target || null,
            verb,
            path,
        });
    }

    function replayPendingRequests() {
        const queued = pendingRequests;
        pendingRequests = [];
        queued.forEach(({elt, target, verb, path}) => {
            if (elt && elt.isConnected) {
                if (elt.tagName === 'FORM') {
                    htmx.trigger(elt, 'submit');
                    return;
                }
                htmx.ajax(verb, path, {source: elt, target: target || elt});
                return;
            }
            // No triggering element (or it's no longer in the DOM) - this
            // was a whole-page/background fetch (loadContent(),
            // appendTable(), loadTableWithOrder()...). loadContent() redoes
            // exactly the same multi-region content load any of those would
            // have wanted on recovery; anything that isn't a plain GET
            // replays as a targetless htmx.ajax() call instead.
            if (verb === 'GET') {
                loadContent(path);
            } else {
                htmx.ajax(verb, path, {target: target || document.body});
            }
        });
    }

    // The one entry point every 401 now goes through - a plain
    // htmx:responseError/htmx:onLoadError event handed straight in by
    // observer.js, whether it came from a form, a plain hx-* element, or a
    // whole-page App.loadContent() fetch (which has no triggering element at
    // all, just document.body).
    function relogin(evt) {
        if (evt && evt.preventDefault) evt.preventDefault();

        token = null;
        localStorage.removeItem("TOKEN");

        queuePendingRequest(evt);

        beginReauth();
        htmx.trigger('#bLogin', 'click');
        return false;
    }

    // --- Passkey (WebAuthn) - the PRIMARY sign-in path ----------------------
    // (moved in from passkey-auth.js unchanged; base64url<->ArrayBuffer glue
    // and the two wire-format converters stay private, nothing outside Auth
    // ever needs them)

    function isPasskeySupported() {
        return typeof window.PublicKeyCredential !== 'undefined';
    }

    // Optional: surface whether this device can do a *platform* passkey
    // (Face ID/Touch ID/Windows Hello) vs. only a security key, so the UI
    // can say "Use Face ID" instead of a generic "Use a passkey".
    async function isPlatformAuthenticatorAvailable() {
        if (!isPasskeySupported() || !window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable) {
            return false;
        }
        try {
            return await window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
        } catch (e) {
            console.log(e);
            return false;
        }
    }

    function base64urlToBuffer(base64url) {
        const padded = base64url.replace(/-/g, '+').replace(/_/g, '/')
            .padEnd(base64url.length + (4 - base64url.length % 4) % 4, '=');
        const raw = atob(padded);
        const buffer = new Uint8Array(raw.length);
        for (let i = 0; i < raw.length; i++) buffer[i] = raw.charCodeAt(i);
        return buffer.buffer;
    }

    function bufferToBase64url(buffer) {
        const bytes = new Uint8Array(buffer);
        let str = '';
        for (const b of bytes) str += String.fromCharCode(b);
        return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
    }

    // Recursively walk a PublicKeyCredentialCreationOptions/RequestOptions
    // JSON blob and turn the well-known base64url fields into ArrayBuffers
    // in place. Pass body.publicKey, not the raw response body - go-webauthn's
    // CredentialCreation/CredentialAssertion both marshal the actual options
    // under a "publicKey" key (see decode call sites below).
    function decodeCredentialOptions(options) {
        if (options.challenge) options.challenge = base64urlToBuffer(options.challenge);
        if (options.user?.id) options.user.id = base64urlToBuffer(options.user.id);
        for (const listName of ['excludeCredentials', 'allowCredentials']) {
            options[listName]?.forEach(cred => {
                cred.id = base64urlToBuffer(cred.id);
            });
        }
        return options;
    }

    // Turn a PublicKeyCredential response into the plain JSON shape the
    // server (go-webauthn's protocol.ParseCredentialCreationResponseBytes /
    // ParseCredentialRequestResponseBytes) expects.
    function encodeCredential(credential) {
        const base = {
            id: credential.id,
            rawId: bufferToBase64url(credential.rawId),
            type: credential.type,
            clientExtensionResults: credential.getClientExtensionResults(),
        };
        if (credential.response.attestationObject) {
            // registration
            return {
                ...base,
                response: {
                    clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
                    attestationObject: bufferToBase64url(credential.response.attestationObject),
                }
            };
        }
        // login
        return {
            ...base,
            response: {
                clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
                authenticatorData: bufferToBase64url(credential.response.authenticatorData),
                signature: bufferToBase64url(credential.response.signature),
                userHandle: credential.response.userHandle ? bufferToBase64url(credential.response.userHandle) : null,
            }
        };
    }

    // --- Registration: "Add a passkey" button in account settings ----------
    // Requires the user already be logged in (token set) - registering a
    // brand new account via passkey-only is a separate, simpler flow (no
    // existing user/session to attach the credential to).
    async function registerPasskey() {
        if (!isPasskeySupported()) {
            showMessage(null, 'Passkeys are not supported in this browser.');
            return false;
        }

        try {
            const beginResp = await fetch('/webauthn/register/begin', {
                method: 'POST',
                credentials: 'same-origin', // sends/receives the px_session cookie
                headers: {'Authorization': 'Bearer ' + token, 'Accept': 'application/json'},
            });
            if (!beginResp.ok) throw new Error(`register/begin failed: ${beginResp.status}`);
            const options = decodeCredentialOptions((await beginResp.json()).publicKey);

            const credential = await navigator.credentials.create({publicKey: options});

            const finishResp = await fetch('/webauthn/register/finish', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'Authorization': 'Bearer ' + token,
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(encodeCredential(credential)),
            });
            const data = await finishResp.json();
            if (!finishResp.ok) {
                showMessage(null, data.error || 'Could not save passkey.');
                return false;
            }
            afterSaveAnyForm(data, 'Success');
            return true;
        } catch (e) {
            console.error(e);
            showMessage(null, 'Passkey registration was cancelled or failed: ' + e.message);
            return false;
        }
    }

    // --- Login: the primary sign-in path ------------------------------------
    // `loginValue` is whatever identifies the account (email/username) - the
    // server needs it in login/begin to look up that user's registered
    // credential ids so the browser only offers a matching passkey. It's NOT
    // repeated in the finish call: the server already tied that user to this
    // ceremony's px_session cookie at begin time, so finish only needs the
    // credential.
    async function loginWithPasskey(loginValue) {
        if (!isPasskeySupported()) {
            return false;
        }

        try {
            const beginResp = await fetch('/webauthn/login/begin', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {'Accept': 'application/json', 'Content-Type': 'application/json'},
                body: JSON.stringify({login: loginValue}),
            });
            if (!beginResp.ok) throw new Error(`login/begin failed: ${beginResp.status}`);
            const options = decodeCredentialOptions((await beginResp.json()).publicKey);

            const credential = await navigator.credentials.get({publicKey: options});

            const finishResp = await fetch('/webauthn/login/finish', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {'Accept': 'application/json', 'Content-Type': 'application/json'},
                body: JSON.stringify(encodeCredential(credential)),
            });
            const userData = await finishResp.json();
            if (!finishResp.ok) {
                showMessage(null, userData.error || 'Sign-in failed.');
                return false;
            }
            // Same response shape a password login already produces, so it
            // plugs straight into the existing session bootstrap.
            return afterLogin(userData);
        } catch (e) {
            console.error(e);
            // A cancelled/failed platform prompt (e.g. Face ID declined)
            // falls back to the password form silently rather than alert() -
            // the browser's own biometric UI already told the user what
            // happened.
            console.log('Passkey sign-in was cancelled or failed: ' + e.message);
            return false;
        }
    }

    // Optional standalone wiring for a "Use Face ID / Touch ID" button placed
    // next to the password field, kept for callers that want an explicit,
    // user-initiated button rather than the automatic try-first behavior of
    // login() below.
    async function tryPasskeyLoginFromForm(thisForm) {
        const loginValue = thisForm.querySelector('[name=login], [name=email], [name=username]')?.value;
        if (!loginValue) {
            showMessage(thisForm, 'Enter your email/username first, then choose "Use Face ID / Touch ID".');
            return false;
        }
        return loginWithPasskey(loginValue);
    }

    // login is the ONE entry point a login form needs. It no longer needs
    // wiring in the template at all: observer.js's cfgHTMX() binds a
    // capture-phase 'submit' listener on document.body that calls
    // Auth.login(evt, form) for any submit whose form's action is the real
    // sign-in endpoint (/user/signin/, matched the same way isSignInForm()/
    // isSignInRequest() do elsewhere) - so don't ALSO add
    // onsubmit="return Auth.login(event, this)" to that form's markup, that
    // would call this twice for the same submit (two concurrent
    // navigator.credentials.get() calls racing each other). This function
    // itself is unchanged and still usable directly for a form matched some
    // other way.
    //
    // It always tries a passkey first; if passkeys aren't supported, nothing
    // has been typed yet, or the ceremony is declined/fails, it falls back
    // to the form's own normal submit (already hx-boosted, per
    // <body hx-boost="true"> - see html.qtpl) - exactly once, via a one-shot
    // data attribute that guards against re-entering this same handler for
    // the fallback submit it triggers (including the re-entry from
    // observer.js's own delegated listener above, which sees that retriggered
    // submit too).
    //
    // Because this function is `async`, a direct `onsubmit="return
    // Auth.login(...)"` wiring would always see a Promise back, which the
    // browser treats as truthy (never "return false") - so this does NOT
    // rely on that return value to stop the native submit either way. It
    // calls event.preventDefault() itself, and only when it's actually
    // taking over (attempting a passkey); the synchronous portion of an
    // async function runs before its first `await`, so that
    // preventDefault() still lands inside the same submit-event dispatch -
    // which is also why observer.js's listener has to be capture-phase, to
    // run before htmx's own submit listener on the form itself.
    async function login(event, thisForm) {
        thisForm = thisForm || event.target;

        if (thisForm.dataset.passkeyTried) {
            // We already tried a passkey for this submit and it declined/
            // failed - this is our own re-triggered fallback submit, let it
            // through as an ordinary password POST.
            delete thisForm.dataset.passkeyTried;
            return true;
        }

        const loginField = thisForm.querySelector('[name=login], [name=email], [name=username]');
        const loginValue = loginField?.value;
        if (!isPasskeySupported() || !loginValue) {
            return true;
        }

        event.preventDefault();

        const ok = await loginWithPasskey(loginValue);
        if (!ok) {
            // Passkey unsupported/declined/failed - fall back to the form's
            // own password submit, exactly once.
            thisForm.dataset.passkeyTried = '1';
            htmx.trigger(thisForm, 'submit');
        }
        return false;
    }

    // No eager ensureUser() call here any more - see the comment on
    // ensureUser() above. App.init() (app.js, wired via `$(App.init)`) is
    // the one and only place it's called now.

    return {
        getUser,
        ensureUser,
        afterLogin,
        SaveUser: saveUserFromEvent,
        saveUser,
        changeLang,
        ChangeTheme: changeTheme,
        logOut,
        headers,
        relogin,
        isReauthInProgress,
        onceReady,
        isPasskeySupported,
        isPlatformAuthenticatorAvailable,
        registerPasskey,
        loginWithPasskey,
        tryPasskeyLoginFromForm,
        login,
        get token() {
            return token;
        },
        set token(v) {
            token = v;
        },
        get lang() {
            return lang;
        },
        set lang(v) {
            lang = v;
        },
        get userStruct() {
            return userStruct;
        },
        get urlAfterLogin() {
            return urlAfterLogin;
        },
        set urlAfterLogin(v) {
            urlAfterLogin = v;
        },
    };
})();

// --- Backward-compatible global aliases ------------------------------------
// Same names as the pre-refactor user.js/passkey-auth.js - anything that
// still calls one of these bare (a data-success="afterLogin" attribute,
// generated HTML's onclick="return logOut(this)", etc.) keeps working with
// zero changes. New code should prefer Auth.* directly.
function getUser() {
    return Auth.getUser();
}

function afterLogin(userData, thisForm) {
    return Auth.afterLogin(userData, thisForm);
}

function SaveUser(event) {
    return Auth.SaveUser(event);
}

function saveUser(userData) {
    return Auth.saveUser(userData);
}

function changeLang(newLang) {
    return Auth.changeLang(newLang);
}

function ChangeTheme(id_themes) {
    return Auth.ChangeTheme(id_themes);
}

function logOut(elem) {
    return Auth.logOut(elem);
}

// Replaces app.js's old `function relogin(url) { return App.relogin(url); }`
// - relogin now takes the failing htmx event itself, not a bare URL, so it
// can tell what actually needs replaying. See Auth.relogin() above.
function relogin(evt) {
    return Auth.relogin(evt);
}

function isPasskeySupported() {
    return Auth.isPasskeySupported();
}

function isPlatformAuthenticatorAvailable() {
    return Auth.isPlatformAuthenticatorAvailable();
}

function registerPasskey() {
    return Auth.registerPasskey();
}

function loginWithPasskey(loginValue) {
    return Auth.loginWithPasskey(loginValue);
}

function tryPasskeyLoginFromForm(thisForm) {
    return Auth.tryPasskeyLoginFromForm(thisForm);
}