/*
 * Copyright (c) 2023-2025. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

var isProcess = false;

function isIgnoreTarget(target) {
    return $(target).attr('rel') > "" || $(target).hasClass('htmx-request') || $(target).hasClass('htmx-added') || $(target).hasClass('htmx-indicator')
        // || $(target).hasClass('htmx-settling')
        || $(target).hasClass('htmx-swapping') || $(target).hasClass('htmx-fancybox-container')
        || ($(target).parents('svg, .fancybox-container, .htmx-request, .flatpickr-calendar').length > 0)
}

function setClickAll(target) {
    if (isProcess) {
        return;
    }

    if (target && (target === '<script>'
        || target.localName && (target.localName === 'script' || target.localName === 'tbody'
            || target.localName.startsWith('svg') || target.localName.startsWith('svg') || target.localName.startsWith('th')
            || target.localName.startsWith('output'))
        || (typeof target === 'string' && (target.includes('htmx-') || target.includes('fancybox') || target.includes('flowchart')
            || target.startsWith('<th')
            || target.startsWith('<svg') || target.startsWith('<output')))
        || (typeof target === 'object' && isIgnoreTarget(target))
    )) {
        return
    }
    isProcess = true;
    console.log(target);

    let cfgDate = {
        format: 'YYYY-MM-DD',
        timepicker: false,
        lang: lang
    };
    let cfgDateTime = {
        format: 'Y-m-d H:i:s',
        singleDate: true,
        lang: lang
    };
    let cfgDateRange = {
        setValue: function (s) {
            let elem = $(this);
            elem.val(`[${s}]`);
            setTimeout(function () {
                return elem.on('change');
            }, 1000);
            return false;
        },
        separator: ',',
        autoClose: true,
        startOfWeek: 'sunday',
        singleDate: false,
        shortcuts:
            {
                'prev-days': [1, 3, 5, 7],
                'next-days': [3, 5, 7],
                'prev': ['week', 'month', 'year'],
                'next': ['week', 'month', 'year']
            },
        ...cfgDate
    };

    // target = target || document.getElementsByTagName("body")[0];
    // add onSubmit event instead default behaviourism of form
    $('form:not([onsubmit])', target).on("submit", function () {
        return saveForm(this);
    });

    $('form:not([rel]), .filt-arrow:not([rel])', target).each(
        (ind, elem) => {
            let dates = $('input[type=date-range]:not([rel])', elem);
            dates.flatpickr({
                mode: "range",
                altInput: true, // Allows a different display format
                altFormat: "Y-m-d", // ✅ Shows as a range format
                dateFormat: "[Y-m-d,Y-m-d]", // Flatpickr saves this format
                allowInput: true, // ✅ Allows manual entry
                showMonths: 2,
                // clickOpens: true, // ✅ Ensure calendar still opens
                onClose: function (dates, dateStr, instance) {
                    if (dates.length === 2) {
                        // Format value as [YYYY-MM-DD,YYYY-MM-DD]
                        const formattedValue = `[${instance.formatDate(dates[0], "Y-m-d")},${instance.formatDate(dates[1], "Y-m-d")}]`;
                        instance.input.value = formattedValue;
                        filterTableData(formattedValue, instance.input.dataset.class);
                    } else {
                        // If only one date is selected, clear the input
                        instance.input.value = "";
                    }
                }
            });
            dates.attr('rel', 'datetimepicker');

            dates = $('input[type=datetime]:not([rel])', elem);
            if (dates.length > 0) {
                dates.dateRangePicker(cfgDateTime).attr('rel', 'datetimepicker');
            }
            dates = $('input[type=date]:not([rel])', elem);
            if (dates.length > 0) {
                dates.flatpickr({
                    // mode: "range",
                    dateFormat: "Y-m-d",
                    allowInput: true,
                    onClose: function (dates, dateStr, instance) {
                        if (dates.length === 2) {
                            const formattedValue = `[${dates[0].toISOString().split('T')[0]},${dates[1].toISOString().split('T')[0]}]`;
                            instance.input.value = formattedValue;
                            // instance.input.dispatchEvent(new Event('change', {bubbles: true})); // Trigger onchange
                        }
                    }
                    // dateRangePicker({
                    //     singleDate: true,
                    //     ...cfgDate
                });
                dates.attr('rel', 'datetimepicker');
            }
        })

    let hxEvents = $('[hx-get], [hx-post], [hx-target], [hx-trigger], [hx-on]', target).not('[rel]');
    if (hxEvents.length > 0) {
        hxEvents.attr("rel", 'htmx');
        htmx.process(target);
    }
    // add click event instead default - response will show on div.#content
    $('a[href!="#"]:not([rel]):not([onclick]):not([target=_blank])', target).each(function () {
        this.rel = 'setClickAll';

        $(this).click(OverClick);
    });
    setTextEdit(target);
    setSliderBox(target);

    if (target === document.getElementById('content')) {
        $('input[autofocus]:last').focus();
    }

    isProcess = false;
}

function setSliderBox(target) {
    $('label.input-label > input.slider', target).each(function (ind, elem) {
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

function setTextEdit(target) {
    let textInputs = $('textarea:not([readonly]):not([raw])', target);
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