/*
 Набор стандартных инструментов для показа страниц со Стилями Артели
 Версия от декабря 2016 г.
 */
var menu_hide, DivPopMenu, go_history = 1, name_folder = ya_search = '', GlobalError;
$(function()   {

    if ( (typeof DivContent === "undefined") && ((DivContent = $("#pane")).length == 0) ) {
        DivContent = $('#content');
    }
})
// запоминание в куках, разные действия
/**
 * Get the value of a cookie with the given name.
 *
 * @example $.cookie('the_cookie');
 * @desc Get the value of a cookie.
 *
 * @param String name The name of the cookie.
 * @return The value of the cookie.
 * @type String
 *
 * @name $.cookie
 * @cat Plugins/Cookie
 * @author Klaus Hartl/klaus.hartl@stilbuero.de
 */
$.cookie = function(name, value, options) {
    if (typeof value != 'undefined') { // name and value given, set cookie
        options = options || {};
        if (value === null) {
            value = '';
            options.expires = -1;
        }
        var expires = '';
        if (options.expires && (typeof options.expires == 'number' || options.expires.toUTCString)) {
            var date;
            if (typeof options.expires == 'number') {
                date = new Date();
                date.setTime(date.getTime() + (options.expires * 24 * 60 * 60 * 1000));
            } else {
                date = options.expires;
            }
            expires = '; expires=' + date.toUTCString(); // use expires attribute, max-age is not supported by IE
        }
        // CAUTION: Needed to parenthesize options.path and options.domain
        // in the following expressions, otherwise they evaluate to undefined
        // in the packed version for some reason...
        var path = options.path ? '; path=' + (options.path) : '';
        var domain = options.domain ? '; domain=' + (options.domain) : '';
        var secure = options.secure ? '; secure' : '';
        document.cookie = [name, '=', encodeURIComponent(value), expires, path, domain, secure].join('');
    } else { // only name given, get cookie
        var cookieValue = null;
        if (document.cookie && document.cookie != '') {
            var cookies = document.cookie.split(';');
            for (var i = 0; i < cookies.length; i++) {
                var cookie = jQuery.trim(cookies[i]);
                // Does this cookie string begin with the name we want?
                if (cookie.substring(0, name.length + 1) == (name + '=')) {
                    cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
                    break;
                }
            }
        }
        return cookieValue;
    }
};
// изменения для магазинов - если присутствуют волшебные слова
function GetStoreScript(str_path) {
    var arr_store = Array( 'product', 'section', 'new_tovar', 'spec' ),
        arr_path  = Array( 'full_lot.php?key_tovary=',
            'show_tovar.php?key_parent=',
            'show_tovar.php?new_tovar=1',
            'show_tovar.php?spec=1' );


    for (var i in arr_store)
        if ( (pos=str_path.search( arr_store[i] )) > -1)
            return arr_path[i] + str_path.substring( pos + arr_store[i].length+1 );

    return str_path;

}
// смена странички-вкладыша при изменении хеша веб-строки
function ShowOknoFromHash() {
    var str_hash = document.location.hash.substring(1),
        str_path = document.location.pathname,
        str_search=document.location.search,
        i = str_hash.search('.php'),
        pane_rel = DivContent.attr( 'rel' ),
        file_ext;

    console.log(str_path);

    if ( ( (str_hash > "") && (pane_rel == str_hash) )
//     	|| ( (pane_rel == 'text_main.htm') && !(str_hash) )
        || ( str_path == pane_rel ) ) //убираем лидирующий /
        return;
    try
    {
        if ( (str_search > "")  && ( str_path.search('sitesearch') > -1) ) // для поиска Яндекса
        {
            if ( str_path.search('php') == -1 )
            {
                DivContent.load( 'http://' + document.domain + '/find.htm' + str_search
                    + ' #yandex-results-outer' );
                BeginSearch();
            }
        }
        else if (str_hash == '')
        {
            if ( (str_path == '') || (str_path == '/') || (str_path == root_page) ) {
                str_path = default_page
            }

            button = $('[data-href="' + str_path + '"], a[href="' + str_path + '"], a:has(img[alt="' + str_path + '"])');

            if ( button.length > 0 )
                button[0].click();
            else
                ShowOkno( ( str_path && ( str_path.search( root_page ) < 0 ) ? ( str_search ? str_path + str_search : str_path ) : (default_page || 'text_main') ), '');
        }
        else if (str_hash == '404')
            DivContent.load( '404.shtml' );
        else if (str_hash == 'admin')
        {
            $( '<div id=admin> <div>' ).appendTo( $('body') );
            if ( (default_page) && ( $( 'a:has(img[alt="' + default_page + '"])').length == 1 ) )
                $( 'a:has(img[alt="' + default_page + '"])').click();
            else
                ShowOkno( default_page || 'text_main', '');


            location.hash = 'admin';

// 			SetDocumentHash( 'admin' );
        }
        else if ( $( 'a:has(img[alt="' + str_hash + '"])').length == 1 ) // показ гуглдокументов из кнопок (для текстов сделать позже)
            $( 'a:has(img[alt="' + str_hash + '"])' ).click();
        else if ( str_hash.substring( 0, 10 ) == 'googledoc-' )
        {

            if ( (this_ref=$( 'a[href*="' + (doc_google=str_hash.substring( 10 )) + '"]' )).length > 0 )
                this_ref[0].click();
            else
                DivContent.attr( 'rel', str_hash ).addClass( 'show_div' ).html( '<iframe id=gdocs style="width:100%; height:' + GetHeightDivContent() + 'px;" frameborder=0 allowtransparency seamless marginheight="0" marginwidth="0" src="https://docs.google.com/document/pub?id='
                    + str_hash.substring( 10 )
                    + '&amp;embedded=true"></iframe>' );
        }
        else {

            if (str_hash.substring(0,1) == '!') {
                file_ext = ( i > -1 ? str_hash.substring( i ) : '.php' );
                str_hash = ( i > -1 ? str_hash.substring( 1, i ) : str_hash.substring(1) );
                ShowOkno( str_hash, '', '', file_ext);
            }
            else if (i > -1)
                ShowOkno( str_hash.substring(0, i), '', '', str_hash.substring( i ) );
            else
                ShowOkno( str_hash, '');
        }
    }
    catch(e) {
        console.log(e, str_hash, str_path );
    }
}
// смена адресной строки с предотвращением перезагрузки Содержимого
function SetDocumentHash( ) {
    // обрезаю доменное имя и меняю скрипты для магазинов, затем готовлю полный путь для записи в Хистори браузера
    var str_path = GetShortURL( GetStoreScript( name_page + type_page ) ),
        origin   = document.location.protocol + '//' + document.location.host + ( str_path[0] == '/' ? '' : "/" )
            + ( ( str_path != root_page ) && (str_path != default_page) ? str_path : '' );

// 	document.location.hash = new_hash;
    if ( (go_history)  ) {

        window.history.pushState( str_path, title_doc, origin  );
        console.log(str_path);
    }
    go_history = 1;
}
// Эта функция отрабатывает при перемещении по истории просмотром (кнопки вперед-назад в браузере)
function MyPopState(event) {
    if ( (go_history == 0) || (event.state == null) /* || (str_hash == DivContent.attr( 'rel') ) */ || (ya_search == 'process') )
        return true;
    go_history = 0;
    ShowOkno( event.state );
}
// какой хеш будем записывать - теперь будем менять на запись в истории
function GetHashFormPage(thisalt) {
    var str_hash = document.location.hash.substring(1),
        str_path = document.location.pathname;

    console.log(str_path)
    // если это не страница дизайна и есть какой-то путь - не ставим хеш
    return ( (!str_hash && str_path && ( str_path.search( root_page )) < 0 ) ? '' :
        (thisalt == default_page) || (thisalt == 'text_main.htm') ? '' : ( thisalt.search( '.php' ) > -1 ? '!' + thisalt : (thisalt.search( '.htm' ) > -1 ? thisalt.substring( 0, thisalt.search( '.htm' ) ) : thisalt) ) );
}
// постобработка при загрузке содержимого
function PostShow(thisalt) {
    var postDiv = '<div id="postDiv" style="height: 10px; width: 100%; float:left;"></div>'; // для удлинения странички и убирания прокрутки

    DivContent.attr( 'rel', thisalt ).addClass( 'show_div' ).children().last().after( postDiv );
    // отключаем перегрузку Содержимого на изменение хеша ибо сами его меняем
//   SetDocumentHash();
}

