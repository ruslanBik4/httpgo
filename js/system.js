/**
 * Created by rus on 12.10.16.
 */
"use strict";
var imgItem, divContent, default_page = '/main/', notSaved;
$(function() {

    // возможно, это можно сделать прямо в заголовке, а не тут
    divContent = $('#content');
    AddClickShowOkno( $("body") );

    // $.get('/user/login/', function (data) {
    //     if (data.substr(0,5) == '<form') {
    //         showFormModal(data);
    //     } else { // уже залогинен, обрабатываем данные
    //         afterLogin( JSON.parse(data) );
    //     }
    // });
});
function AddClickShowOkno( parent_this ) {
    $( 'a[href]:not( a[target="_blank"], a.fancybox-button, a:has(img.fancybox-button), a[href="#"], a[data-toggle="tab"],a[data-toggle="modal"], a[onclick], a[href*="skype:"], a.referal )', parent_this ) //referal и title*=\'(переход) означает ссылки, которые не надо менять [target!=_blank]  (откроется в новом окне)
        .each( function () {
            this.onclick = function () {
                // document.URL = this.href;
                history.pushState(this.pathname, '', this.pathname);
                divContent.load(this.href,
                    function () {
                        document.title = this.title;
                        AddClickShowOkno(divContent);
                        var requestWrap = $('#requestWrap');
                        getData(requestWrap);
                    });
                return false;

            }
        });
}
// после загрузки новой страницы
// тут можно отрабатывать события, например, на расстановку евентов для элементов и так далее
function SitePostShow() {

    //красивый скроллинг, потом проработать что бы он применялся ко всем элементам со скроллом
    // $(".main-form-wrap").niceScroll();
    var dateInputs = $("input[type=date],input[type=datetime]");
    

    //стилизация селектов и инпутов, смена бг
    $(".business-form-select").styler();
    changeBg();
    moveLabel();

    // устанавливаем необходимые датапикеры
    if (dateInputs.length > 0) {
    //TODO: сделать проверку или установить флаг на то, что модуль уже загружен и не загружать если так
        $("<head>").append('<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery-datetimepicker/2.5.4/build/jquery.datetimepicker.full.min.js"></script>')
        dateInputs.each(function() {
            var maxDate = $(this).attr('maxdate');
            var minDate = $(this).attr('mindate');
            if(maxDate){
                $(this).datetimepicker({
                    format: this.type == "datetime" ? 'Y-m-d H:i' : 'Y-m-d',
                    maxDate: maxDate
                });
            } else if(minDate){
                $(this).datetimepicker({
                    format: this.type == "datetime" ? 'Y-m-d H:i' : 'Y-m-d',
                    minDate: minDate
                });
            } else {
                $(this).datetimepicker({format:'Y-m-d'});
            }

        });
    }

$('.get-json[data-href]').each( function() {this.data('href')})

    //TODO: сделать подключение остальных полей (без maxDate) и, других своств - minDate, например

    //используем как событие загрузки формы
    $("form[oninvalid]").trigger('invalid');

    // if($('#fapproximation').length > 0){
    //     //TODO: научить проверять этот метод,  на загруженнсть джс
    //     enableApproximationHandler();
    // }
}
function moveLabel(){
    $(".custom-input-label").click(function(){
        $(this).addClass('small-label');
    });
    $(".input-label").click(function(){
        $(this).next('.form-items-wrap').find(".custom-input-label").addClass('small-label');
    });
    $('.business-form-input').keyup(function(){
        if($(this).val().length != 0){
            $(this).next(".custom-input-label").addClass('small-label');
        } else {
            $(this).next(".custom-input-label").removeClass('small-label');
        }
    });
}
function changeBg(){
    var checkFrom = $('#fbusiness');
    if(checkFrom.length > 0){
        $('.content-wrap').css('background-image','url(/images/bg2.png)')
    } else {
        $('.content-wrap').css('background-image','url(/images/bg1.png)')
    }
}
// перед загрузкой новой страницы, чистим хвосты
function beforeLoadContent() {
    $('canvas, .rmpPlayer, img.draggable').detach();
    $('#dRoomtools').html('');
    divContent.css({background: ''});
}
// собственно, нужен для того, чтобы после регистрации отобразтит в заголовке нечто
function afterSignup(data) {
    if (!data)
        return false;

    divContent.load('/show/forms/?name=signin&email=' + data.email);
}
// собственно, нужен для того, чтобы после авторизации отобразтит в заголовке нечто
function afterLogin(data)
{
    if (!data)
     return false;

    var greetings = ( (data === null) || ( data.login === undefined) ? '' : 'Добро пожаловать, '
                    + (data.sex === undefined ? "" : data.sex + " ") + data.login + '!');

    $('#sTitle').html( greetings );
    // $('#fTools > output').text( 'Можете добавить устройство из меню и перенести его в нужное Вам место.');
    // конфликтует с webcomponent.js
    // loginToggle();

    // $('#dMyTools').load('/user/usertools/menu');
    // $('#dMyRooms').load('/user/rooms/menu');

    $.get('/user/profile/', function (data) {
        data = GetPageParts(data);
        divContent.html(data);
        AddClickShowOkno( divContent );
    } );
}
// события после кнопки Выйти
function logOut(thisElem) {
    $('canvas').detach();
    loginToggle();
    $('#sTitle').html( 'Для начала работы Вам необходимо ' );
    $.get( thisElem.href, function (data) {
            showFormModal(data);
    });

    return false;
}
function loginToggle() {
    $('.btn-login').toggle();
}
function getOauth(thisElem)
{
    var dataElem = $(thisElem).data(),
        props    = dataElem.props,
        width = ( dataElem.width ? dataElem.width : "98%"),
        height = ( dataElem.height ? dataElem.height : "98%"),
        text = '';

    $(thisElem).next().load(dataElem.href);
    return false;
}
// добавление фото для новой комнаты
function handleFileSelect(evt) {
    var files = evt.files || evt.target.files, // FileList object
        $progress = $('#progress').show(),
        reader = new FileReader(),
        f = files[0];

    if (files.length < 1)
        return false;

    reader.onload = (function(theFile) {
        return function(e) {
            // Render thumbnail.
            divContent.css( { backgroundImage: 'url(' + e.target.result + ')' } ).html('');
            // $('svg', divContent).append("<img class='new_photo' src='" + e.target.result +
            //     "' title='" + escape(theFile.name) + "' alt='" + escape(theFile.name) + '/>');
        };
    })(f);

    // Read in the image file as a data URL.
    reader.readAsDataURL(f);
}

