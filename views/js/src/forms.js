/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */

"use strict";

function saveForm(thisForm, successFunction, errorFunction) {
    let title = thisForm.name || $('h2', thisForm).text() || $('figcaption', thisForm).text() || thisForm.id,
        nav = getFormNav(thisForm);
    // hidden fields of not used blocks
    $('form figure:hidden:not([validated]) .input-label').children('input, select').attr('disabled', true);

    if (!validateFields(thisForm))
        return false;

    if (thisForm.noValidate && !confirm(`Do you sure to send form "${title}"?`)) {
        return false
    }
    // TODO: create element form for output form result
    var $out = $('output', thisForm),
        $loading = $('.loading', thisForm),
        $progress = $('progress', thisForm);

    $(thisForm).ajaxSubmit({
        beforeSubmit: function (a, f, o) {
            o.dataType = "json";

            // rm field without values
            var isNewRecord = $('input[name=id]').length === 0;

            for (var i = a.length - 1; i >= 0; --i) {
                if (a[i].readOnly
                    || ((a[i].value === '') && (isNewRecord || a[i].type === 'select-one' || a[i].type === 'file'))
                    || (a[i].value.length === 0)) {
                    let t = a.splice(i, 1);
                    console.log(t);
                }
            }

            $("input[type=checkbox][checked]:not(:checked)", f).each(function () {
                a.push({name: this.name, value: 0, type: this.type, required: this.required});
            });
            a.push({name: "is_get_form_actions", value: true, type: "boolean"});
            $out.html('Start sending...');
            $('.errorLabel').hide();
            $progress.show();
            $loading.show();
        },
        beforeSend: getHeaders,
        uploadProgress: function (event, position, total, percentComplete) {
            $out.html('Progress - ' + percentComplete + '%');
            $progress.val(percentComplete);
        },
        statusCode: {
            206: function (data, status, xhr) {
                console.log(status);
                console.log(data);
                console.log(xhr);
            }
        },
        success: function (data, status, xhr) {
            if (xhr.status === 206) {
//                $out.html(`<pre>${data.message}</pre>`);
//                let socket = new WebSocket(`wss://${location.host}${data.url}`);
//                socket.onmessage = function(event) {
//                  $out.append(`<pre>${event.message}</pre>`);
//                };
//                socket.onerror = function(error) {
//                  console.log(error)
//                  $out.append(`<pre>${error}</pre>`);
//                };
//                    console.log(xhr);
                OverHijack($out, data);
                return
            }
            $out.html(status);
            // TODO: добавить загрузку скрипта, если функция определена, но не подключена!
            if (successFunction !== undefined) {
                successFunction(data, thisForm);
            } else {
                afterSaveAnyForm(data, status);
            }
            $.fancybox.close();
        },
        error: function (xhr, status, error) {
            if (errorFunction !== undefined) {
                errorFunction(error, thisForm);
            } else {
                $out.html(xhr.responseText);
                switch (xhr.status) {
                    case 206: {
                        fancyOpen(xhr.responseText);
                        return
                    }
                    case 400: {
                        if (xhr.responseJSON.formErrors !== undefined) {
                            let formErrors = xhr.responseJSON.formErrors
                            for (let x in formErrors) {
                                let formsInput = $(`input[name=${x}]`, thisForm)
                                if (formsInput.length > 0) {
                                    let errorLabel = formsInput[0].nextElementSibling;
                                    errorLabel.textContent = formErrors[x];
                                    $(errorLabel).show();
                                    break;
                                }
                            }
                        }
                        return
                    }
                    case 401: {
                        urlAfterLogin = thisForm;
                        $('#bLogin').trigger("click");
                        return;
                    }

                    default:
                        alert(xhr.responseText);
                }
            }
        },
        complete: function (xhr, status, obj) {
            $progress.hide();
            $loading.hide();
            console.log(xhr);
            console.log(obj);
        }
    });

    return false;
}

function OverHijack($out, resp) {
    $out.append(`<pre>${resp.message}</pre>`);
    var method = (resp.method !== undefined ? resp.method : "GET");

    $.ajax({
        url: resp.url,
        async: true,
        cache: false,
        contentType: false,
        type: method,
        data: {
            "lang": lang,
            "html": true
        },
        beforeSend: getHeaders,
        success: function (data, status, xhr) {
            switch (xhr.status) {
                case 206:
                    if (data.url !== undefined) {
                        resp.url = data.url
                    }
                    if (data.message !== undefined) {
                        resp.message = data.message;
                    } else {
                        resp.message = data
                    }
                    console.log(data);
                    OverHijack($out, resp);
                    return;
                case 202: {
                    $out.html(data);
                    return
                }
                default:
                    $out.html(data);
            }
        },
        error: function (xhr, status, error) {
            if (xhr.status === 401) {
                urlAfterLogin = url;
                $('#bLogin').trigger("click");
                return;
            }

            fancyOpen("Code : " + xhr.status + ", " + error + ": " + xhr.responseText);
            console.log(xhr);
        }
    });
}

