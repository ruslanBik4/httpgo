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

function setTextEdit() {
    textInputs = $('textarea:not([readonly])');
    if (textInputs.length > 0) {
        let scripts = Array
            .from(document.querySelectorAll('script'))
            .map(scr => scr.src);

        if (!scripts.includes('tinymce.min')) {
            LoadJScript("https://cdn.tiny.cloud/1/2os6bponsl87x9zsso916jquzsi298ckurhmnf7fp9scvpgt/tinymce/6/tinymce.min.js", false, true)
        }

        textInputs.focus(
            function (event) {
                let name = event.target.name;
                tinymce.init({
                    target: event.target,
                    menubar: false,
                    plugins: 'anchor autolink charmap codesample emoticons image link lists media searchreplace table visualblocks wordcount    ',
                    toolbar: 'undo redo | blocks fontfamily fontsize | bold italic underline strikethrough | link image media table mergetags | addcomment showcomments | spellcheckdialog a11ycheck | align lineheight | numlist bullist indent outdent  | removeformat',
                    mergetags_list: [
                        {value: "name", title: name},
                        {value: 'placeholder', title: name},
                    ],
                    setup: (editor) => {
                        editor.on('input', (e) => {
                            console.log(e);
                            // $('#{%s idShake %}_form button.hidden').removeClass('hidden').addClass('main-btn');
                        });

                        editor.on('focusout', (e) => {
                            $('textarea[name="' + name + '"]').text(editor.getContent({format: 'text'}));
                        });
                        editor.focus();
                    }
                });
            });
    }
}