function showFlat(thisElem) {
    var images = '',
        bkgPos = '0% 0%, 50% 0%, 0% 50%, 50% 50%',
        bkgSize= '25%',
        comma  = '';

    beforeLoadContent();

    $('ul > li > a', thisElem.parentElement).each( function() {
        var dataElem = $(this).data();

        if (dataElem.image === undefined)
            images += comma + "url('img/room.svg')";
        else
            images += comma + "url('" + dataElem.image.substr(3) + "')";

        for (var i in dataElem.props ) {
            AddItem( dataElem.props[i] );
        }
        comma   = ',';
    });

    divContent.css( { backgroundImage: images, backgroundPosition: bkgPos, backgroundSize: bkgSize  }).html('');

    return false;
}
// на слуяай отвала зароса по AJAx
function failAjax(data, status) {
    console.Log(data);
    alert(status);
}
function getData(element) {
    var url = element.attr('data-href');
    if(url){
        console.log(url);
        $.get(url).done(function (data) {
            var fields = data.fields;
            var form = data.form;
            console.log(data);
            $.each(fields, function(key,object){
                renderFormElements(key,object,element,form);
            });
        });
        
    }

}

function renderFormElements(name, object, parent,form) {
    parent.find('form').attr({
        action   : form.action,
        id       : form.id,
        name     : form.name,
        onload   : form.onload,
        onsubmit : form.onsubmit,
        oninput  : form.oninput,
        onreset  : form.onreset
    });
    var thisElem = $('input[name=' + name + ']');
    parent.find(thisElem).addClass(object.CSSClass).attr({
        type       : object.type,
        placeholder: object.title,
        title      : object.title,
        maxlenght  : object.maxLenght,
    });
    if(object.required ){
        thisElem.attr('required', true);
    }

}