// средства показа бублов-всплывающих диалогов
var bubble, timer, id = 0,
    idleTimer = null,
    idleState = false, // состояние отсутствия
    idleWait = 5000; // время ожидания в мс. (1/1000 секунды)
function ShowBubble() {

    idleState = false;
    idleWait = 2000;
    clearTimeout(idleTimer); // отменяем прежний временной отрезок
    if ( ($.scrollTo !== undefined) && ($(window).scrollTop() > $(bubble[id]).parent().offset().top) )   //прокручиваем до начала страницы, если подключен соответствующий скрипт
        $('body').scrollTo( 0, 800, {margin:true, axis: 'y'} );

    if (id > 1)
        $(bubble).each( function(idx) {
            if (idx==id)
                return false;
            if ( $(this).css('top') <= $(bubble[id]).css('top') )
                $(this).slideUp( "fast" ); // скрываем пузырек, по высоте потому что там реализован диалог
            idleState = true;
        });

    $(bubble[id]).slideDown( "slow", function () {   // показываем очередной пузырек
        if (++id == bubble.size() ){    // если он крайний - прерываем дальнейшие показы
            clearInterval(idleTimer);
            // кусок ниже только для сайта Артели, потом подумаю - как его поправить для общего случая
// 					  $('body').scrollTo( $('img[alt="action"]'), 1000, {margin:true, offset: -200} );
            for(i=0;i<100;i++)
                $( 'a:has(img[alt="action"])' ).animate( { opacity: 0.1 } ).animate( { opacity: 1 } );
        }
    });
    idleTimer = setTimeout( "ShowBubble();", idleWait ); //  продолжаем смену
}
// устраиваем смену пузырьков диалога
function CreateDialog() {
    bubble = $('p.bubble');

    $(document).bind('mousemove keydown scroll', function(event){ // на любое действие пользователя (мышь, клавииатура, скроллинг)
        clearTimeout(idleTimer); // отменяем прежний временной отрезок
        if(idleState == true){
            // Действия на возвращение пользователя - прекращаем показ совсем
            if ( (event.type == 'scroll') && ($(window).scrollTop() == 0) && (id > 1) )
                ShowBubble(); // снова запускаем смену
            return;
        }

        idleState = false;
        idleTimer = setTimeout( "ShowBubble();", idleWait ); //  запускаем смену
    });

    $("body").trigger("mousemove"); // сгенерируем ложное событие, для запуска скрипта

}

// запись в куках
function cookieSet(name, value) {

    if ($.cookie === undefined)
        console.log('$.cookie === undefined');
    else
        $.cookie( name, value, {expires: null, path: '/'}); // Set mark to cookie (submenu is shown):

}
// получение значения
function cookieGet(name) {
    if ($.cookie === undefined)
        console.log('$.cookie === undefined');
    else
        return $.cookie( name );
}
function cookieDel(name) {

    if ($.cookie === undefined)
        console.log('$.cookie === undefined');
    else
        $.cookie( name, null, {expires: null, path: '/'}); // Delete mark from cookie (submenu is hidden):

}


// cсоздание объекта для чтения страницы из Сети кроссбраузерно
function createRequestObject() {
    if (typeof XMLHttpRequest === 'undefined') {
        XMLHttpRequest = function() {
            try { return new ActiveXObject("Msxml2.XMLHTTP.6.0"); }
            catch(e) {}
            try { return new ActiveXObject("Msxml2.XMLHTTP.3.0"); }
            catch(e) {}
            try { return new ActiveXObject("Msxml2.XMLHTTP"); }
            catch(e) {}
            try { return new ActiveXObject("Microsoft.XMLHTTP"); }
            catch(e) {}
            throw new Error("This browser does not support XMLHttpRequest.");
        };
    }
    return new XMLHttpRequest();

}

// удаление обработчика события у елемента
function removeEvent(elem, type, handler){

    if (elem.removeEventListener)
        elem.removeEventListener(type, handler, false);
    else
        elem.detachEvent("on"+type, handler);
}

// загрузка скриптов через AJAX
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
                alert('Неудачная загрузка скрипта "' + url + '"! Перегрузите страницу!');
        }
    });

}

