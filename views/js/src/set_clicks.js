/*
 * Copyright (c) 2023-2024. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

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
    let cfgDate = {
        format: 'YYYY-MM-DD',
        timepicker: false,
        lang: lang
    };
    let cfgDateTime = {
        format: 'Y-m-d H:i:s',
        lang: lang
    };
    let cfgDateRange = {
        setValue: function (s) {
            $(this).val(`[${s}]`);
            // if (value > '') {
            //     $input.val(value + ',]')
            // } else {
            //     $input.val('['+value)
            // }
            console.log(s);
            return false;
        },
        separator: ',',
        autoClose: true,
        ...cfgDate
    };
    // add onSubmit event instead default behaviourism of form
    $('form:not([onsubmit])').on("submit", function () {
        $('input[type=datetime]:not([rel])', this).datetimepicker(cfgDateTime).attr('rel', 'datetimepicker');
        $('input[type=date]:not([rel])', this).datetimepicker(cfgDate).attr('rel', 'datetimepicker');
        // $('input[type=date-range]:not([rel])', this).dateRangePicker(cfgDateRange).attr('rel', 'datetimepicker');
        return saveForm(this);
    });

    $('form input[type=datetime]:not([rel])').datetimepicker(cfgDateTime).attr('rel', 'datetimepicker');
    $('form input[type=date]:not([rel])').datetimepicker(cfgDate).attr('rel', 'datetimepicker');
    // $('form input[type=date-range]:not([rel])').dateRangePicker(cfgDateRange).attr('rel', 'datetimepicker');

    $('.filt-arrow input[type=datetime]:not([rel])').datetimepicker(cfgDateTime).attr('rel', 'datetimepicker');
    $('.filt-arrow input[type=date]:not([rel])').datetimepicker(cfgDate).attr('rel', 'datetimepicker');
    let dates = $('.filt-arrow input[type=date-range]:not([rel])');
    if (dates.length > 0) {
        dates.dateRangePicker(cfgDateRange).attr('rel', 'datetimepicker');
    }
    dates = $('form input[type=date-range]:not([rel])');
    if (dates.length > 0) {
        dates.dateRangePicker(cfgDateRange).attr('rel', 'datetimepicker');
    }

    // add click event instead default - response will show on div.#content
    $('a[href!="#"]:not([rel]):not([onclick]):not([target=_blank])').each(function () {
        this.rel = 'setClickAll';

        $(this).click(OverClick);
    });
    setTextEdit();
    setSliderBox();

    if (!event || event.target === document.getElementById('content')) {
        $('input[autofocus]:last').focus();
    }
    isProcess = false;
}

function setSliderBox() {
    $('label.input-label > input.slider').each(function (ind, elem) {
        let values = elem.value.split("-");
        $(elem).parent('label').children('div.slider').slider({
            step: parseFloat(elem.step),
            range: true,
            min: parseFloat(elem.min),
            max: parseFloat(elem.max),
            values: values,
            slide: function (event, ui) {
                elem.value = `[${ui.values[0]} - ${ui.values[1]}]`;
                $(elem).next().attr('data-value', elem.value);
            }
        });
    }).removeClass('slider');
}

function setTextEdit() {
    let textInputs = $('textarea:not([readonly]):not([raw])');
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
                let isNotRaw = !event.target.attributes['raw'];
                let plugins = isNotRaw ? 'anchor autolink charmap codesample emoticons image link lists media searchreplace table visualblocks wordcount' : '';
                tinymce.init({
                    target: event.target,
                    menubar: false,
                    auto_focus: event.target.id,
                    highlight_on_focus: true,
                    plugins: plugins,
                    toolbar: 'undo redo | blocks fontfamily fontsize | bold italic underline strikethrough | link image media table mergetags | addcomment showcomments | spellcheckdialog a11ycheck | align lineheight | numlist bullist indent outdent  | removeformat',
                    mergetags_list: [
                        {value: "name", title: name},
                        {value: 'placeholder', title: name},
                    ],
                    setup: (editor) => {
                        editor.on('input', (e) => {
                            return FormIsModified(event, $(event.target).parents('form'));
                        });

                        editor.on('blur', (e) => {
                            if (isNotRaw) {
                                $(`textarea[name="${name}"]`).text(editor.getContent());
                            } else {
                                let content = editor.getContent({format: 'text'});
                                $(`textarea[name="${name}"]`).text(content);
                                console.log(content);
                                editor.setContent(content);
                            }
                            editor.hide();
                        });
                    }
                });
            });
    }
}

function showMap(val) {
    fancyOpen('<div id="map" class="map_showing"></div>');
    var map = L.map('map');
    var marker;
    map.on('load', function onMapClick(e) {
        marker = L.marker(map.getCenter(), {draggable: true}).addTo(map);
        marker.bindPopup("<b>Hello world!</b><br>I am her.").openPopup();
        // on('move' , function () {
        //     marker.savePoint();
        //     FormIsModified(event, elem.parents('form'));
        // });

        // marker.savePoint = function() {
        //     marker.bindPopup("It;s my new place.").openPopup();
        //     let geo = marker.getLatLng();
        //     elem.val(`(${geo.lat},${geo.lng})`);
        // };
    });


    if (val > "") {
        let arr = val.match(/\((\d*\.\d*)\s*,\s*(\d*\.\d*)\)/);
        map.setView([arr[1], arr[2]], 13);
    } else {
        map.locate({setView: true, maxZoom: 16});
    }

    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    }).addTo(map);

}