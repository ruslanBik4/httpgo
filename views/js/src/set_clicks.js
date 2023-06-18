/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

var isProcess = false;

function setClickAll(event) {
    if (isProcess) {
        return;
    }

    if (event && event.target === '<script>') {
        return
    }
    isProcess = true;

    console.log(event);
    // add onSubmit event instead default behaviourism of form
    $('form:not([onsubmit])').on("submit", function () {
        return saveForm(this);
    });
    // add click event instead default - response will show on div.#content
    $('a[href!="#"]:not([rel]):not(onclick):not([target=_blank])').each(function () {
        this.rel = 'setClickAll';

        $(this).click(OverClick);
    });

    if (!event || event.target === document.getElementById('content')) {
        $('input[autofocus]:last').focus();
    }
    isProcess = false;
}