// ОБщие события для форм - стандарт
function formInput(thisForm) {

}

function formReset(thisForm) {
    if (confirm('Reset all data?'))
        return false;

}

function FormIsModified(event, thisForm) {
    event = event || window.event;

    $('input[type=image], input[type=submit], input[type=button]', thisForm).show();
    thisForm.State.value = '✎';
}

function formDelClick(thisButton) {
    $.post('/admin/row/del/', {
        table: $('input[name="table"]').val(),
        id: $('input[name="id"]').val()
    }, succesDelRecord);
}

function succesDelRecord(data, status) {
    if (status === "Success") {
        $('form').hide();
        alert("Success remove record !" + data)

    } else {
        alert(data);

    }
}

function showFormModal(data) {

    $.fancybox({
        content: data,
        scrolling: 'none',
        padding: 5,
        type: 'data',
        autoWidth: true,
        autoHeight: true,
        autoResize: false,
        closeBtn: false,
        modal: true,
        transitionIn: 'elastic',
        transitionOut: 'elastic',
        topRatio: 0.3, // по центру для регистрации
        leftRatio: 0.3,
        title: 'Знаком (*) помечены поля обязательные для ввода!',
        autoDimensions: true,
        overlayShow: true,
        helpers: {
            overlay: {showEarly: true},
            title: {
                type: 'float'
            }
        }
    });

    return false;
}

function alertField(thisElem) {
    let elem = $(thisElem);
    var nameField = elem.next('span').data("placeholder") || elem.next('span').text() ||
        elem.parent('label').text();
    if (nameField === "" || nameField === undefined) {
        nameField = thisElem.placeholder || elem.data("placeholder")
    }
    let errLabel = elem.parent('label').children('.errorLabel');
    let msg = 'need correct data!';
    if (thisElem.required) {
        msg = ' is required. Please, fill it';
    } else if (errLabel && errLabel.text() > "") {
        msg = errLabel.text();
    }
    alert(`Field '${nameField}' ${msg}`);
    if (elem.hasClass('suggestions-constraints')) {
        elem = elem.parents('label').children('input:first');
    }
    elem.addClass('error-field').focus();
    thisElem.scrollIntoView(100);
    errLabel.show();
}

function correctField(thisElem) {
    $(thisElem).removeClass('error-field');
}

function validatePattern(thisElem) {
    let re = thisElem.pattern,
        result = true;

    if (re === "") {
        return true;
    }

    try {

        re = new RegExp(re);
        result = re.test(thisElem.value);
        if (result) {
            thisElem.style.borderColor = 'green';
            $(thisElem).next('.errorLabel').hide();
        } else {
            thisElem.style.borderColor = 'red';
            $(thisElem).next('.errorLabel').show();
        }

    } catch (e) {
        console.log(e)
    }

    return result;
}

//   TODO: добавить попозже проверку типов полей!
function validateReguiredFields(thisForm) {

    var result = true;

    $('input[required]:visible, select[required]:visible', thisForm).each(
        function (index) {
            //TODO: тут поставить проверку чекбоксов на то, что их выставили!!! this.checked
            if (!this.value || ((this.type === "checkbox") && !(this.checked))) {
                result = false;
                alertField(this);

                return false;
            } else {
                correctField(this);
            }
        }
    );

    return result;
}

// проверка полей с выставленными патеррнами
function validatePatternsField(thisForm) {
    var result = true;

    $('input[pattern]:visible', thisForm).each(
        function (index) {
            result = result && validatePattern(this);
            if (!result) {
                alertField(this);
            }

            return result;

        });

    return result;
}

function validateFields(elem) {
    return (validateReguiredFields(elem) && validateEmailFields(elem) && validatePatternsField(elem))
}

function validateEmail(email) {
    var re = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(email);
}

function validateEmailFields(thisForm) {

    var result = true;

    $('input[type=email]:visible', thisForm).each(
        function (index) {
            result = validateEmail(this.value);
            if (result) {
                correctField(this);
            } else {
                alertField(this);
                return false;
            }
        });

    return result;
}

// handling response AnyForm & render result according to structures of data
function afterSaveAnyForm(data, status) {

    if (data.content_url !== undefined) {
        loadContent(data.content_url);
    } else if (data.formActions !== undefined) {
        loadContent(data.formActions[0].url);
    } else if (data.error !== undefined) {
        alert(data.error);
    } else {
        console.log(data);
    }

    if (data.message !== undefined) {
        alert(data.message);
    }
}

// ПОСЛЕ сохранение комнаты
function changeLoginForm() {
    $('#fLogin').attr('action', '/user/login/signup');
    $('#fLogin figcaption').toggle();

    return false;
}

// установка пола при регистрации
function signSuggestion(suggestion) {
    console.log(suggestion);
    switch (suggestion.data.gender) {
        case "MALE":
            $('#sex').show().val(0);
            break;
        case "FEMALE":
            $('#sex').show().val(1);
            break;
    }
}