// похоже, эта функция уже тоже устарела, т.к. вынесены функции теперь в отдельные js-файлы
function GetPreloadImages( str, str_addon ) {

    i = str.search( "preload" );
    if (i == -1) // нечего обрабатывать
        return '';

    str.replace( /sc-pro\/preload\d+\.js/i, function (str) { LoadJScript( str, true, true ); i = -1; return str; } );
    if (i == -1) // есть вызов файла типа preload0121.js
        return '';

    j = str.search( "//-->" );
    if (j > -1) { // подключаю предзагрузку картинок, если она прописана ПРЯМО В ТЕКСТЕ страницы СК (устарело)
        $.getScript( "sc-pro/buttons.js" );
        str_preload = '<script type="text/javascript"> $(function() {'
            + str.substring( i+19, j-4 ).replace( /\s/gm, ' ' ).replace( /"b\d+/gi, "$&_" + str_addon  ) + ' })'
            +  '</script>';


        if (name_folder > '')
            str_preload = str_preload.replace( /sc-pic(?=\/)/gi, name_folder + "$&" );

    }
    else {
        str_preload = '';
    }

    return str_preload;
}

function GetUserCode(str) {
    var begin_code = "<!-- user defined code -->",
        end_code = "<!-- end of user defined code -->";

    i = str.search( begin_code ) + begin_code.length;
    j = str.search( end_code );
    if ( (i > -1) && (j > i) ) { // подключаю загрузку пользовательских кодов
        str_code = str.substring( i, j );
        i = str_code.search('document.location.pathname');
        j = str_code.search( "window.location.assign");
        if ( (i > -1) && (j > -1) ) // есть код страницы-вкладыша, надо попозже поточнее рассчитать
            return '';

        return str_code;
    }

    return '';

}

function ReplaceSCText( str, str_addon, name_folder ) {
    var str_result = '';


    i = str.search( "body { background" );
    if (i > -1) {
        j = str.search( "</style>" );


        str_result = " .body-pane" + str.substring( i+4, j+8 );

        str_result = '<style type="text/css"> \n /* Стили Артели для Сайткрафт версии 6.5 и выше */ \n' + str_result ;
    }

    i = str.search( "<div" );
    j = str.search( "</body>" ); // - 14;
    str_body =  str.substring( i, j );
    if (name_folder > '')
        str_body = str_body.replace( /"sc-pic(?=\/)/gi, '"' + name_folder + "sc-pic" ).replace( /"download(?=\/)/gi, '"' + name_folder + "download" ).replace( /href="(\S+)(?=\.htm")/gi, 'href="' + name_folder + "$1" );

    str_preload = GetPreloadImages( str, str_addon );
    if (str_preload > '') {
        str_result +=  str_body.replace( /\b(name|id)="b\d+\b/gi, "$&_" + str_addon  ).replace( /'b\d+/gi, "$&_" + str_addon  ) + str_preload;

    }
    else { str_result += str_body; }

    return  GetUserCode(str) + str_result;

}

function GetShortURL(m_adress) {
    var origin   = document.location.protocol + '//' + document.location.host + '/',
        i = m_adress.search( origin );

    if (i > -1)
        m_adress = m_adress.substring( i + origin.length );

    return m_adress;
}

// Обработка страниц Сайткрафта
function GetAjax( m_adress, file_ext, multilang ) {
    var Res, str_body, str_preload, str, j, i,
        str_addon = ExtractPageName( m_adress ),
        str_location = document.domain;

    m_adress = GetShortURL(m_adress);

    if (m_adress[0] == '/')
        m_adress = m_adress.substring(1);

    i = m_adress.search('/') + 1;

    if (i > 0) // добавляем имя папочки, если она есть!
    {
        name_folder = m_adress.substring( 0, i );
        str_location += ('/' + name_folder);
        m_adress  = m_adress.substring( i );
    }
    else
    {
        name_folder = '';
        i = location.pathname.substring(1).search('/') + 2;
        str_location += (i > 1 ? location.pathname.substring( 0, i ) : '/' );
    }

    j = m_adress.search( '.htm' );
    if (j == -1)
        j = m_adress.search( '.php' );

    if (j > 0)
    {
        file_ext = m_adress.substring( j );
        m_adress = m_adress.substring( 0, j );
    }
    // поддержка многоязычности
    isAdmin  = html_multilang = '';
    preAdmin = ( file_ext.search( /\.php\?/ ) > -1 ? '&' : '?' );

    if ( multilang )
    {
        if (file_ext == ".htm" )
            html_multilang = multilang;
        else
            isAdmin = preAdmin + 'multilang=' + multilang;
    }
    // если находимся в админке, добавляю параметр admin
    if ( ( $('#admin').length > 0 ) && ( file_ext.search( '.php' ) > -1 ) )
    {
        isAdmin += ( isAdmin ? '&' : preAdmin ) + 'admin';
    }



    try {

        Res = createRequestObject();

        Res.open("GET", 'http://' + str_location + m_adress  + html_multilang + file_ext + isAdmin , false );
        Res.setRequestHeader('Content-Type', 'text/html; charset=windows-1251');
        Res.send( null );

        if ( Res.status == 404 ) {
            if (m_adress == "404")
                return 'Страница не найдена!';
            else if (file_ext == '.php')
                return GetAjax( '404', ".shtml" );
            else
                return GetAjax( m_adress, '.php' );
        } // данные не загружены

        if ( Res.readyState != 4 )
            return 'не удалось загрузить данные!';

        if (file_ext == ".shtml")
            file_ext = ".htm";
        else if ( Res.status != 200 )
            return "Ошибка №" + Res.status + " при загрузке страницы " + m_adress ;  // статус не ОК

        str = Res.responseText;
        title_doc = GetTitle(str) || title_doc;

        if (str.search( "/* WEBaby Corp., Sitecraft") > -1 ) /*  обработка СК-страниц ДО версии Сайткрафт-Студии 6.5 */
        {

            i = str.search( "body { background" );
            if (i == -1) {
                i = str.search( "body { }" );
            };

            j = str.search( "</style>" );



            str = " .body-pane" + str.substring( i+4, j+8 ).replace(/\.[apt]\d+/gi, "$&_" + str_addon  );

            str = '<style type="text/css"> \n <!-- /* Стили Артели для Сайткрафт версии 4.31 */ \n' + str ;
            i = Res.responseText.search( "<!-- user defined code -->" );
            j = Res.responseText.search( "<!-- end of user defined code -->" );
            if ( (i > -1) && (j > -1) ) { // подключаю загрузку пользовательских кодов
                str_code = Res.responseText.substring( i, j );
                i = str_code.search('document.location.pathname');
                j =str_code.search( "window.location.assign");

                if ( (i > -1) && (j > -1) ) // есть код страницы-вкладыша, надо попозже поточнее рассчитать
                    str_code = str_code.substring(0, j) + ' } </script>';

                str += str_code;
            }



            i = Res.responseText.search( "preload" );
            j = Res.responseText.search( "//-->" );
            if (i > -1) { // подключаю предзагрузку картинок
                $.getScript( "sc-pro/buttons.js" );
                str_preload = '<script type="text/javascript">\n'
                    + ' $(function() {'
                    + Res.responseText.substring( i+19, j-4 ).replace( /\s/gm, ' ' ).replace( /"b\d+/gi, "$&_" + str_addon  ) + '\n  })'
                    +  '\n</script>';


                if (name_folder > '')
                    str_preload = str_preload.replace( /sc-pic(?=\/)/gi, name_folder + "$&" );
                // sc-pic/

            }
            else {  str_preload = '';  }

            i = Res.responseText.search( "<div" );
            j = Res.responseText.search( "</body>" ); // - 14;
            str_body =  Res.responseText.substring( i, j ).replace( /(\s|")[apt]\d+(?=\s|")/gi, "$&_" + str_addon  );
            if (name_folder > '')
                str_body = str_body.replace( /"sc-pic(?=\/)/gi, '"' + name_folder + "sc-pic" ).replace( /"download(?=\/)/gi, '"' + name_folder + "download" );

            if (str_preload > '') {
                str +=  str_body.replace( /\b(name|id)="b\d+\b/gi, "$&_" + str_addon  ).replace( /'b\d+/gi, "$&_" + str_addon  ) + str_preload;

            }
            else { str += (str_body /* + ' </div>' */); }


        } // завершение обработки СК-страниц ДО версии Сайткрафт-Студии 6.5
        else if (str.search( /<link rel="stylesheet" type="text\/css" href="(..\/)*sc-pro\/sc.css">/) > -1 )
        // обработка СК-страниц версии Сайткрафт-Студии 6.5 и выше
        { if (file_ext.search('.php') == -1) //временно не трогаем php
            str = ReplaceSCText( str, str_addon, name_folder );
        }
    }
    catch( e ) {

        alert( e );
        str = e.description;
    }
    return str;
}

function Make3Cols( ImgLogo ) {
    var DivImg = ImgLogo.parent().parent(),
        DivWdt = parseInt( DivImg.width() ) / 2,
        DivEmp = '<div class="completeRezina" style="float: left; width: ' + DivWdt + 'px;"></div>',
        DivPrev = DivImg.prev(), DivNext = DivImg.next() ;

    DivImg.addClass('completeRezina');
    DivPrev.addClass('completeRezina').css( {  width: '', 'margin-right': DivWdt } ).wrap('<div class="completeRezina" style="width: 50%; float: left; margin-right: -'+ DivWdt + 'px;"><div>').after( DivEmp );
    DivPrev.children('div[style*=width]').each( function() {
        RezinaWidth( $(this), $(this).width() );
    })

    DivNext.addClass('completeRezina').css( {  width: '', 'margin-left': DivWdt } ).wrap('<div class="completeRezina" style="width: 50%; float: right; margin-left: -'+ DivWdt + 'px;"><div>').before( DivEmp );
    DivNext.children('div[style*=width]').each( function() {
        RezinaWidth( $(this), $(this).width()  );
    })
    //parent().addClass('completeRezina');
}

function RoundPre2( value ) {
    return /* new Number( value ).toFixed( 2 );  */Math.floor( value * 100 ) /100;
}
function RezinaWidth(ch_div, parent_width) {
    var //parent_width = parent_width || parseInt( ch_div.parent().css( "width" ) ) || 998,
        str_style = ch_div.attr('style'), ch_width, ch_div_width, min_width, fMargin, margin_left, fPadding, pad_left;

    if ( str_style )
    {
        if ( str_style.match( /width:(\s)*(\d+)%/ ) ) // указан размер - ничего не меняем
            return;

        ch_width = str_style.match( /width:(\s)*(\d+)px/ );
        ch_div_width = parseInt( ch_width ? ch_width[2] : 0, 10);
        min_width = str_style.match( /min-width:(\s)*(\d+)px/ );
        fMargin = str_style.match( /margin-left:(\s)*(\d+)px/ );
        margin_left = parseInt( fMargin ? fMargin[2] : 0, 10 );
        fPadding = str_style.match( /padding-left:(\s)*(\d+)px/ );
        pad_left = parseInt( fPadding ? fPadding[2] : 0, 10 );

    }
    else
        ch_div_width = min_width = pad_left = margin_left = 0;

    if ( (ch_div.css('width').search('%') > 0) /* || !ch_width */  || (ch_div.hasClass('completeRezina'))  /* || (ch_div.hasClass('fullW')) */ || (ch_div.hasClass('slide-pane')) || (ch_div.hasClass('pop-pane')) || (ch_div.parents('.pop-pane').length > 0)  )
        return;


    if (pad_left  > 0) // с min_width  разобраться позже как будет время
    {
        pad_left = RoundPre2( (pad_left * 100) / parent_width ); // пока держимся одной ширины
        ch_div.css( { 'padding-left' : pad_left + '%'} );
    }

    if ( margin_left > 0) // с min_width  разобраться позже как будет время
    {
        margin_left = RoundPre2( (margin_left * 100) / parent_width ); // пока держимся одной ширины
        ch_div.css( { 'margin-left' : margin_left + '%'} );
    }

    childPNG32 =  ch_div.children('.png32:not(.fullW):has(.noresize)'); // не трогать картинки, на которых лежат заготовки с классом noresize
    if ( ( (childPNG32.length > 0) && (childPNG32.find('div.png32').length == 0) )
        || ( ch_div.children('img.noresize, img[alt*=noresize]').length == 1) || (ch_div.attr('title') == 'noresize' )
    )
    {
        ch_div.wrap('<center></center>');
        console.log( ch_div, str_style, ch_div_width, parent_width, RoundPre2( (ch_div_width * 100) / parent_width ) + '%', ch_div.css( 'width' ) )

        return;
    }
    if ( (ch_div.children('.noresize').length > 0) || ch_div.hasClass('noresize') ) // строим двухколоночный макет
    {
        console.log( ch_div, str_style, ch_div_width, parent_width, RoundPre2( (ch_div_width * 100) / parent_width ) + '%', ch_div.css( 'width' ) );

        if ( ch_div.next('div').children('#pane').length == 1)
        {
            new_margin = ( ch_div.hasClass('noresize') && ch_div_width ? ch_div_width : ch_div.children('.noresize').width() );
            ch_div.next('div').css({ width:'100%', float: 'right', 'margin-left': -new_margin + 'px' })
                .prepend('<div class="completeRezina" style="float: left; height: 1px;width:' + new_margin + 'px;"></div><div class="completeRezina" style="float: left; height: 1px; width: 50%; margin-right:-'+ DivContent.width() / 2 + 'px;"></div>');
// 		   DivContent.css({ width:'', 'margin-left': new_margin }).wrap('<center></center>');
            return;
        }

        return;
    }

    if ( ch_div_width > 0 ) // меняем ширину тем, у кого показан реальный размер
    {
        if ( min_width  )	    // устанавливаем мин. ширину, только если была запись в самом элементе  (от СК)
            ch_div.css( { "min-width": min_width[2] } );
        else
            ch_div.css( { "min-width": ch_div_width } );

        if ( ch_div_width < parent_width )
            ch_div.css( { 'width': RoundPre2( (ch_div_width * 100) / parent_width - margin_left - pad_left*2 ) + '%' } );
        else
            ch_div.css( { 'width': 100 - margin_left - pad_left*2 + '%' } );
    }
    else
        ch_div_width = parent_width;

    ch_div.addClass('completeRezina');
//   parent_width = ch_div_width;

    ResizeChildren( ch_div, ch_div_width );
    ResizeChildren( ch_div.find('div:not(.slide-pane):not([style*=width]):has(div[style*=width]):first'), ch_div_width );
//   ch_div.children('img, a:has(img):not(.fullW, img[alt*=fullW])').wrap('<center></center>');

    /*
     if ( margin_left > 0)
     console.log( ch_div, str_style, ch_div_width, parent_width, RoundPre2( (ch_div_width * 100) / parent_width ) + '%', ch_div.css( 'width' ) );
     */
}

function ResizeChildren( ch_div, parent_width ) {
    var   isNeedResize = 'div[style*=width]:not(.completeRezina)';

    ch_div.children( isNeedResize ).each( function () { RezinaWidth( $(this), parent_width ) });

}

function SetFullW( parent_this ) {
    var divFirst;

    if ( $( '.fullW, img[alt*=fullW]', parent_this ).length > 0)
        RezinaWidth( $('div:first', parent_this), 998 );

    $( 'img[alt*=fullW]', parent_this  ).each( function() {

        if ( !($(this).css( "min-width" ) > '0px') && ($(this).css( "width" ) > 0) )
            $(this).css( { "min-width" : $(this).width() } );

        $(this).css( { width: '100%' } )

            .parentsUntil('[style*="float: left"],[style*="margin: auto"],[style*="margin-top: auto"], body').each(
            function() {
                if ( $(this).height() > 0)
                    $(this).css( { "min-height" : $(this).height() } );

                $(this).height('auto');
            } );
    } );

    $( ".fullW", parent_this  ).parents('div').each( function() {

        if ( $(this).height() > 0)
            $(this).css( { "min-height" : $(this).height() } );

        $(this).height('auto');
    });

}

function SetElementPosition( parent_this ) {
    $( "img[alt*=button]", parent_this ).each( function () {
        $(this).parent('a').addClass('button');
        this.alt = this.alt.replace(/button/, '');

    }); // старье
// Привязка фиксированных элементов - картинок(картинки с наложенными элементами определяются по заготовке) и кнопок с alt=fixed либо bottom,top,left,right для прилепления к краю страницы
    $( "[alt*=fixed]", parent_this ).css( { "position" : "fixed" } );
    $( ".fixed", parent_this ).closest('div[style*=background]').css( { "position" : "fixed" } );
    $( "[alt*=top]", parent_this ).addClass('topfixed');
    $( ".top", parent_this ).closest('div[style*=background]').addClass('topfixed');

    $( "[alt*=left]", parent_this ).css( { "position" : "fixed", left: 0 } );
    $( ".left", parent_this ).closest('div[style*=background]').css( { "position" : "fixed", left: 0 } );
    $( "[alt*=bottom]", parent_this ).css( { "position" : "fixed", bottom: 0 } );
    $( ".bottom", parent_this ).closest('div[style*=background]').each( function () {
        if ($(this).parent('div[style*=border]').length > 0 )
            element = $(this).parent('div[style*=border]'); //.css( { "position" : "fixed", bottom: 0 } ).after('<div style="height:' + $(this).height() + 'px"> </div>');
        else
            element = $(this);

        element.css( { "position" : "fixed", bottom: 0 } ).after('<div style="height:' + element.height() + 'px"> </div>');

        if ( element.attr('style').search('width:') < 0 ) // если ширина не выставлена в СК - ставим 100%
            element.css( "width", "100%" );
    });

    $( "[alt*=right]", parent_this ).css( { "position" : "fixed", right: 0 } );
    $( ".right", parent_this ).closest('div[style*=background]').css( { "position" : "fixed", right: 0 } );

    // сглаживаем уголки
    $( ".border-radius", parent_this  ).each(  function () {
        $(this).closest('div[style*=background]').css( { "border-radius" : $(this).css( "border-radius"), "-moz-border-radius" : $(this).css( "border-radius" ) } )
            .parent('div[style*=border]').css( { "border-radius" : $(this).css( "border-radius" ), "-moz-border-radius" : $(this).css( "border-radius" )  } );
        $(this).parent('div[style*=border]').css( { "border-radius" : $(this).css( "border-radius" ), "-moz-border-radius" : $(this).css( "border-radius" )  } );
    });

    if ( $("#okno").hasClass('border-radius') )
        DivContent.css( { "border-radius" : $("#okno").css( "border-radius") } )


    if ( !parent_this.hasClass('noresize') ) // не растягивать элемент, если стоит соответствующий класс
        SetFullW( parent_this ); // резиновость

    // оставляем наверху или в подвале спецэлементы, подключаем родительские элементы с фоном к прокрутке
    $( ".notHidden:not([style*=background]):not([style*=border])", parent_this ).each( function() { SetParentFixedElements ( $(this), 'notHidden' )  });
    $( ".bottomHidden:not([style*=background]):not([style*=border])", parent_this ).each( function() { SetParentFixedElements ( $(this), 'bottomHidden' )  });

    // для картинок alt заменяем на классы (тем более, что скоро в СК можно будет их ставить
    $( "[alt*=notHidden]", parent_this ).addClass( 'notHidden' );
    $( "[alt*=bottomHidden]", parent_this ).addClass( 'bottomHidden' );
    $("[alt*=hide]").addClass( 'hide' );

    // скрываем элементы класса 'hide'
    $('.hide').hide();

    // определяем элементы, которые всегда должны оставаться на экране при прокрутке окна вверх либо становиться видимыми при данном событии
    $(window).bind('scroll', function(){
        var hhead = $('.notHidden').each( function()
            { SetFixedElements( $(this), 'topfixed' ); return true; }),
            hfooter = $('.bottomHidden').each( function()
            { SetFixedElements( $(this), 'bottomfixed' ); return true; });

    });
    $("a[title*='text-shadow']", parent_this ).each( function () {
        var title_param = this.title;
        $(this).parent('div').css( { // '-webkit-transform': metod_tranf.replace(/(\S+){(\S+)}/, '$2'),
            'text-shadow': title_param.replace(/(text-shadow){(.+)}/, '$2') } );
        this.title =   title_param.replace(/(text-shadow){(.+)}/, '') ;
    });
}

function SetParentFixedElements ( elem, class_name ) {
    parent_back = elem.closest('div[style*=background]');

    if ( parent_back.length == 0)
        parent_back = elem.closest('div[style*=border]'); // если не находим никого с фоном, то ищем с бордюром

    if ( parent_back.length > 0)
    {
        parent_back.addClass( class_name );
        elem.removeClass( class_name );
    }


}
// определяем элементы, которые всегда должны оставаться на экране. Зоной контроля будет Заготовка СОДЕРЖИМОЕ - как только она уходит наверх или вниз, меняет расположение таких элементов либо специальная заготовка, содержащая элемент с id="scrlzone"
function SetFixedElements( catalog_pane, class_name ) {
    /* 	 var catalog_alt = catalog_pane.attr('alt'); */
    var scrlzone = $('#scrlzone, #pane');
    /*      alert($(window).scrollTop()); */
// закрепляем элемент на экране, если прокрутка окна превышает верхнюю гшраницу СОДЕРЖИМОГО, но только при условии, что ширина окна БОЛЬШЕ, чем ширина элемента
    if ( ($(window).scrollTop() > scrlzone.offset().top ) && ($(window).width() > catalog_pane.width()) )
    {
        if ( catalog_pane.hasClass( class_name ) )
        {} // ничего не делаем, уже класс присвоен раньше
        else {
            catalog_pane.css('width', catalog_pane.width()+'px');
            if ( (catalog_pane.prev().length == 0) && (catalog_pane.next().length == 0) )
                catalog_pane.parent().height( catalog_pane.height() );
            catalog_pane.addClass( class_name ).show();
        }
    }
    else
    {
        catalog_pane.removeClass( class_name );
        if ( catalog_pane.hasClass('hide') )
            catalog_pane.hide();
    }
}

function GetHTTPSSource(){
    var str_div='<iframe id=gdocs style="width:100%; height:' + ( ($('#okno').height() || DivContent.height() ) - 5) + 'px;" frameborder=0 allowtransparency seamless marginheight="0" marginwidth="0" src="',

        thisalt = this.alt || 'https-' + this.href.substr( this.href.search("https") + 8 );

    str_div += this.href + '&amp;embedded=true&output=html"></iframe>';

    HtmlToPane( str_div, thisalt );

    return false;
}

function OldChangeButton(this_ref) {

    Button_ref = $( '.down_button');
    if (Button_ref.length > 0 ) {
        Button_ref.removeClass( 'down_button' )
            .each( function () {
                this.onmouseout = function onmouseout(event) { over_off(); } } );

        Button_img = Button_ref.children();
        Button_img.attr( 'src', Button_img.attr( 'rel' )  );
    }

    if ( $(this_ref).children().is('img') )
    {	Button =   $(this_ref).children();
        Button_name = Button.attr( 'id' );


        if ( (Button_name) && ( document.wb_pre) && ( document.wb_pre[Button_name] ) )
        {

            if (document.wb_normal) {
                Button.attr( 'rel', document.wb_normal );
            }
            this_ref.onmouseout = function onmouseout(event) {  };
            $(this_ref).mouseout( );
            $(this_ref).addClass('down_button');
        }
    }
}

// получаем минимальную высоту для встроенных фреймов - чтобы не сжимались в полоску и не растягивалис сильно
function GetHeightDivContent() {

    return ( (DivContent.height() || (DivContent.hasClass('pane-fullH') ? parseInt( DivContent.css('min-height') ) : DivContent.height() ) || DivContent.parent().height()  || DivContent.parent().parent().height() ) - 5);
// берем сперва окно(если есть), потом минимальную высоту, если Пане растягивается, иначе полную высоту, если все этого нет - ищем высоту по родителям
}

function GetGoogleDoc(){
    var str_div='<iframe id=gdocs style="width:100%; height:' + GetHeightDivContent() + 'px;" frameborder=0 allowtransparency seamless marginheight="0" marginwidth="0"src="',

        point_pub = this.href.search(/pub\?/), point_document = this.href.search('document'),
        thisalt = this.alt || 'googledoc-' + ( point_pub > -1 ? this.href.substr( point_pub + 7 ) : this.href.substr( point_document + 11, this.href.search("/pub") - point_document - 11 ) ),
        end_str = '"></iframe>';

    if ( $('#admin').length > 0 ) // режим администрирования
    {
        if (point_document > -1)
            str_div +=  ( point_pub > -1 ? this.href.replace( "pub?id=", "d/") :  this.href.replace( "/pub", "") ) + '/edit' + end_str;
        else
            str_div +=  this.href.replace( "pub?key=", "ccc?key=").replace('&output=html', '') + end_str;
    }
    else
        str_div += this.href + ( this.href.search('edit') > -1 ? '' : ( point_pub > -1 ? '&amp;' : '?') + 'embedded=true&amp;widget=true' ) + end_str;

    //HtmlToPane( str_div, thisalt );
    name_page = '/';
    type_page = '#' + thisalt; //обнуляем адресную строку и включаем хеш
    if (DivContent.length == 0)
        window.location.assign( root_page + '#' + thisalt );
    else
        DivContent.animate( { opacity: 0 }, function () {
            DivContent.html(str_div).animate(  { opacity: 1 },
                function () { PostShow(thisalt); } ).css( { 'width' : '100%'} ); // постобработка при загрузке содержимого
        });

    ChangeButton(this);

    return false;
}

function LoadHTTPSPage( href ) {
    var str_div='<iframe id=picasawebAlbom style="width:100%; height:' + GetHeightDivContent() + 'px;" frameborder=0 allowtransparency seamless marginheight="0" marginwidth="0"src="' + href + '" ></iframe>',
        thisalt = this.alt || 'picasawebAlbom';
    go_history = 0;
    HtmlToPane( str_div, thisalt ) ;

    return false;
}
function HtmlToPane( str_div, thisalt ) {

    DivContent.animate( { opacity: 0 }, function () {
        DivContent.html(str_div).animate(  { opacity: 1 },
            function () { PostShow(thisalt); } ); // постобработка при загрузке содержимого
    });
}

function GetPicasaAlbom(){ //https://picasaweb.google.com/111959787566500387373/hQeadB?authuser=0&feat=directlink

    var url_albom="return LoadHTTPSPage('"
            + this.href.replace( "directlink", "embedwebsite&noredirect=1" ) + "');",
        //https://picasaweb.google.com/111959787566500387373/hQeadB?authuser=0&feat=embedwebsite')",
        str_div='<table id=picasaweb style="width:100%;"><tr><td align="center" style="height:194px;"><a href="#" onclick="'
            + url_albom
            + '"><img src="https://lh3.googleusercontent.com/-XAXDmGBYEQs/UEZAZHJl5LE/AAAAAAAAABs/hTq_h_eXF_I/s160-c/hQeadB.jpg" width="160" height="160" style="margin:1px 0 0 4px;"></a></td></tr><tr><td style="text-align:center;font-family:arial,sans-serif;font-size:11px"><a href="#" onclick="'
            + url_albom
            + '" style="color:#4D4D4D;font-weight:bold;text-decoration:none;">Мой первый альбом</a></td></tr></table>',

        thisalt = this.alt || 'picasaweb';
    HtmlToPane( str_div, thisalt );

    // добавить сюда постобработку при загрузке содержимого
    return false;
}

function GetYoutube() {
    var str_div='<iframe id=yuotube style="width:100%; height:' + GetHeightDivContent() + 'px;" frameborder=0 allowtransparency seamless marginheight="0" marginwidth="0"src="' + this.rel + '" ></iframe>',
        thisalt = this.alt;

    HtmlToPane( str_div, thisalt );
    ChangeButton(this);

    return false;
}

function SetTitleAlt(Elemthis, this_img) {

    if (!Elemthis.alt)
        Elemthis.alt = ( this_img.is('img') ? this_img.attr( 'alt' ) || this_img.attr( 'title' ) : Elemthis.title );

    if (!Elemthis.title)
        Elemthis.title = ( this_img.is('img') ? this_img.attr( 'title' ) : $(Elemthis).text() || '');

}

function ExtractPageName(full_page_name) {
    return full_page_name.split("?")[0].split("/").slice(-1)[0];
}
// функции для проверки, подключения FancyBox и его плагинов в случае необходимости
var isFancyLoaded = false, target_blank;

function VerifyIncludeFancyBox( AfterFancyLoad ) {
    var
        isFancyNotLoad = ($.fancybox === undefined);

    if ( isFancyNotLoad && !isFancyLoaded ) // не загружен , загружаем немедленно
    {
        $('head').append('<link rel="stylesheet" type="text/css" href="http://solution.allservice.in.ua/js/fancybox2/jquery.fancybox.css" media="screen">');
        $('head').append('<script src="http://solution.allservice.in.ua/js/fancybox2/jquery.fancybox.js" />');
//      , true, true, AfterFancyLoad );
//       AfterFancyLoad();
        isFancyLoaded = true;
    }

    return !($.fancybox === undefined);
}
// подключение хелпов для слайд-шоу
function VerifyIncludeFancyHelpers() {
    if ( !VerifyIncludeFancyBox( VerifyIncludeFancyHelpers ) )
        return false;

    if ( ( $.fancybox.helpers === undefined ) || ( $.fancybox.helpers.buttons === undefined ) )
    {
        $('head').append('<link rel="stylesheet" type="text/css" href="http://solution.allservice.in.ua/js/fancybox2/helpers/jquery.fancybox-buttons.css" media="screen">');
        LoadJScript("http://solution.allservice.in.ua/js/fancybox2/helpers/jquery.fancybox-buttons.js", true, true, PutSlideShow );
        return false;
    }
    else
        PutSlideShow();
    return true;
}
// показ всплывающих окон через FancyBox
function MakePutFuncybox( ) {

    if ( VerifyIncludeFancyBox(MakePutFuncybox) )
        target_blank.each( PutFancyBox );

    return false;
}
// добавление плагинов для показа фотогалереи с слайд-шоу
function PutSlideShow() {
    var
        SlideClassName = 'fancybox-button';

    $('a:has(img.fancybox-button)').each( function () { var img = $('img', this ).removeClass( SlideClassName ), img_style = img.attr('style'), rel = img_style.replace(/rel(\S\s)+/i, '$1'); img.attr('style', img_style.replace(/rel(\S\s)+/i, '') ); $(this).addClass( SlideClassName ).attr( 'rel', rel ); this.title += ' (просмотр со слайд-шоу)';  } );


    $('a.' + SlideClassName ).fancybox({
        closeBtn : false,
        transitionIn	 : 'elastic',
        transitionOut	 : 'elastic',
        /*
         prefEffect: 'none',
         nextEffect: 'none',
         */
// 		title			: title || 'Для закрытия окна шелкните мышкой за его пределами!',
        helpers : {
            overlay : { showEarly  : false },
            title:  { type : 'none' },
            buttons: {
                position: 'bottom'
            }
        },
        afterLoad: function(current, previous) {
            console.info( 'Current: ' + current.href );
            console.info( 'Previous: ' + (previous ? previous.href : '-') );

            if (previous) {
                console.info( 'Navigating: ' + (current.index > previous.index ? 'right' : 'left') );
            }
        }
    });
}

function PutFancyBox( ) {
    var this_href = this.href, title = this.title ? this.title : $(this).children('img').attr( 'title' ), params;

    if (title > '')
        title = title.replace( 'откроется во всплывающем окне', '' ) ;

    if (this_href.search(/\.(jpg|png|jpeg|gif)/ig) > -1) // показ картинок
    {
        params = {
            padding: 5,
            closeBtn		: true,
            autoSize: true,
            autoResize: true,
            centerOnScroll : true,
            transitionIn	 : 'elastic',
            transitionOut	 : 'elastic',
            autoDimensions: true,
            overlayShow: true,
            title			: title || 'Для закрытия окна шелкните мышкой за его пределами!',
            helpers		: {
                title	: { type : 'float' }
            }
        };

    }
    else
        params = {
// 		href	: this_href,
            scrolling : 'auto',
            padding: 5,
            type : ( ShowIframe(this_href) ? 'iframe' : 'ajax'),
            ajax	: {dataType : 'html', content : 'text/html; charset=windows-1251'},
            autoSize: true,
            autoResize: true,
            /* 		width	  : 'auto', */
            centerOnScroll : true,
            transitionIn	 : 'elastic',
            transitionOut	 : 'elastic',
            title			: title || 'Для закрытия окна шелкните мышкой за его пределами!',
            //'content' 		: data,
            autoDimensions: true,
            overlayShow: true,
            helpers		: {
                overlay : { showEarly  : true },
                title	: { type : 'float'
                },
            }
        };

    /*
     if ( this.rel )
     $('a[rel=' + this.rel + ']').fancybox( params );

     else
     */

    if ( $.fancybox === undefined )
        $(this).click( function() { $(this).fancybox( params ); } );
    else
        $(this).fancybox( /* $.extend( */ params/* , { href	: this_href } ) */ );

    return true;

}

function AddClickToElement(data) {
    return AddClickShowOkno( $(this) );
}
// показ ссылки в диве pane
function ShowInPane(){
    var result = false;
    try
    {
        ShowOkno( this.href, this.title, this  );
    }
    catch(e)
    {
        result = true;
        GlobalError = e.message;
    }
    return result;
}
function ShowModal() {

    $.fancybox( {
        href	: this.href,
        scrolling : 'auto',
        padding: 5,
        type : 'ajax',
        ajax	: {dataType : 'html', content : 'text/html; charset=utf-8', headers  : { 'X-fancyBox': true } },
        autoWidth: false,
        autoHeight: true,
        autoResize: true,
        closeBtn	: false,
        modal		: true,
        'transitionIn'	 : 'elastic',
        'transitionOut'	 : 'elastic',
        topRatio	: 0.5, // по центру для регистрации
        leftRatio	: 0.5,
        title		: 'Знаком (*) помечены поля обязательные для ввода!',
        //'content' 		: data,
        /* 		'autoDimensions': true, */
        'overlayShow': true,
        helpers		: {
            overlay : { showEarly  : true },
            title	: { type : 'float'
            },
        }
    });

    return false;
}

// заменяю ссылки на AJAX - вызов
function AddClickShowOkno( parent_this ) {
    var str_location = document.domain,
        str = document.location.pathname,
        i = str.substring(1).search('/') + 2;


    if (i > 0)
        str_location += str.substring( 0, i ); // добавляем имя папочки, если она есть!

    // для показа в модальном окне
    $('a.modal').click( ShowModal );

    $('a[title*="откроется в новом окне"], a[target*="в новом окне"], a:has(img[alt*=modal]), a[title*="откроется в модальном окне"], a[alt*=referal], a:has( img[alt*=referal] ), a[title*="(переход)"], a[rel=nofollow], a.modal', parent_this ).addClass('referal');

    if( (target_blank = $('a[target="_blank"]:not( a.fancybox-button, a:has(img.fancybox-button), a.referal )', parent_this )).length > 0 )
    // показ в отдельном окне переделаем в вызов всплывающего окна
        MakePutFuncybox();


    if ( $('.fancybox-button', parent_this).length > 0 )
        VerifyIncludeFancyHelpers();


// обрабатываем все кнопки, кроме имеющих:
// 1. В alt (заменяющий текст) слово "referal", в title (всплывающая подсказка) "(переход)" или "откроется в новом окне"
// (это соглашение для сборки в Сайткарфте для пометки кнопок-не вкладышей
// 2. В href (ссылка) содержит "skype"
// 3. обработчик onclick
    $( 'a[href]:not( a[target="_blank"], a.fancybox-button, a:has(img.fancybox-button), a[href="#"], a[onclick], a[href*="skype:"], a.referal )', parent_this ) //referal и title*=\'(переход) означает ссылки, которые не надо менять [target!=_blank]  (откроется в новом окне)
        .each( function () {
            var str = this.href, this_img =  $(this).children();
            var i, j;

            i = str.search( "javascript" ); // оставлено для совместимости со старыими версиями СК
            if (i > -1)
            {
                this.data = this.href.substring(11);
                this.onclick = function() { eval( this.data ); return false; };
                /* 			          	  this.href= "#" + this.href.substring(11);   */
                return true;
            }

            j = str.search( '.htm' );
            i = str.search( document.domain);

            if (i == -1)
            {
                i = str.search( 'http://' );
                if  (i > -1) { // ссылка на внешний домен - пока оставляем как есть http://www.youtube.com/embed/P5stpcLie7U?feature=player_detailpage
                    i = str.search( 'youtube.com' );
                    if ( i > -1 )
                    { this.onclick = GetYoutube;
                        this.rel = str.replace( 'watch?v=', 'embed/' ).replace( "&feature=player_embedded", "?feature=player_detailpage");
                        this.alt = 'youtube' + str.substring( str.search('v=') + 2, str.search('/?feature') ); // позже добавить уникальное имя
                    }
                    else if (str.search( 'picasaweb' ) > -1) // Picasa albom
                        this.onclick = GetPicasaAlbom;
                    else // для остальных ссылок
                    {
                        this.onclick = function(){
                            var this_href = this.href, thisalt = this.alt;
                            DivContent.animate({ opacity: 0 },
                                function () {
                                    DivContent.load(this_href).animate({ opacity: 1 });
                                    PostShow(thisalt);
                                    ChangeButton(this_href);

                                    return false;
                                });
                            return false;
                        };
                        this.rel = 	this.href;
                    }

                    SetTitleAlt(this, this_img);
                    return true;
                }
                else {
                    i = str.search( 'https://' );

                    if  (i > -1) //обработки сссылок на https:
                    {
                        if (str.search( 'picasaweb' ) > -1) // Picasa albom
                            this.onclick = GetPicasaAlbom;
                        else if (str.search( 'docs.google.com' ) > -1) //  документы Гугла
                            this.onclick = GetGoogleDoc;
                        /*
                         else
                         this.onclick = GetHTTPSSource;
                         */

                        this.rel = str;

                        SetTitleAlt(this, this_img);
                        return true;
                        //i = i + 8;
                    } // завершение обработки сссылок на https://

                }
            } // нет в ссылке имени домена

            if ( j > 0 ) // для обработки на '.htm'- файлы
            {
                i = str.search( str_location );
                if (i == -1)
                    i = document.domain.length + (str.substring(0,7) == 'http://' ? 8 : 0); // учитываем http:// + 1 символ
                else
                    i += str_location.length;

                if (str.search( '#' ) > j) // якоря
                {  str = str.substring( i, j ) + '#' + str.substring( str.search( '#' )+1 );
                }
                else str = str.substring( i, j );
                //	alert(str);

                if ( str.substring( 0, 4 ) == 'menu' )
                {
                    if ( AddLoadMenu( str, this ) ) // на случай рекурсии
                        this.rel = str;
                    else
                        this.rel = "";

                }
                else if ( str.substring( 0, 7 ) == 'popmenu' )
                {
                    if ( AddPopMenu( str, $(this) ) ) // на случай рекурсии
                        this.rel = str;
                    else
                        this.rel = "";

                    this.onclick = function(){ return false;};
                }
                else
                {
                    this.onclick = ShowInPane;
                    this.rel = str;
                }


            } // завершение обработки на '.htm'- файлы
            else if ( str.search( '.php' ) > 0)
            {
                str = str.substring( i );
                j = str.search( '.php' );
                // отсекаю доменное имя
                i = str.search( str_location );
                if (i == -1)
                {

                    i = document.domain.length + (str.substring(0,7) == 'http://' ? 8 : 0); // учитываем http://, если он есть + 1 символ

                }
                else
                    i += str_location.length;

                this.rel = str.substring( i, j );
                this.data = str.substring( j );

                if (parent_this == document.body)
                    this.onclick = function(){ ShowOkno( this.rel, this.title, this, this.data ); return false; };
                else
                    this.onclick = function(){ ShowOkno( this.rel, this.title, this, this.data ); return false; };
            }// завершение обработки скриптов php
            else if (str.search( '#' ) > 0) // якоря
            {
                this.onclick = function(){ return GotoAncor( this.rel ); };

                this.rel     = str.substring( str.search( '#' )+1 );
                this.href 	  = str.substring( str.search( '#' ) );
            } else if ( str.match(/\/.*\/$/) || str.match(/\/.*\?.*/) ) { // route /menu/link/ ect.
                this.onclick = function(){ ShowOkno( this.href, this.title, this, this.data ); return false; };
            } else {

                str = str.substring( i );
                i = str.search( '/' );
                while (i > 0) {
                    str = str.substring( i+1 );
                    i = str.search( '/' );
                }

                if ( str.substring(0,1) == '#'){

                    this.onclick = function(){ LoadPrice( this.rel, this.title );  return false; } ;
                    this.rel = str.substring(1);

                    return ;
                }
            }

            if (!this.title)
            { this.title = ( this_img.is('img')  ? this_img.attr( 'title' ) || $(this).text() : $(this).text() ).replace( /\s\s\s\s/gm, '' );
            }

        });

    /*
     $( 'form[action]', parent_this ).not( 'form:has(a[href="#"])' ) //referal означает ссылки, которые не надо менять [target!=_blank]
     .each( function () {
     this.onsubmit = function() {var DivContent = DivContent; alert($(this).serialize());
     DivContent.load( this.rel, $(this).serializeArray() ); AddClickShowOkno( DivContent );
     return false; };
     this.rel = this.action;
     this.action = '#' + this.action;
     });
     */
    SetElementPosition( parent_this );
}

function GotoAncor( href ) {

    if ( ($('#okno').length > 0) && (DivContent.css('overflow-x') == 'hidden') )
        DivContent.scrollTo( $('a[name="' + href + '"]'), 800, {margin:true}  );
    else
        $('body').scrollTo( $('a[name="' + href + '"]'), 800, {margin:true}  );
    return false;
}

// Определить, показывать ли во фрейме или просто скачать контент
function ShowIframe(this_href) {

// Фрейм для GoogleDocs, страниц обратной связи, Яндекс-Метрики и всех не html-файлов, не имеющих параметра AJAX
    return ( (this_href.search('https://') > -1) || (this_href.search('poisk') > -1) || (this_href.search('metrika.yandex.ru') > -1)
        || (this_href.search('send') > -1) || (this_href.search('Inquiry') > -1)
        || ( (this_href.search('.htm') == -1) && (this_href.search(/ajax/i) == -1) )
    );
}

function GetTitle(str) {

    i = str.search( "<title>" ) + 7;
    j = str.search( "</title>" );

    return str.substring( i, j);
}


function AddPopMenu( str, aRef )
{
    if ( $("#pane_" + ExtractPageName(str) ).length > 0)
        return false;

    var html_menu = GetAjax( str, '.htm' ),
        DivForLoad =  $( '<div id="pane_' + ExtractPageName(str) + '" class="pop-pane body-pop-pane" ></div>' ).appendTo( aRef );

    DivForLoad.html( '<center>' + html_menu.replace( 'body-pane', "body-pop-pane" ) + '</center>');

    $(aRef).parent().addClass('pop_pane_parent');

    AddClickShowOkno( DivForLoad );

    aRef.mouseover( function onmouseover(event) {
        DivPopMenu = $( "#pane_" + ExtractPageName(this.rel) );
        if (menu_hide)
            clearTimeout( menu_hide );

        if (DivPopMenu.hasClass( 'pop_menu' ))
        { return true;}

        $( '.pop_menu' ).not( $(this).parents( ' .pop-pane' ) ).removeClass( 'pop_menu' ).hide();
        DivPopMenu.addClass( 'pop_menu' ).slideDown( "fast" ).show();

        return true;
    } );

    aRef.mouseleave(	function () {
        DivPopMenu = $( "#pane_" + ExtractPageName(this.rel) );
        if (menu_hide)
            clearTimeout( menu_hide);

        menu_hide = setTimeout( function(){
            DivPopMenu.removeClass( 'pop_menu' ).hide( )
        }, 1000 )
    });

    aRef.onclick = function(){ return false;};
    $(aRef).click( function(){ return false;} );

    DivForLoad.mouseover( function onmouseover(event) {
        if ( DivPopMenu && ( DivPopMenu.attr('id') == $(this).attr('id') ) )
        { if (menu_hide)
            clearTimeout( menu_hide);
            DivPopMenu = null;	}
        return true; } );

    DivForLoad.mouseleave( function mouseleave(event) {

        if (menu_hide)
            clearTimeout( menu_hide);
        DivPopMenu = $(this);
        menu_hide = setTimeout( function(){
            DivPopMenu.removeClass( 'pop_menu' ).hide( )
        }, 1000 );

        return true;
    } );
    return true;
}

function AddLoadMenu( name_page, aRef ) {
    var idDiv = 'pane_' + name_page,
        DivForLoad = $( '#' + idDiv );

    if ( DivForLoad.length > 0 )
        return false;


    DivForLoad =  $( '<div id="' + idDiv + '" class="slide-pane" ></div>' ).appendTo( $(aRef).parent() );

    $.post( name_page + '.htm', function (data, status)
    {
        DivForLoad.html( data.replace(/(<style[\s\S]*>\s*)(body)(?=\s+\{[^<]+<\/style>)/mg, '$1#' + idDiv) );
        AddClickShowOkno( DivForLoad );
        aRef.onclick = function(){ return LoadMenuToDiv( this.rel, '.htm', this ); };
    }).fail( function(data) {alert(data)});

    return true;
}
// переключение классов оживления кнопок
function ToggleOpacityClasses(this_ref) {
    if ( $(this_ref).hasClass( 'button_bright' ) || $(this_ref).hasClass( 'button_dark' ) )
        console.log(this_ref);
// 		$(this_ref).toggleClass( 'button_dark' ).toggleClass('button_bright');
    else
        return false;
    return true;
}
// поиск css hover состояния кнопки
function FindCascadinSheets( this_ref ) {
    var style = '';
    $(document.styleSheets).each( function() {
        try
        {
            $(this.cssRules).each( function(){
                if( ( match=this.cssText.match(/(.+)a:hover/) )
                    && ($(this_ref).closest(match[1]).length > 0)
                )
                {
                    style = GetStylesFromCSSRules( this );
                    return false;
                }
            });
            if ( style )
                return false;
        }
        catch(e) {
            console.log(e, this );
        }
    });
    return style;
}
//  складываю свойства в объект
function GetStylesFromCSSRules( css ) {
    return css.cssText;

    var  styles = new Object();
    for( var k=0; k < css.style.length; k++ )
        if ( css.style[k] )
            styles[ css.style[k] ] = css.style[ css.style[k] ]; // переписываем свойства в массив!
    return styles;
}
function FindCSSinSheets( selector ) {
    for (var i = 0; i < document.styleSheets.length; i++ )
    {
        try
        {
            if ( document.styleSheets[i].cssRules )
                for (var j = 0; j < document.styleSheets[i].cssRules.length; j++ )
                {
                    var css = document.styleSheets[i].cssRules[j];
                    if (css.selectorText == selector)
                    {
                        return GetStylesFromCSSRules( css );
                    }
                }
        }
        catch(e) {
            console.log( e, 'i='+i, 'j='+j, document.styleSheets[i] );
        }
    }
    return false;
}
// заморозка нажатой кнопки
function RemoveOver_Off( class_name, this_ref ) {

    if ( $(this_ref).children().is('img') ) //старе кнопки СК
    {
        img =   $(this_ref).children();
        img_id = img.attr( 'id' );


        if ( (img_id) && ( document.wb_pre) && ( document.wb_pre[img_id] ) ) // СК-кнопка
        {
            if (document.wb_normal)
            {
                img.attr( 'old_src', document.wb_normal );
                document.wb_normal = '';
            }
            else
            {
                img.attr( 'old_src', img.attr('src') );
                img.attr( 'src', document.wb_pre[img_id].src );
            }
            removeEvent( this_ref, 'mouseout', over_off ); // удаляем все обработчики ухода мышки с кнопки
// 	    	$(this_ref).trigger('focus');
        }
        else if ( $(this_ref).hasClass( 'button' ) )
        {
            $(this_ref).addClass( 'button_down' );
        }
        else if ( !ToggleOpacityClasses(this_ref) )

            $(this_ref).trigger('hover');
    }
    else  // будем искать в стилях описание Ховера
    {
        if ( classes = $(this_ref).attr('class') )
        {
            classes = classes.substring( classes.search('a') );

            if ( styles = FindCSSinSheets( '.' + classes + ':hover' ) )
            {
                styles = styles.replace(/(.+){/, '.down_button {' );
            }
        }
        else if ( styles = FindCascadinSheets( this_ref ) )
            styles = styles.replace(/:hover/, '.down_button ' ); //$(this_ref).css( styles );
        else if ( styles = FindCSSinSheets( 'a:hover' ) )
            styles = styles.replace(/:hover/, '.down_button ' ); //$(this_ref).css( styles );
        else
        {
            styles = '.down_button { color: ' + $(this_ref).css( 'color' ) + '; opacity:' + $(this_ref).css( 'opacity' ) + '}';
        }
        if ( $('head > style#down_button').length == 0 )
            $('head').append('<style type="text/css" id="down_button"> </style>');
        $('head > style#down_button').html( styles );
    }

    $(this_ref).addClass( class_name );

    return true;
}
// восстановление оживления кнопки, которую размораживаем
function RestoreOver_Off( class_name ) {
    var  Button_ref = $( '.' + class_name );

    $('.button_down').removeClass( 'button_down' ).removeClass( class_name );

    if (Button_ref.length > 0 ) {
        Button_ref.removeClass( class_name )
            .each( function () {

                if ( $(this).children().is('img') )
                {
                    Button_img = $(this).children('img');
                    if ( old_img = Button_img.attr( 'old_src' ) )
                    {
                        Button_img.attr( 'src', old_img ); // восстанавливаем изначальную картинку
                        buttonsAE( this, "mouseout", over_off ); // назначаем обратно оживление кнопки картинкой
                    }
                    else if ( !ToggleOpacityClasses(this) )
                        return Button_ref;
                }
                /* уже не нужно
                 else  // добавить потом удаление присвоенных стилей
                 $(this).css( { color: '', opacity : '' } );
                 */
            } );
    }
    return Button_ref;
}

/*   здесь отрабатываем замораживание картинки у нажатой кнопки, отмораживаем замороженные до этого */
function  ChangeButton( this_ref )  {
    RestoreOver_Off( 'down_button' );
    return  RemoveOver_Off(  'down_button', this_ref );
}

// показ выпадающего меню и т.п.
function LoadMenuToDiv( name_page, str_ext, this_ref ) {
    var DivForLoad = $( "#pane_" + name_page );
    var str;
    var Button = $(this_ref).children('img');
    var Button_name = Button.attr( 'id' );

    name_page = ( name_page.search('/') > 0 ? name_page.substring( name_page.search('/')+1 ): name_page );

    if (DivForLoad.attr( 'rel' ) == name_page) {
        document.wb_normal = Button.attr( 'old_src' );
        this_ref.onmouseout = function onmouseout(event) { over_off(); };

        DivForLoad.slideUp( "slow" ).attr( 'rel', '' ).removeClass( 'show_div' );

        return false;
    }

    Down_menu = $( 'div.show_div' );
    RestoreOver_Off( 'show_div' );

    if (Down_menu.attr( 'id' ) != 'pane' )
    {
        Down_menu.slideUp( "fast",
            function () {
                Down_menu.attr( 'rel', '' ).removeClass( 'show_div' );
            });
    }

    RemoveOver_Off( 'show_div', this_ref );

    DivForLoad.slideDown( "fast",
        function () {
            DivForLoad.attr( 'rel', name_page ).addClass( 'show_div' );
        } );
    return false;
}

function sema(p2, p1)
{ var res = ''; for (var i = 0; i < p2.length; i++) { res = res + p2.charAt(i); res = res + p1.charAt(i); } window.location.href = res;
}

// получить окно по тагу
function getIframeDocument(iframeNode) {
    if (iframeNode.contentDocument) return iframeNode.contentDocument
    if (iframeNode.contentWindow) return iframeNode.contentWindow.document
    return iframeNode.document
}

function getCoords(elem) { // кроме IE8-
    var box = elem.getBoundingClientRect();

    return {
        top: box.top + pageYOffset,
        left: box.left + pageXOffset
    };

}

// обработка загрузки графического файла
var isProcess = 0;
function handleFileImageChange(evt) {
    var files = evt.files || evt.target.files; // FileList object

    if (files.length < 1)
        return false;

    f = files[0];
    // Only process image files.
    if (!f.type.match('image.*')) {
        alert('Файл не графического типа!');
        return true;
    }

    if (f.size > 2000000) {
        alert('Файл ' + f.name + ' слишком велик! Загрузка может оказаться очень долгой и прерваться в неподходящий момент. Советуем сжать его с помощью любого графического редатора.' + f.size + ' б');
        return true;
    }

    var reader = new FileReader();

    // Closure to capture the file information.
    reader.onload = (function(theFile) {
        return function(e) {
            // Render thumbnail.
            $('#img_' + evt.target.id).attr( 'src', e.target.result ).attr( 'title', escape(theFile.name) );
            // показываем кнопку Сохранить,вызываем событие измененние данных
            $('#' + $('#' + evt.target.id).attr('form') ).trigger('oninput');
        };
    })(f);

    // Read in the image file as a data URL.
    reader.readAsDataURL(f);
    //}
    return true;
}


