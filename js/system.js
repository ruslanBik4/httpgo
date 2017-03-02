/**
 * Created by rus on 12.10.16.
 */
"use strict";
var imgItem, divContent, default_page = '/main/', notSaved;
$(function() {

    // возможно, это можно сделать прямо в заголовке, а не тут
    divContent = $('#content');

    // $.get('/user/login/', function (data) {
    //     if (data.substr(0,5) == '<form') {
    //         showFormModal(data);
    //     } else { // уже залогинен, обрабатываем данные
    //         afterLogin( JSON.parse(data) );
    //     }
    // });

});
// после загрузки новой страницы
// тут можно отрабатывать события, например, на расстановку евентов для элементов и так далее
function SitePostShow() {

    $(".business-form-select").styler();
    changeBg();
    moveLabel();
    if ($("input[type=date]").length > 0) {
    //TODO: сделать проверку или установить флаг на то, что модуль уже загружен и не загружать если так
        $("<head>").append('<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery-datetimepicker/2.5.4/build/jquery.datetimepicker.full.min.js"></script>')
    }

    $("input[type=date][maxDate]").each(function() {
        $(this).datetimepicker({
        format:'Y-m-d',
        onShow:function(){
            this.setOptions({
                maxDate: $(this).attr('maxDate')
            })
        }
        });
    });
    $("input[type=datetime][maxDate]").each(function() {
        $(this).datetimepicker({
        format:'Y-m-d',
        onShow:function(){
            this.setOptions({
                maxDate: $(this).attr('maxDate')
            })
        }
        });
    });

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
    loginToggle();

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
        if (data.substr(0,5) == '<form') {
            showFormModal(data);
        } else {
            alert(data);
        }
    } );

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

