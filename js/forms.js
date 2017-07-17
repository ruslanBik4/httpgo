"use strict";
// ОБщие события для форм - стандарт
function formInput(thisForm) {

}
function formReset(thisForm) {
    if ( confirm('Очистить все введенные данные?') )
        return false;

}
function FormIsModified( event, this_form ) {
    event = event || window.event;

    $( 'input[type=image], input[type=submit], input[type=button]', this_form ).show();
    this_form.State.value = '✎';
}
function formDelClick(thisButton) {
    $.post('/admin/row/del/', {table: $('input[name="table"]').val(), id: $('input[name="id"]').val() }, succesDelRecord);
}
function succesDelRecord(data, status) {
    if (status == "Success") {
        $('form').hide();
        alert("Успешно удалили запись!" + data )

    } else {
        alert(data);

    }
}
function showFormModal(data) {

    $.fancybox( {
        content	: data,
        scrolling : 'none',
        padding: 5,
        type : 'data',
        autoWidth: true,
        autoHeight: true,
        autoResize: false,
        closeBtn	: false,
        modal		: true,
        transitionIn	 : 'elastic',
        transitionOut	 : 'elastic',
        topRatio	: 0.3, // по центру для регистрации
        leftRatio	: 0.3,
        title		: 'Знаком (*) помечены поля обязательные для ввода!',
        autoDimensions: true,
        overlayShow: true,
        helpers		: {
            overlay : { showEarly  : true },
            title	: { type : 'float'
            }
        }
    });

    return false;
}
function alertField(thisElem) {
    var nameField = $('label[for="' + thisElem.id + '"]').text();
    if (nameField == "") {
        nameField = thisElem.placeholder
    }
    alert( 'Заполните поле - ' + nameField );
    $(thisElem).css( { 'border-bottom': '1px red solid' } ).focus();
}
function correctField(thisElem) {
    $(thisElem).css( { border: '' } );
}
function validatePattern(thisElem) {
    var re = thisElem.pattern,
        result = true;

    if (re == "") {
        return true;
    }
    // TODO: добавить проверку на валидность - иногда вылетает тут
    try {

        re = new RegExp(re);
        result = re.test(thisElem.value);
        if(result){
            $(thisElem).css('border-color','green');
        } else {
            $(thisElem).css('border-color','red');
        }

    } catch (e) {
        console.log(e)
    }
    return result;
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
// проверка запоолнения обязательных полей
//   TODO: добавить попозже проверку типов полей!
function validateReguiredFields(thisForm) {

    var result = true;

    $('input[required]:visible, select[required]:visible', thisForm).each(
        function (index) {
            //TODO: тут поставить проверку чекбоксов на то, что их выставили!!! this.checked
            if ( !this.value || ( (this.type == "checkbox") && !(this.checked) ) ) {
                result = false;
                alertField(this);

                return false;
            }
            else {
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
function validateFields(thisForm) {
    //TODO : что бы подсвечивало все невалидные поля

    return (validateReguiredFields(thisForm) && validateEmailFields(thisForm) && validatePatternsField(thisForm))
}
// стандарт созранения форм, требуется расширить количество обработчиков (callback) типа ошибок записи и т.П.
function saveForm(thisForm, successFunction, errorFunction)
{
    // TODO: create element form for output form result
    var $out = $('output', thisForm),
        $loading = $('.loading', thisForm),
        $progress = $('progress', thisForm);

    if (!validateFields(thisForm))
        return false;

    $(thisForm).ajaxSubmit({
        beforeSubmit: function(a,f,o) {
            o.dataType = "json";

            //удаляем пустые поля
            var isNewRecord = $('input[name=id]').length == 0;

            for( var i = a.length -1; i >= 0; --i){
                if ( (a[i].value === '') && (isNewRecord || a[i].type === 'select-one') ) {
                    a.splice(i,1);
                }
            }

            // добавляем чекбокс-поля, которые были отменены в форме

            if (f.attr('data-form-id')) {
              $("input[form='"+ f.attr('id') + "'][type=checkbox][checked]:not(:checked)").each(function() {
                a.push({ name: this.name, value: 0, type : this.type, required: this.required })
              });
            } else {
              $("input[type=checkbox][checked]:not(:checked)", f).each(function() {
                  a.push({ name: this.name, value: 0, type : this.type, required: this.required })
              });
            }


            $out.html('Начинаю отправку...');
            $progress.show();
            $loading.show();
        },
        uploadProgress: function(event, position, total, percentComplete) {
            $out.html( 'Progress - ' + percentComplete + '%' );
            $progress.val( percentComplete );
        },
        success: function(data, status) {
            $out.html('Успешно изменили запись.');
            // TODO: добавить загрузку скрипта, если функция определена, но не подключена!
            if (successFunction !== undefined) {
                successFunction(data, thisForm);
            } else {
                afterSaveAnyForm(data);
            }
            // $.fancybox.close();
        },
        error: function(error, status) {
            $out.html( error.responseText );
            if (errorFunction !== undefined) {
                errorFunction(error, thisForm);
            } else {
                alert(error);
            }
        },
        complete: function(data, status) {
            $progress.hide();
            $loading.hide();
            console.log(status);
            console.log(data);
        }
    });

    return false;
}
// стандартная обработка формы типа AnyForm после успшного сохранения результата
function afterSaveAnyForm(data) {

    if (data.contentURL !== undefined) {
        ShowOkno(data.contentURL);
        // TODO  catalog after
    } else if (data.error !== undefined) {
        alert(data.message);
    } else {
        console.log(data);
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
        newItem   = $(thisButton).prev().val(),
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
