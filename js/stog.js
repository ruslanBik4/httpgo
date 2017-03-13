// Эта функция jQuery выполняется после полной загрузки страницы - запускаем дополнительные скрипты и заполняем содержимое страницы
var default_page = default_edit_page = 'text_main.htm', new_editor = multilang = correct_pass = isNewSitecraft = RunDialog = adminMode = name_page = type_page = scrollcolor = body_scrollcolor = '', title_doc ='Главная', tools_is_load = isNoSitecraft = false, editor, DivContent, input_focused, typeAnimate, password, topEditorOffset, root_page = 'index.htm', NotZagl, min_rez;

var multilang_ext = {  'Рус' : '',  'Укр' : '_ua', 'Eng' : '_eng' };
var anchor = '', str_location = document.domain + document.location.pathname;

$(function()   {
    var password = '';

    if( (DivContent = $("#pane")).length == 0) {
        DivContent = $('#content');
    }

    if ( !default_page )
        default_page = 'text_main';

// инициализация доп. скрипов при запуске сайта	
    try
    {
        if (!$('body').hasClass( 'complete' ) ) //первая загрузка, догружаем скрипты и содержимое
        {
            if (body_scrollcolor || scrollcolor)
            // подключаю продвинутые полосы прокрутки к окну или Содержимому
                LoadJScript( "http://solution.allservice.in.ua/js/jquery.nicescroll.js", true, true,
                    function () {
                        if (scrollcolor)
                            DivContent.niceScroll( {cursorcolor: scrollcolor } );
                        if (body_scrollcolor)
                            $('body').niceScroll( { cursorcolor: body_scrollcolor } );
                    }
                );
            LoadJScript( "/tools.js", true, true, FirstLoads );
            LoadJScript( "http://solution.allservice.in.ua/js/scrollto/jquery.scrollTo-min.js", true, true );
            $('body').addClass( 'complete' );
        }

        if ( NotZagl=($("#zag").length == 0) )
            DivContent.before( '<div id="zag" style="display:none" > </div>');
    }
    catch(e) {
        alert(e.message);

    }

    isNewSitecraft = ($('link[href*="sc.css"]').length > 0) // признак СК со стилями
        || ($("meta[artel='noSC']").length > 0);  // либо вообще без СК
    isNoSitecraft = ($("meta[artel='noSC']").length > 0);
}) // $(document).ready

