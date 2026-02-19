/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

function createObserver() {
// Create a MutationObserver instance
    const observer = new MutationObserver((mutationsList, observer) => {
        console.debug("Attributes changed:", mutationsList);
        for (const mutation of mutationsList) {
            if (mutation.type === "childList") {
                mutation.addedNodes.forEach(node => { processAll(node); } );
            } else if (mutation.type === "attributes") {
                let ignores = ["style", "class", "rel"];
                if (ignores.indexOf(mutation.attributeName) > -1) {
                    return
                }
                console.log("Attributes changed:", mutation);
            } else if (mutation.type === "characterData") {
                console.log("Text content changed:", mutation);
            } else {
                console.error("Attributes not changed:", mutation);
            }
        }
    });

// Configure observer options
    const config = {
        childList: true,      // Detect when children are added/removed
        attributes: true,     // Detect attribute changes
        subtree: true,        // Observe all descendants
        characterData: true   // Detect text content changes
    };

    observer.observe(document.body, config);
    console.debug('observer created')
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

function cfgHTMX() {
    document.body.addEventListener('htmx:onLoadError', function (evt) {
        console.log(evt);
        handleError(evt.detail.xhr, evt.detail.xhr.status, evt.detail.xhr.error);
    });

    document.body.addEventListener('htmx:configRequest', function (evt) {
        if (token) {
            evt.detail.headers['Authorization'] = 'Bearer ' + token;
        }
        if (lang) {
            evt.detail.headers['Accept-Language'] = lang;
        }
        evt.detail.headers['X-Requested-With'] = 'XMLHttpRequest';
        evt.detail.headers['HX-Request'] = 'html';
        const pathInfo = evt.detail.path;
        if (pathInfo && pathInfo.finalRequestPath) {
            console.log(pathInfo);
            pathInfo.responsePath = replMacros(pathInfo.finalRequestPath);
        } else {
            console.log(evt);
        }
    });

    document.body.addEventListener('htmx:beforeSwap', evt => {
        if (!evt.detail.elt.attributes.target) {
            if (evt.detail.xhr.status === 204) {
                alert('empty response!');
                evt.preventDefault();
            }
            return;
        }
        let responseText = evt.detail.xhr.responseText;
        switch (evt.detail.elt.attributes.target.nodeValue) {
            case "_modal":
                fancyOpen(responseText);

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

    document.body.addEventListener('htmx:afterRequest', evt => {
        var $out = $('#content');
        if (evt.target.localName === 'form') {
            $.fancybox.close();
        }
        const xhr = evt.detail.xhr;
        switch (xhr.status) {
            case 204:
                evt.preventDefault();
                alert('empty response!');
                return false;

            case 206:
                readEvents($out, data);
                return false;

            case 201:
                alert(`Successful add record #${xhr.responseText}`);
            case 200:
                if (evt.target.matches("[data-fancybox]")) {
                    fancyOpen(evt.detail.xhr.response);
                    return;
                }
                htmx.trigger(evt.detail.target, 'on_change');
                processAll(evt.detail.target);
        }
    });

    document.body.addEventListener('htmx:responseError', evt => {
        const xhr = evt.detail.xhr;
        switch (xhr.status) {
            case 401:
                evt.preventDefault();
                return relogin(xhr.responseURL);
            case 400:
                // stop the regular request from being issued
                evt.preventDefault();
                const src = evt.detail.elt;
                let obj = JSON.parse(xhr.responseText);
                showErrors(obj.formErrors, src)
        }
    });
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
