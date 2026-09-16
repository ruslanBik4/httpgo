/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
"use strict";

function isPasskeySupported() {
    return typeof window.PublicKeyCredential !== 'undefined';
}

// Optional: surface whether this device can do a *platform* passkey
// (Face ID/Touch ID/Windows Hello) vs. only a security key, so the UI can
// say "Use Face ID" instead of a generic "Use a passkey".
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

// --- base64url <-> ArrayBuffer helpers -------------------------------------
// go-webauthn (like every WebAuthn server library) sends/expects challenge
// and credential ids as base64url strings; the browser API wants
// ArrayBuffers. These two are the only glue code needed either direction.

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

// Recursively walk a PublicKeyCredentialCreationOptions/RequestOptions JSON
// blob and turn the well-known base64url fields into ArrayBuffers in place.
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

// Turn a PublicKeyCredential response into the plain JSON shape the server
// (go-webauthn's ParseCredentialCreationResponseBody/ParseCredentialRequestResponseBody) expects.
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

// --- Registration: "Add a passkey" button in account settings -------------
// Requires the user already be logged in (token set) - registering a brand
// new account via passkey-only is a separate, simpler flow (no existing
// user/session to attach the credential to).
async function registerPasskey() {
    if (!isPasskeySupported()) {
        alert('Passkeys are not supported in this browser.');
        return false;
    }

    try {
        const beginResp = await fetch('/webauthn/register/begin', {
            method: 'POST',
            headers: {'Authorization': 'Bearer ' + token, 'Accept': 'application/json'},
        });
        if (!beginResp.ok) throw new Error(`register/begin failed: ${beginResp.status}`);
        const options = decodeCredentialOptions(await beginResp.json());

        const credential = await navigator.credentials.create({publicKey: options});

        const finishResp = await fetch('/webauthn/register/finish', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(encodeCredential(credential)),
        });
        const data = await finishResp.json();
        if (!finishResp.ok) {
            alert(data.error || 'Could not save passkey.');
            return false;
        }
        afterSaveAnyForm(data, 'Success');
        return true;
    } catch (e) {
        console.error(e);
        alert('Passkey registration was cancelled or failed: ' + e.message);
        return false;
    }
}

// --- Login: replaces (or sits next to) the password form ------------------
// `login` is whatever identifies the account (email/username) - go-webauthn
// needs it in login/begin to look up that user's registered credential ids
// so the browser only offers a matching passkey.
async function loginWithPasskey(login) {
    if (!isPasskeySupported()) {
        alert('Passkeys are not supported in this browser.');
        return false;
    }

    try {
        const beginResp = await fetch('/webauthn/login/begin', {
            method: 'POST',
            headers: {'Accept': 'application/json', 'Content-Type': 'application/json'},
            body: JSON.stringify({login}),
        });
        if (!beginResp.ok) throw new Error(`login/begin failed: ${beginResp.status}`);
        const options = decodeCredentialOptions(await beginResp.json());

        const credential = await navigator.credentials.get({publicKey: options});

        const finishResp = await fetch('/webauthn/login/finish', {
            method: 'POST',
            headers: {'Accept': 'application/json', 'Content-Type': 'application/json'},
            body: JSON.stringify({login, ...encodeCredential(credential)}),
        });
        const userData = await finishResp.json();
        if (!finishResp.ok) {
            alert(userData.error || 'Sign-in failed.');
            return false;
        }
        // Same response shape the password-login form already produces, so
        // it plugs straight into the existing session bootstrap.
        return afterLogin(userData);
    } catch (e) {
        console.error(e);
        // A cancelled/failed platform prompt (e.g. Face ID declined) should
        // fall back to the password form silently rather than alert() —
        // the browser's own biometric UI already told the user what happened.
        console.log('Passkey sign-in was cancelled or failed: ' + e.message);
        return false;
    }
}

// Convenience wiring for a login form: try the passkey first (if the browser
// supports it and the user has typed a login), fall back to the normal
// hx-post submit otherwise. Wire this to the login input's a "Use Face ID /
// Touch ID" button placed next to the password field, not to the form's
// submit itself, since a passkey attempt is optional/best-effort.
async function tryPasskeyLoginFromForm(thisForm) {
    const login = thisForm.querySelector('[name=login], [name=email], [name=username]')?.value;
    if (!login) {
        alert('Enter your email/username first, then choose "Use Face ID / Touch ID".');
        return false;
    }
    return loginWithPasskey(login);
}