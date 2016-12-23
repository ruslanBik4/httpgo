"use strict";
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
    $(thisElem).css( { border: '1px red solid' } ).focus();
}
function correctField(thisElem) {
    $(thisElem).css( { border: '' } );
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
            if ( !this.value ) {
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

function saveForm(thisForm, successFunction)
{
    var $out = $('output', thisForm),
        $loading = $('.loading', thisForm),
        $progress = $('progress', thisForm);

    if (!(validateReguiredFields(thisForm) && validateEmailFields(thisForm)))
        return false;

    $(thisForm).ajaxSubmit({
        beforeSubmit: function(a,f,o) {
            o.dataType = "json";
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
            successFunction( data );
            $.fancybox.close();
        },
        error: function(error, status) {
            $out.html( error.responseText );
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
function afterSaveAnyForm(data) {

    if (data.contentURL !== undefined) {
        ShowOkno(data.contentURL);
        // TODO  catalog after
    } else {
        alert(data.error)
    }

}
function showAddToolsForm(thisElem, notSaved) {
    var dataElem = $(thisElem).data(),
        props = dataElem.props,

        inputName = '<label class="required">Название</label><input type="text" name="name" required placeholder="Например, место его расположения"/>',
        propsValue = '', comma = '',
        endform = '<input id="isAdding" type="checkbox" checked/><label>Добавить в текущую комнату</label><button>Добавить в реестр</button>',
        text = '<form method="post" class="login" action="/user/usertools/add" onsubmit="return saveForm(this, afterSaveTools);">'
                + '<input hidden name="id_tools" value="' + dataElem.id + '"/>'
                + (notSaved ? inputName : '<figcaption>' + thisElem.title + '</figcaption>'),
        prefixText = notSaved ? 'Введите ' : '',
        rflog = '';

    for (var i in props ) {

        if ( props[i]== "?" || !notSaved) {
            text +=  '<div><label ' + (notSaved ? 'class="required"' : '') + '>' + prefixText + getNameFromID(i)
                    + '</label><input type="text" required name="props:' + i
                    + (notSaved ? '' : '" value="' + props[i] ) + '"/></div>';
        } else {
            propsValue += comma + '\"' + i + '\":\"' + props[i] +'\"';
            comma = ',';
        }
    }
    // если есть свойства, которые не редактируют, то их сохраняем в отдельном параметре
    if (propsValue) {
        propsValue = "<input hidden name='props' value='" + propsValue + "'/>";
    }

    $.fancybox( text + propsValue + endform + rflog + '</form>',
        {
            title: 'Данные для прибора ' + thisElem.title,
            transitionIn	 : 'elastic',
            transitionOut	 : 'elastic',
            helpers		: {
                overlay : { showEarly  : true },
                title	: { type : 'float'
                }
            },
            'onClosed': function (currentArray, currentIndex, currentOpts) {
                // Use closedParam here
                if (notSaved && confirm( 'Вы внесли изменения и не сохранили результат. Закрыть форму?' ))
                    return false;
            }
        });

    $(thisElem).parent('div').slideUp();

    return false;

}
function saveRoom(thisForm)
{
    var props = '',
        comma = '';

    $('.draggable').each( function () {
        var coords    = getCoords(this),
            propsData = $(this).data();
        props += comma + '{' + '"left":' + coords.left + ',"top":' + coords.top ;
        for (var i in propsData ) {
            props += ',' + '"' + i + '":"' + propsData[i] + '"';
        }
        props += '}';
        comma = ',';
    });
    $('input[name=props]', thisForm).val(props);

    saveForm(thisForm, afterSaveRoom);

    return false;
}
// ПОСЛЕ сохранения инструмента
function afterSaveTools(data) {
    $('#dMyTools').load('/user/usertools/menu');

    if ($('#isAdding:checked').length > 0)
        AddItem( data );

}
// ПОСЛЕ сохранение комнаты
function afterSaveRoom(data) {
    if (data.result !== undefined) {
        divContent.load(data.result);
    } else {
        alert(data.error)
    }
}

function changeLoginForm() {
    $('#fLogin').attr('action', '/user/login/signup');
    $('#fLogin figcaption').toggle();

    return false;
}
function busnessSuggestion(suggestion) {
    var data = suggestion.data;

    console.log(suggestion);
    if (!data)
        return;

    $('#inn').val(data.inn);
    $('#kpp').val(data.kpp);
    $('#ogrn').val(data.ogrn);
    $('#okpo').val(data.okpo);
    $('#type').val(data.type);
    $('#address').val(data.address.value);
//            $('#inn').val(data.inn);
//            $('#inn').val(data.inn);

}

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