// инициализация редактора
function InitEditor(id) {
    var str_hash = document.location.hash.substring(1),
        btnEditorUp = '<a href="" onclick="return EditorTop();" > <img src="http://solution.allservice.in.ua/images/uup.png" /> </a>',
        btnEditorDown = '<a href="" onclick="return EditorDown();" > <img src="http://solution.allservice.in.ua/images/ddn.png" /> </a>',
        i = str_hash.search('.php');

    if ( !name_page && (i > 0)  )
    {
        ShowOknoFromHash();
        return;
    }

    id = name_page || ( (str_hash && (str_hash != 'admin') ) ? str_hash : default_edit_page || default_page || 'text_main' );

    $.post( 'loadpage.php', {name_page : id },
        function(data, status) {
            // подключаем редактор ТОЛЬКО  в случае подключенного редактирования страниц
            DivContent.html('<form id="formCKeditor" action="savetext.php" target="_blank" enctype="multipart/form-data" method="post"> <input type="hidden" name="name_page" value"' + id + '"/> <textarea id="edit_text" name="edit_text" class="ckeditor"></textarea></form>').show();

            editor = CKEDITOR.replace( 'edit_text' );// настройки в конфиге
            editor.setData( data, function () {
                var topEditor = cookieGet('cke_top') || topEditorOffset || DivContent.offset().top - $('.cke_top').height();


                // если есть навигатор, то под него кладем панельку
                if ($('nav').length > 0) {

                    topEditor = $('nav').offset().top + $('nav').height();
                    console.log($('nav'), topEditor, $('nav').offset().top, $('nav').height() );
                }
                // делаю изменеия внешнего вида редактора - обязательно после загрузки текста !
                $('.cke_top').css( { top : topEditor } ).addClass('notHidden').append( btnEditorUp ).append( btnEditorDown );
                $('.cke_top .cke_toolbar_break').detach();
                $('.cke_contents').css( { height : DivContent.height() } );
                $('.cke_bottom').css( { top : DivContent.offset().top + DivContent.height() + 5 } );

                name_page = id;
                type_page = '.htm';
                SetDocumentHash( id );

            } );
        }).fail( function(success) { console.log('На этом сайте не предусмотрено редактирование страниц! Либо это ошибка загрузки loadpage.php ' + success.statusText ); });


}
// приподнять редактор выше
function EditorTop() {
    var cke_top = $('.cke_top'), new_top = parseInt( cke_top.css( 'top' ), 10 ) - 10;

    new_top = (new_top > 0 ? new_top : 0);
    cke_top.css( { top : new_top } );
    cookieSet('cke_top', new_top);

    return false;
}
// опустить редактор ниже
function EditorDown() {
    var cke_top = $('.cke_top'), new_top = parseInt( cke_top.css( 'top' ), 10 ) + 10;

    new_top = (new_top > 0 ? new_top : 0);
    cke_top.css( { top : new_top } );
    cookieSet('cke_top', new_top);

    return false;
}
// смена языка текущей страницы
function ChangeLang( new_lang ) {
    var id = DivContent.attr( 'rel').replace(/&multilang=(\S)+/i, ''),
        arr_id = id.split('.')/*
         ,
         input_focused = $('input:focus').attr('id')
         */;

    if ( $(this).hasClass('red_border') )
        return false;
    // получаем имя страницы, для русских обрезаем только расширение (по точке),
    // для остальных - по суффиксу для языка
    name_page = arr_id[0];
    type_page = '.' + (arr_id[1] || 'htm');
    multilang = multilang_ext[ new_lang ];

    HidePane( AfterHide );
    $('#multilang > a.selected')/* .removeAttr('disabled') */.removeClass('selected');
    $('#multilang > a[href="#' + new_lang + '"]').addClass( 'selected' )/* .attr('disabled', 'disabled') */;

    // меняю тексты на элементах страницы
    $('input[' + (multilang || '_rus') + ']').each( SwapLanguage );
    cookieSet('language', multilang);
    if (input_focused !== undefined)
    {
        $('#'+input_focused).focus();
        $('body').scrollTo( $('#'+input_focused), 800, {margin:true, axis: 'y'} );
    }
    return false;

}

function SwapLanguage() {
    var ua = $(this).attr( multilang || '_rus' );

    if ( $(this).attr('_rus') === undefined )
        $(this).attr('_rus', this.placeholder);

    this.placeholder = ua;

}
function FirstLoads(data, textStatus) {
    var str_hash = document.location.hash.substring(1), isPHP = ( ( str_hash ? str_hash.search('.php') : (default_page ? default_page.search('.php')	 :  -1 )) > 0);

    tools_is_load = true;

    if ( multilang ) // указано несколько языков для сайта, создаем кнопки переключения
    {
        arr_lang = multilang.split( ' ' );
        multilang= cookieGet('language') || ''; // пока по умолчанию русский язык, позже решим с куками
        if ( $('#multilang').length == 0 )
            $('body').append('<div id="multilang" style="position:absolute;right:100px;top:5px;" > </div>');

        for( var key in arr_lang  )
            $('#multilang').append( '<a href="#' + (new_lang = arr_lang[key]) + '" class="lang_ref' + ( multilang== multilang_ext[ new_lang ] ? ' selected' : '') + '" onclick="return ChangeLang(\'' + new_lang + '\');" >' + new_lang + ' </a>' );
    }

    /*
     if ( ( $( 'img[alt*=fullW]' ).length > 0 ) || ( $( '.fullW' ).length > 0 ) )
     {
     divFirst = $('div:first');
     RezinaWidth( divFirst, (wDiv = parseInt( divFirst.width() ) )  );
     //divFirst.css( { "min-width": 998 } );
     }
     */

    if ( DivContent.hasClass('fullW') )
    {


        divFirst  =  $('div:first' );
        divFirstw = parseInt( divFirst.width() );
        str_style = DivContent.attr('style');

        if ( str_style && (fMargin = str_style.match( /margin-left:(\s)*(\d+)px/ )) && ( (margin_left = parseInt( fMargin[2], 10 )) > 0) ) // с min_width  разобраться позже как будет время
        {
            margin_left = new Number( (margin_left * 100) / divFirstw).toFixed( 2 ); // пока держимся одной ширины
            DivContent.css( { 'min-width': DivContent.width(), 'margin-left' : margin_left + '%', width : 100 - margin_left*2 + '%' } );
        }

        ( $('#okno').length > 0 ? $('#okno') : DivContent ).parent('div').each( function() {

            RezinaWidth( $(this), divFirstw );
            $(this).height('auto');
        });
        //$('body').css( {  'width': '100%' } );
    }


    AddClickShowOkno( $('body') );
    $('div[rez]').each( function () {

        var rez = parseInt(  $(this).attr('rez') ),
            window_h = parseInt( $(window).height() );
        if (rez > 0)
            $(this).css( { "overflow-y" : "auto", "height" : window_h - rez } );

    });

    if ( DivContent.hasClass('scroll-pane') && min_rez ) //( rez = DivContent.attr('style').match( /rez:(\d+)px/ ) ) )
    {
        DivContent.css( { "overflow-y" : "auto", "height" : parseInt( $(window).height() ) - min_rez } );
        console.log($(window).height(), min_rez);
    }

    // режим редактирования - запрашивает пароль, если верный, инициируем редактор и загружаем в него главную страницу		
    if ( adminMode = ( ( ($('#admin').length > 0) || (str_hash == 'admin') ) && (!correct_pass)
        && (password = prompt('Введите пароль:') ) ) ) //режим редактирования
    {
        $.post('login.php', { pass: password }, function(data) {
            correct_pass = data;
            if( !correct_pass)
                DivContent.html('Вы ввели неверный пароль. Перегрузите страницу, чтобы попробовать еще раз, если Вы - наш друг. В противном случае советуем попытать счастья на менее беопасных сайтах. :-)');
            else if ( default_edit_page && (default_edit_page.search('.php') > 0) )
                ShowOkno( default_edit_page, '');
            else if ( window.CKEDITOR  === undefined )
                LoadJScript( "http://solution.allservice.in.ua/js/ckeditor-full/ckeditor.js", true, true, InitEditor );
            else if ( !isPHP )
                InitEditor();
        }  ).fail( function (success) {
            alert('Ошибка при проверки пароля! ' + success.statusText);
            input_focused = success;
        });
    }
    else // если не редактор - подгружаем страницу
        ShowOknoFromHash();

    if (!window.onpopstate)   // подключаем запись посещенных вкладышей в истории посещений
        window.onpopstate = MyPopState;

    if ( RunDialog )
        CreateDialog();
}