// создаем новый элемент из панели набора галочек, меняем название - этот механизм нужно изменить потом
// TODO: change this code in future
function addNewItems(thisButton) {
    var data = $(thisButton).data(),
        parentDiv = $('div#' + data.parentDiv),
        newItem = $(thisButton).prev().val(),
        li = $('li:last', parentDiv).clone(),
        input = $('input', li).val(newItem);

    $("label", li).text(newItem).append(input);
    $('ul', parentDiv).append(li);
    // li.append("<span>" + newItem + "</span>");

    return false;
}

// создаем новый элемент из панели набора галочек, меняем название - этот механизм нужно изменить потом
// TODO: change this code in future
function addNewRowTableID(thisButton) {
    var data = $(thisButton).data(),
        lastTr = $('tr#' + data.lastTr),
        parentTable = lastTr.parents('table').first(),
        tr = lastTr.clone();

    // обнуляем id
    tr[0].id = '';
    // обнуляем поля ввода
    $('input, select', tr).val('');

    parentTable.append(tr);

    // переносим фокус в первый элемент ввода
    $('input, select', tr).first().focus();

    return false;
}

function inputSearchKeyUp(thisElem, event, forceEnter) {

    let x = event.which || event.keyCode;
    var thisClass = 'select.suggestions-select-show.' + thisElem.attributes.data.value
    if (x === 13) {
        $(thisClass).removeClass('suggestions-select-show').addClass('suggestions-select-hide');
        return true;
    }
    // todo handling arrows, nonchanges value ect.

    let elem = $(thisElem)
    var thisClassH = 'select.suggestions-select-hide.' + thisElem.attributes.data.value

    if (x === 40) {
        elem.unbind("blur");
        $(thisClass).focus().children('option:first').selected();

        return false;
    }

    elem.on("blur", function () {

        if (event.relatedTarget && event.relatedTarget.className === "suggestions-select-show") {
            return;
        }

        $(thisClass).removeClass('suggestions-select-show').addClass('suggestions-select-hide');
    })

    if (elem.val().length < 2) {
        return true;
    }

    $.ajax({
        url: thisElem.src,
        data: {
            "value": thisElem.value,
            "count": 10
        },
        beforeSend: getHeaders,
        success: function (data, status) {
            $(thisClassH).removeClass('suggestions-select-hide').addClass('suggestions-select-show');
            if (typeof data === 'string') {
                $(thisClass).html(data)
            } else if (Array.isArray(data)) {
                let opts = '';
                data.forEach(function (item) {
                    opts += `<option value=${item.id} title=${item.title}>${item.name}</option>`;
                })
                $(thisClass).html(opts);
            } else if (data === null) {
                $(thisClass).html(`Not fount from '${thisElem.value}'`);
            } else {
                console.log(typeof data);
                return
            }
            $(thisClass).off("keydown");
            $(thisClass).on("keydown", event => {
                let x = event.which || event.keyCode;
                if ((x === 32) || (x === 13)) {
                    thisElem.value = $(thisClass + ' option:selected').text();
                    $(thisClass).removeClass('suggestions-select-show').addClass('suggestions-select-hide');
                    if (forceEnter === true) {
                        elem.focus();
                    }
                    event.stopPropagation();
                    return forceEnter && (x === 13);
                }
            });
            $(thisClass + ' option').off("mousedown");
            $(thisClass + ' option').on("mousedown", function (e) {
                thisElem.value = $(this).text();
                $(thisClass).removeClass('suggestions-select-show').addClass('suggestions-select-hide');

                return false;
            });

        },
        error: function (xhr, status, error) {
            alert("Code : " + xhr.status + " error :" + error);
            console.log(error);
        }
    });

    return false;
}

function ShowBlocks(thisElem) {
    let d = $(thisElem).data('show-blocks');
    let f = $(thisElem).parents('form');
    f.children('figure').hide();
    d[$('option:selected', thisElem).val()].every(e => {
        $('#block' + e).show();
        return true;
    })
}

function GotoBlock(elem, blockId) {
    $(elem).parents('form').children('figure:visible').hide();
    $(`#${blockId}`).show();
    return false;
}

function Prev(elem, id) {
    $(elem).parents('figure').hide();
    $(`#block${id}`).show();
    return false;
}

function Next(elem, id) {
    let block = $(elem).parents('figure');
    let oldId = block[0].id;
    let f = block.parent('form');
    if (!validateFields(block[0]))
        return false;

    block.attr('validated', true).hide();

    let newBlock = $(`#block${id}`).show()[0];
    newBlock.scrollIntoView();
    let fields = $('input, select', newBlock)
    if (fields.length > 0) {
        fields[0].focus();
    }

    let nav = getFormNav(f);
    if (nav.children(`button#go${oldId}`).length === 0) {
        let caption = block.children('figcaption').text();
        nav.append(`<button class="button" onClick="return GotoBlock(this, '${oldId}')" data-block="${oldId}">${caption}</button>`);
    }

    return false;
}

function getFormNav(thisForm) {
    return $('header#navBlocks', thisForm);
}