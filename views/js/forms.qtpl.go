/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

// Code generated by qtc from "forms.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

// All the text outside function templates is treated as comments,
// i.e. it is just ignored by quicktemplate compiler (`qtc`). It is for humans.
//
// .

//line forms.qtpl:5
package js

//line forms.qtpl:5
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line forms.qtpl:5
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line forms.qtpl:5
func StreamHeadJSForForm(qw422016 *qt422016.Writer, afterAuthURL string) {
//line forms.qtpl:5
	qw422016.N().S(`
<script>
var user = ''
var userStruct
var urlAfterLogin = ''

function fancyOpen(data) {
      $.fancybox.open({
            'autoScale': true,
            'transitionIn': 'elastic',
            'transitionOut': 'elastic',
            'speedIn': 500,
            'speedOut': 300,
             'type':'html',
            'autoDimensions': true,
            'centerOnScroll': true,
            'content' : data
         })
}


function getUser() {
    user = localStorage.getItem("USER")
    if (user > '') {
        userStruct = JSON.parse(user)
        document.getElementById('bLogin').textContent = userStruct.name + "(" + userStruct.lang +")";
        token =  userStruct.token
        lang  =  userStruct.lang
        $('.auth').removeClass("auth");
`)
//line forms.qtpl:34
	if afterAuthURL > "" {
//line forms.qtpl:34
		qw422016.N().S(`           loadContent("`)
//line forms.qtpl:35
		qw422016.N().S(afterAuthURL)
//line forms.qtpl:35
		qw422016.N().S(`")
        `)
//line forms.qtpl:36
	}
//line forms.qtpl:36
	qw422016.N().S(`
    }

    return ''
}

var token = '';
var lang  = 'ua'
var isProcess = false;

function setClickAll() {
 if (isProcess) {
   return;
 }

  isProcess = true;

      console.log('ready click events!')
      // add onSubmit event instead default behaviourism form
      $('form:not([onsubmit])').on("onsubmit", function () {return saveForm(this); });
       // add click event instead default - response will show on div.#content
     $( 'a[href!="#"]:not([rel]):not(onclick):not([target=_blank])').each( function () {
        var url = this.href
        var target = this.target
        this.rel = 'setClickAll';
        isSearch = (this.target=="search");

        $(this).click( `)
//line forms.qtpl:63
	StreamOverClick(qw422016)
//line forms.qtpl:63
	qw422016.N().S(` )

      })
  isProcess = false;
}

function LoadStyles(id, styles) {
        let $head = $('head > style#' + id);
       if ( $head.length == 0 )
            $head = $('head').append('<style type="text/css" id="' + id + '"> </style>');
        $head.html( styles );
}

var go_history=1;
// Эта функция отрабатывает при перемещении по истории просмотром (кнопки вперед-назад в браузере)
function MyPopState(event) {
    if ( (go_history == 0) || (event.state == null) /* || (str_hash == DivContent.attr( 'rel') )|| (ya_search == 'process') */  )
        return true;
    go_history = 0;
    console.log(event);
    loadContent( event.state );
}
// смена адресной строки с предотвращением перезагрузки Содержимого
function SetDocumentHash( str_path ) {
let root_page ="/";
let default_page = "index.html";
    // обрезаю доменное имя и меняю скрипты для магазинов, затем готовлю полный путь для записи в Хистори браузера
    str_path = GetShortURL( str_path )
    var  origin   = document.location.origin + ( str_path[0] == '/' ? '' : "/" )
            + ( ( str_path != root_page ) && (str_path != default_page) ? str_path : '' );

// 	document.location.hash = new_hash;
    if ( (go_history)  ) {

        window.history.pushState( str_path, document.title, document.location.origin );
        console.log(str_path);
    }
    go_history = 1;
}
function GetShortURL(m_adress) {
    var origin   = document.location.origin + '/',

    i = m_adress.search( origin );
    if (i > -1)
        return m_adress.substring( i + origin.length );

    return m_adress;
}


$(function()   {
    if (!window.onpopstate)   // подключаем запись посещенных вкладышей в истории посещений
        window.onpopstate = MyPopState;

    if (user == '') {
      getUser();
    }
      console.log('ready function events!')

    setClickAll();
    $('#content').on('DOMSubtreeModified', setClickAll );

        $("#inpS:not([rel])").on("blur", function(){

          if (event.relatedTarget && event.relatedTarget.className == "suggestions-constraints") {
                    return;
            }
            console.log(event);
            $('select.suggestions-constraints').hide();
        }).on('keyup',function(e){
        var x = event.which || event.keyCode;
            if (x == 40) {
                $("#inpS").unbind("blur");
                $('select.suggestions-constraints').focus();
                $('select.suggestions-constraints option:first').selected();
                return;
            }

             if ($("#inpS").val().length < 2) {
                return true;
             }

             $.ajax({
                 url: "/api/v1/search/list",
                 data: {
                         "lang": lang,
                         "value": $("#inpS").val(),
                         "count": 10,
                         "html": true
                 },
               beforeSend: function (xhr) {
                   xhr.setRequestHeader('Authorization', 'Bearer ' + token);
               },
               success: function (data, status) {
                 $('select.suggestions-constraints').html(data).show().on('keyup', function(e) {
                                                     var x = event.which || event.keyCode;
                                                     if (x == 32) {
                                                             $("#inpS").val( $('select.suggestions-constraints option:selected').text() );
                                                             $('select.suggestions-constraints').hide();
                                                             $('button[type="search"]').click();
                                                             return false;
                                                      }
                                               });
                 $('select.suggestions-constraints option').on('mouseup', function(e) {
                    $("#inpS").val( $(this).text() );
                    $('select.suggestions-constraints').hide();
                   $('button[type="search"]').click();
                    return true;
                 });
               },
               error: function (xhr, status, error) {
                   alert( "Code : " + xhr.status + " error :"+ error);
                   console.log(error);
               }
              });
        }).attr('rel', true);

}) // $(document).ready

`)
//line forms.qtpl:182
	StreamSaveForm(qw422016)
//line forms.qtpl:182
	qw422016.N().S(`

// стандартная обработка формы типа AnyForm после успшного сохранения результата
function afterSaveAnyForm(data, status) {

    if (data.content_url !== undefined) {
        loadContent(data.content_url);
    } else if (data.formActions !== undefined) {
        loadContent( data.formActions[0].url );
    } else if (data.error !== undefined) {
        alert(data.error);
    } else {
        console.log(data);
    }

    if (data.message !== undefined) {
        alert(data.message);
    }
}
// собственно, нужен для того, чтобы после авторизации отобразит в заголовке нечто
 function afterLogin(data, thisForm) {
    if (!data) {
      alert("Need users data!")
      return false;
    }

     token = data.token;
     lang = data.lang;
     localStorage.setItem("USER",  JSON.stringify(data) );

   $('#bLogin').text(data.name + "(" + lang +")");
    $('.auth').removeClass("auth");

   if (urlAfterLogin == '') {
    if (data.formActions !== undefined) {
     urlAfterLogin = data.formActions[0].url;
    } else {
     urlAfterLogin = "/user/profile";
    }
   }

   loadContent(urlAfterLogin)
}
// run request & show content
function loadContent(url) {

     $.ajax({
            url: url,
            data: {
                    "lang": lang,
                    "html": true
            },
            beforeSend: function (xhr) {
                xhr.setRequestHeader('Authorization', 'Bearer ' + token);
            },
            success: function (data, status) {
                PutContent(data, url);
            },
            error: function (xhr, status, error) {
                 if (xhr.status == 401) {
                    urlAfterLogin = url;
                    $('#bLogin').trigger("click");
                   return;
                  }

                alert( "Code : " + xhr.status + " error :"+ error);
                console.log(error);
            }
       });
}

function PutContent(data, url) {
	$('#content').html(data);
	SetDocumentHash(url);
}

function showObject(data, thisForm) {
    if (!data) {
      alert('no results!')
      return false;
    }

    $('#content').html( '' );

    for (x in data) {
        $('#content').append('<div>');
        div = $('#content').children(':last').attr('id', data[x].id)
        div.append('<div>');
        titleDiv = div.children(':last')
        titleDiv.append('<h3> Group: <a href="/api/v1/search/?name=' + data[x].title + '" target="search">' + data[x].title + '</a></h3>');
        if (data[x].abbr > "") {
            titleDiv.append('<h4>' + data[x].abbr + '</h4>');
        }

        if (data[x].id_article > "") {
            titleDiv.append('<h4><a href="article/' + data[x].id_article + '" target="search">' + data[x].id_article + '</a></h4>');
        }

        for (y in data[x].company_names ) {
            titleDiv.append('<span>' + data[x].company_names[y] + '</span>');
        }
        
        for (y in data[x].list ) {
          div.append('<a href="' + data[x].list[y].document_url + '" rel="true" target="_blank">PDF</a>')
          brand = data[x].list[y].brand
          if (brand !== undefined) {
            div.append('<div>' + brand + '<a href="/api/v1/search/analog/' + brand + '" target="search"> Search analog</a></div>')
          }

           idPolymers = data[x].list[y].id_polymers
          if (data[x].list[y].has_additives !== false) {
            div.append('<div><a href="/api/v1/search/additives/?id=' + idPolymers + '" target="search"> Search additives</a></div>')
          }

          if (data[x].list[y].has_fillers !== false) {
            div.append('<div><a href="/api/v1/search/fillers/?id=' + idPolymers + '" target="search"> Search fillers</a></div>')
          }

          div.append('<div>' + data[x].list[y].company_name+ '</div>')
          div.append('<table>');
            div.append('<thr><thd> Values </thd><thd> Qty </thd> </thr>')

            for (z in data[x].list[y].characteristics) {
              div.append('<tr><td>'+z+'</td><td></td><td>'+data[x].list[y].characteristics[z]+'</td></tr>');
            }

           for (z in data[x].list[y].files) {
              div.append('<p><a href="'+data[x].list[y].files[z].url+'" rel="true" target="_blank">'+data[x].list[y].files[z].title+'</a></p>');
           }

           notes = data[x].list[y].description
           if (notes  > "") {
                div.append('<div>' + notes + '</div>')
           }
        }
    }
}
</script>
`)
//line forms.qtpl:320
}

//line forms.qtpl:320
func WriteHeadJSForForm(qq422016 qtio422016.Writer, afterAuthURL string) {
//line forms.qtpl:320
	qw422016 := qt422016.AcquireWriter(qq422016)
//line forms.qtpl:320
	StreamHeadJSForForm(qw422016, afterAuthURL)
//line forms.qtpl:320
	qt422016.ReleaseWriter(qw422016)
//line forms.qtpl:320
}

//line forms.qtpl:320
func HeadJSForForm(afterAuthURL string) string {
//line forms.qtpl:320
	qb422016 := qt422016.AcquireByteBuffer()
//line forms.qtpl:320
	WriteHeadJSForForm(qb422016, afterAuthURL)
//line forms.qtpl:320
	qs422016 := string(qb422016.B)
//line forms.qtpl:320
	qt422016.ReleaseByteBuffer(qb422016)
//line forms.qtpl:320
	return qs422016
//line forms.qtpl:320
}