function LoadJScript(url, asyncS, cacheS, successFunc) {

    $.ajax({
        type: "GET",
        async: asyncS,
        cache: cacheS,
        url: url,
        global: false,
        dataType: "script",
        success: successFunc,
        error: function (XMLHttpRequest, textStatus, errorThrown) {
            if (textStatus == '')
                alert('Неудачная загрузка скрипта "' + url + '"! (' + textStatus + '). Перегрузите страницу!');
        }
    });

}

/* var CKEDITOR_BASEPATH = "http://solution.allservice.in.ua/js/ckeditor-full/"; */

function ShowOkno( id, zag, this_ref, file_ext ) {
    var Button, Button_name, DivRel = DivContent.attr( 'rel' ) || '',
        isCKEDITOR = ( window.CKEDITOR  !== undefined ), // подключен CKEDITOR
        editForCKEDITOR = ( window.pageCKEDITOR  !== undefined ); // CKEDITOR редактор вкладышей 

    /* всякие проверки корректности работы */
    if (!tools_is_load)
    {
        $("#zag").html('Ждем загрузки программных модулей!');
        return false;
    }

    /*
     if ( isCKEDITOR && editor )
     editor.destroy();
     */
    if ( !isNoSitecraft ) {

        id = ( id ?  GetShortURL(id) : default_page ); // если передан URL - ОБРЕЗАЕМ ЕГО ДО КОРОТКОГО И ПРОВЕРЯЕМ НА ПРЕДМЕТ ССЫЛОК ОНЛАЙН-МАГАЗИНА
    }
    // для загрузки ссылок на главную страницу
    if ( ( id.search( /index\.htm/i ) > -1 ) || (id == "#") )
        id = default_page;

    if ( (!file_ext) && ( id.substring(0,1) == '!' )) {
        id = id.substring(1);
        file_ext = '.php';
    }
    else if ( (!file_ext) && ( (i=id.search(/\.[htm|php|svg|psd|pdf]/)) > -1 )) {
        file_ext = id.substring( i );
        id = id.substring( 0, i );
    }

//   id = ( id.search(str_location) > -1 ? id.substring( id.search(str_location) + str_location.length ): id);
    if ( (anchor_split = id.split( '#' )).length > 1) // якоря
    {
        anchor = anchor_split[0];
        id     = anchor_split[1];
    }
    else if ( file_ext && ((anchor_split = file_ext.split( '#' )).length > 1) ) // якоря
    {
        file_ext = anchor_split[0];
        anchor   = anchor_split[1];
    }

    file_ext = ( !file_ext ? ( (id.match(/\/.*\?.*/) || id.substr(-1) == '/') ? '' : ".htm") : file_ext);

    if ( (DivRel == id + ( !file_ext ? "" : file_ext) )
        || ( (document.location.pathname == '/') && (DivRel == 'text-main') )
        || (id == '') )
    {
        if (anchor > '')
            GotoAncor( anchor );
        return false;
    }

    // Костыли для ЭРстроя
    /*
     posSlesh = DivRel.search('/') + 1; 
     if ( (anchor) && (posSlesh > 0) && (DivRel.substring(posSlesh) == id) )
     {
     GotoAncor( anchor );
     return false;   
     }
     */

    // сохраняем в глобальные переменные
    name_page = id;
    type_page = file_ext;
    title_doc = zag;

    /*   здесь отрабатываем замораживание картинки у нажатой кнопки, отмораживаем замороженные до этого */
    if (this_ref) // при условии, что нам передана ссылка на кнопку ("<a href>")
        ChangeButton(this_ref);
    else if ( $('a[href*="' + name_page + type_page + '"]').length > 0 ) // если нет  пробуем найти по адресу перехода
        ChangeButton( $('a[href*="' + name_page + type_page + '"]')[0] );

    if ( (file_ext == ".htm") && correct_pass) // режим редактирования
    {
        if ($('.tinyedit#admin').length > 0 )  // старое редактирование от Цитовича
            ReadFileToEditor( id + file_ext );
        else
            try
            {
                if ( $('input[name=name_page]').length == 0 )
                    InitEditor( id );
                else
                // пока будет всегда брать исходный
                    $.post( document.location.protocol + '//' + document.location.host + '/loadpage.php', {name_page : id },
                        function(data, status) {
                            editor.setData( data , function() {
                                this.checkDirty(); // установить признак редактирования
                                SetDocumentHash( id );
                            });
                        }).fail( function(data, status) { alert( status ) } );
                $('input[name=name_page]').val( id );
            }
            catch(e) {
                alert(e);
            }
    }
    else
    {
        // индикатор загрузки						
        $("#zag").css('display', 'block').html( "<img src='http://solution.allservice.in.ua/images/load.gif' />" ).show();
        // поддержка мультиязычности только при загрузке, больше ничего не меняем
        HidePane( AfterHide );
    }

    return false;
};
// скрытие содержимого
function HidePane( procAfterHide ) {
    if (typeAnimate  === undefined )
        DivContent.hide( "slow", procAfterHide );
    else if (typeAnimate == 'slide')
        DivContent.slideUp( "slow", procAfterHide );
    else if (typeAnimate == 'fade')
        DivContent.fadeOut( "slow", 'swing', procAfterHide );
    else
        DivContent.animate( typeAnimate, "slow", procAfterHide );
}
// операции после скрытия содержимого
function AfterHide() {

    // если определена отдельная фукция ПЕРЕД сменой контента для сайта либо странички
    if ( window.beforeLoadContent !== undefined )
        beforeLoadContent();

    if (type_page == '.svg') {
        $.get( name_page + type_page, PutContentIntoPane, 'html' ).fail(failLoadPage);
    } else if (new_editor || isNewSitecraft) { // загрузка для свежих версий СК 
        if (type_page == ".htm")
        {
            if (new_editor) // при использование редактора страниц
                $.post( document.location.protocol + '//' + document.location.host + '/loadpage.php', {name_page : name_page + multilang}, // довести до ума попозже
                    PutContentIntoPane ).fail( LoadPageToPane );
            else
                LoadPageToPane();
        } else {
            $.post( name_page + type_page, $.extend(( multilang ? { 'multilang': multilang } : '' ), (correct_pass ? {admin : ''} : '' ) ), PutContentIntoPane ).fail(failLoadPage);
        }
    } else {// загрузка старым способом
        DivContent.html( GetAjax( name_page, type_page, multilang ) );
        PostLoadOkno();
    }
}
// сообщение на случай ошибки агрузки страницы
function failLoadPage(status) {
    alert('Ошибка при запуске скрипта  ' + name_page + type_page + ' : ' + status.statusText);
}
// считывание фона - TODO добавить обрезание начала и конца html
function GetBodyBackground(data) {
    var regStr = new RegExp("<script([^>]*)>([\\s\\S]*?)<\/script>", "igm");
    /*
     if ( arr_style = data.match(/<style([\s|\S]*)body([\s|\S]+)(?=<\/style>)/i) )
     style = '<style type="text/css"> \n  .body-pane ' + arr_style[2] + '</style>';
     else
     style = '';
     */
//     Именяем селектор для фона страницы на селектор для Содержимого

    data = data.replace(/(<style[\s\S]*>\s*)(body)(?=\s+\{[^<]+<\/style>)/mg, '$1.body-pane');

    data = GetPageParts(data);

    if (isNoSitecraft || isNewSitecraft)
        return data;

    script = ( scripts = data.match( regStr) ? GetUserCode(data) : '' );

    return data.replace(regStr, script );
}
// вычленение частей кода страницы и распихивание по элементам на странице дизайна - весьма полезна!
function FindAndCutPart( data, element, re, default_html ) {
    var match; // нам нужно схватить один див

    if (element.length == 0)
        return data;

    if ( match = data.match( re ) )
    {
        element.html( match[1] );
        AddClickShowOkno( element );
        data = data.replace( match[0], '' );
    }
    else if ( typeof default_html !== "undefined")
        element.html( default_html );

    return data;
}
function GetPageParts(data) {

    data = FindAndCutPart( data, $('#theBreadCrumbs'), /<div[^>]*theBreadCrumbs.*?>\s*(.*(?=<\/div>))*<\/div>/i, (name_page + type_page != default_page ?"<a id='bBreadCrumbs' href='#' onclick='return ShowOkno( default_page );' > Главная </a> " + title_doc : '' ) );
    data = FindAndCutPart( data, $('#divpager'), /<div[^>]*divpager[^>]*>\s*(<([a-z]*)[\s\S]*<\/\2>)*\s*<\/div>/im );
    data = FindAndCutPart( data, $('#sort_pane'), /<div[^>]*sort_pane[^>]*>\s*(<([a-z]*)[\s\S]*<\/\2>)*\s*<\/div>/im );
    data = FindAndCutPart( data, $('#catalog_pane'), /<div[^>]*catalog_pane[^>]*>\s*(<([a-z]*)[\s\S]*<\/\2>)*\s*<\/div>/im );

    return data;
}
// загрузка страниц нового СК без использования loadpage.php, 
// отсекаем лищние теги заголовка, сохраняем фон странички-вкладыша
function LoadPageToPane() {

    $.get( name_page + multilang + type_page, PutContentIntoPane );
    return false;
}
// загрузка из php-скриптов
function LoadPHPToPane( name_php, params ) {

    $.post( name_php, params, PutContentIntoPane );
    return false;
}
// загрузка Содержимого как есть без обработки (например, работа фильтров
function LoadContent( href, this_ref ) {
    DivContent.load( href, function() {
        /*   здесь отрабатываем замораживание картинки у нажатой кнопки, отмораживаем замороженные до этого */
        if (this_ref) // при условии, что нам передана ссылка на кнопку ("<a href>")
            ChangeButton(this_ref);
        else if ( $('a[href*="' + name_page + type_page + '"]').length > 0 ) // если нет  пробуем найти по адресу перехода
            ChangeButton( $('a[href*="' + name_page + type_page + '"]')[0] );
        SetDocumentHash();
        PostLoadOkno();
    });
    return false;
}
// заполнение содержимого при удачной загрузке, разделяют текст на составляющие, убирает лишнее
function PutContentIntoPane(data, status) {
    if (status == 'success' )
    {
        try {
            if (typeof data == 'object')
                data = data.toString();

            if( title = data.match(/<title>(.+)?(?=<\/title>)/i) )
                document.title = ( title_doc ? title_doc : title[1] );

            SetDocumentHash(); //меняем адресную строку ДО применения html, чтобы взять ресурсы из папок, если они есть
            DivContent.html( GetBodyBackground(data) );
        } catch(e) {
            alert(e.message);
        }
        PostLoadOkno();
    }
    else
        alert('Ошибка при загрузке страницы :' + status + '(' + name_page + type_page + ')' )
}
// спецобработки после заполнение содержимого (счетчики, обработка ссылок, история ссылок и т.п.
function PostLoadOkno( ) {
    "use strict";
    var wDiv, divFirst;

    AddClickShowOkno( DivContent );
    $('div.autoload').each( function () {
        var pane = $(this),
            URL  = pane.data('href') ? pane.data('href') : this.id + ( this.id.search('.php') > 0 ? '' : '.php' );
        if ( this.id == 'show_tovar') //пока временно специально для обработки выборки лотов
            $.get( 'show_tovar.php', function (data, success) { pane.html( GetBodyBackground(data) ); AddClickShowOkno( pane ); } );
        else
            $.get( URL, function (data, success) { pane.html( GetBodyBackground(data) ); AddClickShowOkno( pane ); } );
    });

    divFirst = $('div:first', DivContent);
    wDiv		= divFirst.width( );
    // divFirst.width( '99%' ).find('[style*="width:' + wDiv + '"]').width( '100%' ); 
    DivContent.addClass( 'body-pane' );

    if (typeAnimate  === undefined )
        DivContent.show( "slow", AfterShow );
    else if (typeAnimate == 'slide')
        DivContent.slideDown( "slow", AfterShow );
    else if (typeAnimate == 'fade')
        DivContent.fadeIn( "slow", 'linear', AfterShow );
    else
        DivContent.show( "slow", AfterShow );
}

// действия после показа Содержимого
function AfterShow() {

    $("#zag").html( title_doc ? title_doc : document.title );

    if ( document.title != title_doc )
        document.title = $("#zag").html();

    // проставляю заголовки для ЛЮБЫХ страниц
    $("#zag").html( '<h1>' + document.title + '</h1>' ); // это все немного странно, но нужно
    if ( NotZagl )
        $("#zag").hide().height(0);

    PostShow( name_page + type_page );

    if (anchor > '')
        GotoAncor( anchor );
    else if ( DivContent.hasClass('scrollTo') )
        $('body').scrollTo( DivContent, 800, {margin:true, offset: -(DivContent.attr('offset') || 0) }   );
    else if ( $(window).scrollTop() > DivContent.offset().top )
        $('body').scrollTo( DivContent, 800, {margin:true, axis: 'y'} );


    $('.captcha').attr( 'src', 'my_codegen260808.php?' + Date() );


    try
    {  // устанавливаем значение Яндекс-Метрики, если она подключена к сайту
        if (window.yaCounterXXXXXX)
            yaCounterXXXXXX.hit( name_page + type_page, document.title, document.domain);
        if ( window.SitePostShow !== undefined ) // если определена отдельная фукция для сайта либо странички
            SitePostShow();
    }
    catch(e) {
        alert(e);
    }
    // загружаем скрипт увеличения картинок, если есть подходящие по классу картинки

}
// старье
function LoadPageToDiv( name_page, str_ext, NameDiv, main_page ) {
    var DivForLoad = $( "#" + NameDiv );
    var str;

    if ( DivForLoad.length == 0 ) return;

    name_page = ( name_page.search('/') > 0 ? name_page.substring( name_page.search('/')+1 ): name_page );
    main_page = ( ( (main_page) && (main_page.search('/') > 0) ) ? main_page.substring( main_page.search('/')+1 ): main_page );

    if (DivForLoad.attr( 'rel' ) == name_page) {
        DivForLoad.slideUp( "slow" , function () {

            DivForLoad.html( '' );
            DivForLoad.attr( 'rel', '' );

        });
        return;
    }
    str = GetAjax( name_page, str_ext, multilang );

    if (!str) {
        if (DivForLoad.attr( 'rel' ) == main_page) { return; }
        str = GetAjax( main_page, ".htm", multilang );
        name_page = main_page;
    }

    DivForLoad.slideUp( "slow" , function () {

        DivForLoad.html( str );
        DivForLoad.attr( 'rel', name_page );

    });


    DivForLoad.slideDown( "slow",  function () {

        AddClickShowOkno( DivForLoad );

    } );
    return false;
};


function LoadPrice( id, zag ) {

    $("#zag").html( "Ждите, загружаю страницу" );
    DivContent.hide( "slow" , function () {
        DivContent.css( 'height', 730 );
        DivContent.load( 'http://' + document.domain + '/' +  id + '.htm',
            function () {
                $('div', DivContent).css( 'height', 700 );
                $( 'iframe', DivContent ).css( 'height', 700 );
                DivContent.show( "slow",  function () {
                    $("#zag").html( '<h2>' + zag + '</h2>' );
                    document.location.hash = id;
                    document.title = zag
                });
            });
    });
    return false;
}
