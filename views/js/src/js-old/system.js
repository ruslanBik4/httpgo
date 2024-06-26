/*
 * Copyright (c) 2022. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */

/**
 * Created by rus on 12.10.16.
 */
"use strict";
var imgItem, divContent, default_page = '/main/', notSaved;
window.onload = function () {

    // возможно, это можно сделать прямо в заголовке, а не тут
    divContent = $('#content');
    // AddClickShowOkno( $("body") );

    // $.get('/user/login/', function (data) {
    //     if (data.substr(0,5) == '<form') {
    //         showFormModal(data);
    //     } else { // уже залогинен, обрабатываем данные
    //         afterLogin( JSON.parse(data) );
    //     }
    // });
    /*function AddClickShowOkno( parent_this ) {
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
     }*/
// после загрузки новой страницы
// тут можно отрабатывать события, например, на расстановку евентов для элементов и так далее
    function SitePostShow() {

        //красивый скроллинг, потом проработать что бы он применялся ко всем элементам со скроллом
        // $(".main-form-wrap").niceScroll();
        var dateInputs = $("input[type=date],input[type=datetime]");


        //стилизация селектов и инпутов, смена бг
        //$(".business-form-select").styler();
        changeBg();
        moveLabel();

        // устанавливаем необходимые датапикеры
        if (dateInputs.length > 0) {
            //TODO: сделать проверку или установить флаг на то, что модуль уже загружен и не загружать если так
            $("<head>").append('<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery-datetimepicker/2.5.4/build/jquery.datetimepicker.full.min.js"></script>')
            dateInputs.each(function () {
                var maxDate = $(this).attr('maxdate');
                var minDate = $(this).attr('mindate');
                if (maxDate) {
                    $(this).datetimepicker({
                        format: this.type == "datetime" ? 'Y-m-d H:i' : 'Y-m-d',
                        maxDate: maxDate
                    });
                } else if (minDate) {
                    $(this).datetimepicker({
                        format: this.type == "datetime" ? 'Y-m-d H:i' : 'Y-m-d',
                        minDate: minDate
                    });
                } else {
                    $(this).datetimepicker({format: 'Y-m-d'});
                }

            });
        }

        $('.get-json[data-href]').each(function () {
            this.data('href')
        })

        //TODO: сделать подключение остальных полей (без maxDate) и, других своств - minDate, например

        //используем как событие загрузки формы
        $("form[oninvalid]").trigger('invalid');

        // if($('#fapproximation').length > 0){
        //     //TODO: научить проверять этот метод,  на загруженнсть джс
        //     enableApproximationHandler();
        // }
    }

    function moveLabel() {
        $(".custom-input-label").click(function () {
            $(this).addClass('small-label');
        });
        $(".input-label").click(function () {
            $(this).next('.form-items-wrap').find(".custom-input-label").addClass('small-label');
        });
        $('.business-form-input').keyup(function () {
            if ($(this).val().length != 0) {
                $(this).next(".custom-input-label").addClass('small-label');
            } else {
                $(this).next(".custom-input-label").removeClass('small-label');
            }
        });
    }

    function changeBg() {
        var checkFrom = $('#fbusiness');
        if (checkFrom.length > 0) {
            $('.content-wrap').css('background-image', 'url(/images/bg2.png)')
        } else {
            $('.content-wrap').css('background-image', 'url(/images/bg1.png)')
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
    function afterLogin(data) {
        if (!data)
            return false;

        var greetings = ( (data === null) || ( data.login === undefined) ? '' : 'Добро пожаловать, '
            + (data.sex === undefined ? "" : data.sex + " ") + data.login + '!');

        $('#sTitle').html(greetings);
        // $('#fTools > output').text( 'Можете добавить устройство из меню и перенести его в нужное Вам место.');
        // конфликтует с webcomponent.js
        // loginToggle();

        // $('#dMyTools').load('/user/usertools/menu');
        // $('#dMyRooms').load('/user/rooms/menu');

        $.get('/user/profile/', function (data) {
            data = GetPageParts(data);
            divContent.html(data);
            AddClickShowOkno(divContent);
        });
    }

// события после кнопки Выйти
    function logOut(thisElem) {
        $('canvas').detach();
        loginToggle();
        $('#sTitle').html('Для начала работы Вам необходимо ');
        $.get(thisElem.href, function (data) {
            showFormModal(data);
        });

        return false;
    }

    function loginToggle() {
        $('.btn-login').toggle();
    }

    function getOauth(thisElem) {
        var dataElem = $(thisElem).data(),
            props = dataElem.props,
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

        reader.onload = (function (theFile) {
            return function (e) {
                // Render thumbnail.
                divContent.css({backgroundImage: 'url(' + e.target.result + ')'}).html('');
                // $('svg', divContent).append("<img class='new_photo' src='" + e.target.result +
                //     "' title='" + escape(theFile.name) + "' alt='" + escape(theFile.name) + '/>');
            };
        })(f);

        // Read in the image file as a data URL.
        reader.readAsDataURL(f);
    }

    function checkDB(elem) {
        if (elem.checked) {
            $('input[name=ddl]').click();

            tinymce.init({
                selector: '#editorDDL',
                plugins: 'anchor autolink charmap codesample emoticons image link lists media searchreplace table visualblocks wordcount checklist mediaembed casechange export formatpainter pageembed linkchecker a11ychecker tinymcespellchecker permanentpen powerpaste advtable advcode editimage tinycomments tableofcontents footnotes mergetags autocorrect',
                toolbar: 'addNewTableButton | undo redo | blocks fontfamily fontsize | bold italic underline strikethrough | link image media table mergetags | addcomment showcomments | spellcheckdialog a11ycheck | align lineheight | checklist numlist bullist indent outdent | emoticons charmap | removeformat',
                tinycomments_mode: 'embedded',
                tinycomments_author: 'Author name',
                mergetags_list: [
                    {value: 'First.Name', title: 'First Name'},
                    {value: 'Email', title: 'Email'},
                ],
                setup: (editor) => {
                    editor.ui.registry.addButton('addNewTableButton', {
                        text: 'Add new table',
                        onAction: () => editor.insertContent("CREATE TABLE NAME( id serial, name  character varying not null, columns, PRIMARY KEY (id)); COMMENT ON TABLE NAME IS 'Description'; COMMENT ON COLUMN NAME.name IS 'description';create unique index NAME_name_idx ON NAME (name);");
                    });
                }
            });
        }
    }

    function showFlat(thisElem) {
        var images = '',
            bkgPos = '0% 0%, 50% 0%, 0% 50%, 50% 50%',
            bkgSize = '25%',
            comma = '';

        beforeLoadContent();

        $('ul > li > a', thisElem.parentElement).each(function () {
            var dataElem = $(this).data();

            if (dataElem.image === undefined)
                images += comma + "url('img/room.svg')";
            else
                images += comma + "url('" + dataElem.image.substr(3) + "')";

            for (var i in dataElem.props) {
                AddItem(dataElem.props[i]);
            }
            comma = ',';
        });

        divContent.css({backgroundImage: images, backgroundPosition: bkgPos, backgroundSize: bkgSize}).html('');

        return false;
    }

// на слуяай отвала зароса по AJAx
    function failAjax(data, status) {
        console.Log(data);
        alert(status);
    }

    //
    // function getData(element) {
    //   var url = element.attr('data-href');
    //   if (url) {
    //     console.log(url);
    //     $.get(url).done(function (data) {
    //       var fields = data.fields;
    //       var form = data.form;
    //       console.log(data);
    //       $.each(fields, function (key, object) {
    //         renderFormElements(key, object, element, form);
    //       });
    //     });
    //
    //   }
    //
    // }
    //
    // function renderFormElements(name, object, parent, form) {
    //   parent.find('form').attr({
    //     action: form.action,
    //     id: form.id,
    //     name: form.name,
    //     onload: form.onload,
    //     onsubmit: form.onsubmit,
    //     oninput: form.oninput,
    //     onreset: form.onreset
    //   });
    //   var thisElem = $('input[name=' + name + ']');
    //   parent.find(thisElem).addClass(object.CSSClass).attr({
    //     type: object.type,
    //     placeholder: object.title,
    //     title: object.title,
    //     maxlenght: object.maxLenght,
    //   });
    //   if (object.required) {
    //     thisElem.attr('required', true);
    //   }
    //
    // }
}


/*
 * SystemJS v0.20.12 Dev
 */
!function(){"use strict";function e(e){return lt?Symbol():"@@"+e}function t(e,t){it||(t=t.replace(st?/file:\/\/\//g:/file:\/\//g,""));var r,n=(e.message||e)+"\n  "+t;r=dt&&e.fileName?new Error(n,e.fileName,e.lineNumber):new Error(n);var o=e.originalErr?e.originalErr.stack:e.stack;return at?r.stack=n+"\n  "+o:r.stack=o,r.originalErr=e.originalErr||e,r}function r(e,t){throw new RangeError('Unable to resolve "'+e+'" to '+t)}function n(e,t){e=e.trim();var n=t&&t.substr(0,t.indexOf(":")+1),o=e[0],i=e[1];if("/"===o&&"/"===i)return n||r(e,t),n+e;if("."===o&&("/"===i||"."===i&&("/"===e[2]||2===e.length)||1===e.length)||"/"===o){var a,s=!n||"/"!==t[n.length];if(s?(void 0===t&&r(e,t),a=t):"/"===t[n.length+1]?"file:"!==n?(a=t.substr(n.length+2),a=a.substr(a.indexOf("/")+1)):a=t.substr(8):a=t.substr(n.length+1),"/"===o){if(!s)return t.substr(0,t.length-a.length-1)+e;r(e,t)}for(var u=a.substr(0,a.lastIndexOf("/")+1)+e,l=[],c=void 0,f=0;f<u.length;f++)if(void 0===c)if("."!==u[f])c=f;else{if("."!==u[f+1]||"/"!==u[f+2]&&f!==u.length-2){if("/"!==u[f+1]&&f!==u.length-1){c=f;continue}f+=1}else l.pop(),f+=2;s&&0===l.length&&r(e,t),f===u.length&&l.push("")}else"/"===u[f]&&(l.push(u.substr(c,f-c+1)),c=void 0);return void 0!==c&&l.push(u.substr(c,u.length-c)),t.substr(0,t.length-a.length)+l.join("")}var d=e.indexOf(":");return-1!==d?at&&":"===e[1]&&"\\"===e[2]&&e[0].match(/[a-z]/i)?"file:///"+e.replace(/\\/g,"/"):e:void 0}function o(e){if(e.values)return e.values();if("undefined"==typeof Symbol||!Symbol.iterator)throw new Error("Symbol.iterator not supported in this browser");var t={};return t[Symbol.iterator]=function(){var t=Object.keys(e),r=0;return{next:function(){return r<t.length?{value:e[t[r++]],done:!1}:{value:void 0,done:!0}}}},t}function i(){this.registry=new u}function a(e){if(!(e instanceof l))throw new TypeError("Module instantiation did not return a valid namespace object.");return e}function s(e){if(void 0===e)throw new RangeError("No resolution found.");return e}function u(){this[vt]={},this._registry=vt}function l(e){Object.defineProperty(this,yt,{value:e}),Object.keys(e).forEach(c,this)}function c(e){Object.defineProperty(this,e,{enumerable:!0,get:function(){return this[yt][e]}})}function f(){i.call(this);var e=this.registry.delete;this.registry.delete=function(r){var n=e.call(this,r);return t.hasOwnProperty(r)&&!t[r].linkRecord&&delete t[r],n};var t={};this[bt]={lastRegister:void 0,records:t},this.trace=!1}function d(e,t,r){return e.records[t]={key:t,registration:r,module:void 0,importerSetters:void 0,linkRecord:{instantiatePromise:void 0,dependencies:void 0,execute:void 0,executingRequire:!1,moduleObj:void 0,setters:void 0,depsInstantiatePromise:void 0,dependencyInstantiations:void 0,linked:!1,error:void 0}}}function p(e,t,r,n,o){var i=n[t];if(i)return Promise.resolve(i);var a=o.records[t];return a&&!a.module?h(e,a,a.linkRecord,n,o):e.resolve(t,r).then(function(t){if(i=n[t])return i;a=o.records[t],(!a||a.module)&&(a=d(o,t,a&&a.registration));var r=a.linkRecord;return r?h(e,a,r,n,o):a})}function g(e,t,r){return function(){var e=r.lastRegister;return e?(r.lastRegister=void 0,t.registration=e,!0):!!t.registration}}function h(e,r,n,o,i){return n.instantiatePromise||(n.instantiatePromise=(r.registration?Promise.resolve():Promise.resolve().then(function(){return i.lastRegister=void 0,e[wt](r.key,e[wt].length>1&&g(e,r,i))})).then(function(t){if(void 0!==t){if(!(t instanceof l))throw new TypeError("Instantiate did not return a valid Module object.");return delete i.records[r.key],e.trace&&v(e,r,n),o[r.key]=t}var a=r.registration;if(r.registration=void 0,!a)throw new TypeError("Module instantiation did not call an anonymous or correctly named System.register.");return n.dependencies=a[0],r.importerSetters=[],n.moduleObj={},a[2]?(n.moduleObj.default={},n.moduleObj.__useDefault=!0,n.executingRequire=a[1],n.execute=a[2]):b(e,r,n,a[1]),n.dependencies.length||(n.linked=!0,e.trace&&v(e,r,n)),r}).catch(function(e){throw n.error=t(e,"Instantiating "+r.key)}))}function m(e,t,r,n,o,i){return e.resolve(t,r).then(function(r){i&&(i[t]=r);var a=o.records[r],s=n[r];if(s&&(!a||a.module&&s!==a.module))return s;(!a||!s&&a.module)&&(a=d(o,r,a&&a.registration));var u=a.linkRecord;return u?h(e,a,u,n,o):a})}function v(e,t,r){e.loads=e.loads||{},e.loads[t.key]={key:t.key,deps:r.dependencies,dynamicDeps:[],depMap:r.depMap||{}}}function y(e,t,r){e.loads[t].dynamicDeps.push(r)}function b(e,t,r,n){var o=r.moduleObj,i=t.importerSetters,a=!1,s=n.call(ut,function(e,t){if("object"==typeof e){var r=!1;for(var n in e)t=e[n],"__useDefault"===n||n in o&&o[n]===t||(r=!0,o[n]=t);if(r===!1)return t}else{if((a||e in o)&&o[e]===t)return t;o[e]=t}for(var s=0;s<i.length;s++)i[s](o);return t},new x(e,t.key));r.setters=s.setters,r.execute=s.execute,s.exports&&(r.moduleObj=o=s.exports,a=!0)}function w(e,r,n,o,i,a){return(n.depsInstantiatePromise||(n.depsInstantiatePromise=Promise.resolve().then(function(){for(var t=Array(n.dependencies.length),a=0;a<n.dependencies.length;a++)t[a]=m(e,n.dependencies[a],r.key,o,i,e.trace&&n.depMap||(n.depMap={}));return Promise.all(t)}).then(function(e){if(n.dependencyInstantiations=e,n.setters)for(var t=0;t<e.length;t++){var r=n.setters[t];if(r){var o=e[t];o instanceof l?r(o):(r(o.module||o.linkRecord.moduleObj),o.importerSetters&&o.importerSetters.push(r))}}}))).then(function(){for(var t=[],r=0;r<n.dependencies.length;r++){var s=n.dependencyInstantiations[r],u=s.linkRecord;u&&!u.linked&&(-1===a.indexOf(s)?(a.push(s),t.push(w(e,s,s.linkRecord,o,i,a))):t.push(u.depsInstantiatePromise))}return Promise.all(t)}).then(function(){return n.linked=!0,e.trace&&v(e,r,n),r}).catch(function(e){throw e=t(e,"Loading "+r.key),n.error=n.error||e,e})}function k(e,t){var r=e[bt];r.records[t.key]===t&&delete r.records[t.key];var n=t.linkRecord;n&&n.dependencyInstantiations&&n.dependencyInstantiations.forEach(function(t,o){if(t&&!(t instanceof l)&&t.linkRecord&&(t.linkRecord.error&&r.records[t.key]===t&&k(e,t),n.setters&&t.importerSetters)){var i=t.importerSetters.indexOf(n.setters[o]);t.importerSetters.splice(i,1)}})}function x(e,t){this.loader=e,this.key=this.id=t}function O(e,t,r,n,o,i){if(t.module)return t.module;if(r.error)throw r.error;if(i&&-1!==i.indexOf(t))return t.linkRecord.moduleObj;var a=E(e,t,r,n,o,r.setters?[]:i||[]);if(a)throw k(e,t),a;return t.module}function S(e,t,r,n,o,i,a){return function(s){for(var u=0;u<r.length;u++)if(r[u]===s){var c,f=n[u];return c=f instanceof l?f:O(e,f,f.linkRecord,o,i,a),c.__useDefault?c.default:c}throw new Error("Module "+s+" not declared as a System.registerDynamic dependency of "+t)}}function E(e,r,n,o,i,a){a.push(r);var s;if(n.setters)for(var u,c,f=0;f<n.dependencies.length;f++)if(u=n.dependencyInstantiations[f],!(u instanceof l)&&(c=u.linkRecord,c&&-1===a.indexOf(u)&&(s=c.error?c.error:E(e,u,c,o,i,c.setters?a:[])),s))return n.error=t(s,"Evaluating "+r.key);if(n.execute)if(n.setters)s=j(n.execute);else{var d={id:r.key},p=n.moduleObj;Object.defineProperty(d,"exports",{configurable:!0,set:function(e){p.default=e},get:function(){return p.default}});var g=S(e,r.key,n.dependencies,n.dependencyInstantiations,o,i,a);if(!n.executingRequire)for(var f=0;f<n.dependencies.length;f++)g(n.dependencies[f]);s=P(n.execute,g,p.default,d),d.exports!==p.default&&(p.default=d.exports);var h=p.default;if(h&&h.__esModule)for(var m in p.default)Object.hasOwnProperty.call(p.default,m)&&"default"!==m&&(p[m]=h[m])}if(s)return n.error=t(s,"Evaluating "+r.key);if(o[r.key]=r.module=new l(n.moduleObj),!n.setters){if(r.importerSetters)for(var f=0;f<r.importerSetters.length;f++)r.importerSetters[f](r.module);r.importerSetters=void 0}r.linkRecord=void 0}function j(e){try{e.call(kt)}catch(e){return e}}function P(e,t,r,n){try{var o=e.call(ut,t,r,n);void 0!==o&&(n.exports=o)}catch(e){return e}}function R(){}function M(e){return e instanceof l?e:new l(e&&e.__esModule?e:{default:e,__useDefault:!0})}function _(e){return void 0===xt&&(xt="undefined"!=typeof Symbol&&!!Symbol.toStringTag),e instanceof l||xt&&"[object Module]"==Object.prototype.toString.call(e)}function C(e,t){(t||this.warnings&&"undefined"!=typeof console&&console.warn)&&console.warn(e)}function L(e,t,r){var n=new Uint8Array(t);return 0===n[0]&&97===n[1]&&115===n[2]?WebAssembly.compile(t).then(function(t){var n=[],o=[],i={};return WebAssembly.Module.imports&&WebAssembly.Module.imports(t).forEach(function(e){var t=e.module;o.push(function(e){i[t]=e}),-1===n.indexOf(t)&&n.push(t)}),e.register(n,function(e){return{setters:o,execute:function(){e(new WebAssembly.Instance(t,i).exports)}}}),r(),!0}):Promise.resolve(!1)}function A(e,t){if("."===e[0])throw new Error("Node module "+e+" can't be loaded as it is not a package require.");if(!Ot){var r=this._nodeRequire("module"),n=t.substr(st?8:7);Ot=new r(n),Ot.paths=r._nodeModulePaths(n)}return Ot.require(e)}function I(e,t){for(var r in t)Object.hasOwnProperty.call(t,r)&&(e[r]=t[r]);return e}function F(e,t){for(var r in t)Object.hasOwnProperty.call(t,r)&&void 0===e[r]&&(e[r]=t[r]);return e}function K(e,t,r){for(var n in t)if(Object.hasOwnProperty.call(t,n)){var o=t[n];void 0===e[n]?e[n]=o:o instanceof Array&&e[n]instanceof Array?e[n]=[].concat(r?o:e[n]).concat(r?e[n]:o):"object"==typeof o&&null!==o&&"object"==typeof e[n]?e[n]=(r?F:I)(I({},e[n]),o):r||(e[n]=o)}}function D(e){if(!Mt&&!_t){var t=new Image;return void(t.src=e)}var r=document.createElement("link");Mt?(r.rel="preload",r.as="script"):r.rel="prefetch",r.href=e,document.head.appendChild(r),document.head.removeChild(r)}function q(e,t,r){try{importScripts(e)}catch(e){r(e)}t()}function T(e,t,r,n,o){function i(){n(),s()}function a(t){s(),o(new Error("Fetching "+e))}function s(){for(var e=0;e<Ct.length;e++)if(Ct[e].err===a){Ct.splice(e,1);break}u.removeEventListener("load",i,!1),u.removeEventListener("error",a,!1),document.head.removeChild(u)}if(e=e.replace(/#/g,"%23"),Rt)return q(e,n,o);var u=document.createElement("script");u.type="text/javascript",u.charset="utf-8",u.async=!0,t&&(u.crossOrigin=t),r&&(u.integrity=r),u.addEventListener("load",i,!1),u.addEventListener("error",a,!1),u.src=e,document.head.appendChild(u)}function U(e,t){for(var r=e.split(".");r.length;)t=t[r.shift()];return t}function z(e,t,r){var o=J(t,r);if(o){var i=t[o]+r.substr(o.length),a=n(i,ot);return void 0!==a?a:e+i}return-1!==r.indexOf(":")?r:e+r}function N(e){var t=this.name;if(t.substr(0,e.length)===e&&(t.length===e.length||"/"===t[e.length]||"/"===e[e.length-1]||":"===e[e.length-1])){var r=e.split("/").length;r>this.len&&(this.match=e,this.len=r)}}function J(e,t){if(Object.hasOwnProperty.call(e,t))return t;var r={name:t,match:void 0,len:0};return Object.keys(e).forEach(N,r),r.match}function $(e,t,r,n){if("file:///"===e.substr(0,8)){if(Kt)return B(e,t,r,n);throw new Error("Unable to fetch file URLs in this environment.")}e=e.replace(/#/g,"%23");var o={headers:{Accept:"application/x-es-module, */*"}};return r&&(o.integrity=r),t&&("string"==typeof t&&(o.headers.Authorization=t),o.credentials="include"),fetch(e,o).then(function(e){if(e.ok)return n?e.arrayBuffer():e.text();throw new Error("Fetch error: "+e.status+" "+e.statusText)})}function B(e,t,r,n){return new Promise(function(r,o){function i(){r(n?s.response:s.responseText)}function a(){o(new Error("XHR error: "+(s.status?" ("+s.status+(s.statusText?" "+s.statusText:"")+")":"")+" loading "+e))}e=e.replace(/#/g,"%23");var s=new XMLHttpRequest;n&&(s.responseType="arraybuffer"),s.onreadystatechange=function(){4===s.readyState&&(0==s.status?s.response?i():(s.addEventListener("error",a),s.addEventListener("load",i)):200===s.status?i():a())},s.open("GET",e,!0),s.setRequestHeader&&(s.setRequestHeader("Accept","application/x-es-module, */*"),t&&("string"==typeof t&&s.setRequestHeader("Authorization",t),s.withCredentials=!0)),s.send(null)})}function W(e,t,r,n){return"file:///"!=e.substr(0,8)?Promise.reject(new Error('Unable to fetch "'+e+'". Only file URLs of the form file:/// supported running in Node.')):(At=At||require("fs"),e=st?e.replace(/\//g,"\\").substr(8):e.substr(7),new Promise(function(t,r){At.readFile(e,function(e,o){if(e)return r(e);if(n)t(o);else{var i=o+"";"\ufeff"===i[0]&&(i=i.substr(1)),t(i)}})}))}function G(){throw new Error("No fetch method is defined for this environment.")}function H(){return{pluginKey:void 0,pluginArgument:void 0,pluginModule:void 0,packageKey:void 0,packageConfig:void 0,load:void 0}}function Z(e,t,r){var n=H();if(r){var o;t.pluginFirst?-1!==(o=r.lastIndexOf("!"))&&(n.pluginArgument=n.pluginKey=r.substr(0,o)):-1!==(o=r.indexOf("!"))&&(n.pluginArgument=n.pluginKey=r.substr(o+1)),n.packageKey=J(t.packages,r),n.packageKey&&(n.packageConfig=t.packages[n.packageKey])}return n}function X(e,t){var r=this[jt],n=H(),o=Z(this,r,t),i=this;return Promise.resolve().then(function(){var r=e.lastIndexOf("#?");if(-1===r)return Promise.resolve(e);var n=me.call(i,e.substr(r+2));return ve.call(i,n,t,!0).then(function(t){return t?e.substr(0,r):"@empty"})}).then(function(e){var a=oe(r.pluginFirst,e);return a?(n.pluginKey=a.plugin,Promise.all([te.call(i,r,a.argument,o&&o.pluginArgument||t,n,o,!0),i.resolve(a.plugin,t)]).then(function(e){if(n.pluginArgument=e[0],n.pluginKey=e[1],n.pluginArgument===n.pluginKey)throw new Error("Plugin "+n.pluginArgument+" cannot load itself, make sure it is excluded from any wildcard meta configuration via a custom loader: false rule.");return ie(r.pluginFirst,e[0],e[1])})):te.call(i,r,e,o&&o.pluginArgument||t,n,o,!1)}).then(function(e){return ye.call(i,e,t,o)}).then(function(e){return ne.call(i,r,e,n),n.pluginKey||!n.load.loader?e:i.resolve(n.load.loader,e).then(function(t){return n.pluginKey=t,n.pluginArgument=e,e})}).then(function(e){return i[Pt][e]=n,e})}function Y(e,t){var r=oe(e.pluginFirst,t);if(r){var n=Y.call(this,e,r.plugin);return ie(e.pluginFirst,V.call(this,e,r.argument,void 0,!1,!1),n)}return V.call(this,e,t,void 0,!1,!1)}function Q(e,t){var r=this[jt],n=H(),o=o||Z(this,r,t),i=oe(r.pluginFirst,e);return i?(n.pluginKey=Q.call(this,i.plugin,t),ie(r.pluginFirst,ee.call(this,r,i.argument,o.pluginArgument||t,n,o,!!n.pluginKey),n.pluginKey)):ee.call(this,r,e,o.pluginArgument||t,n,o,!!n.pluginKey)}function V(e,t,r,o,i){var a=n(t,r||ot);if(a)return z(e.baseURL,e.paths,a);if(o){var s=J(e.map,t);if(s&&(t=e.map[s]+t.substr(s.length),a=n(t,ot)))return z(e.baseURL,e.paths,a)}if(this.registry.has(t))return t;if("@node/"===t.substr(0,6))return t;var u=i&&"/"!==t[t.length-1],l=z(e.baseURL,e.paths,u?t+"/":t);return u?l.substr(0,l.length-1):l}function ee(e,t,r,n,o,i){if(o&&o.packageConfig&&"."!==t[0]){var a=o.packageConfig.map,s=a&&J(a,t);if(s&&"string"==typeof a[s]){var u=le(this,e,o.packageConfig,o.packageKey,s,t,n,i);if(u)return u}}var l=V.call(this,e,t,r,!0,!0),c=pe(e,l);if(n.packageKey=c&&c.packageKey||J(e.packages,l),!n.packageKey)return l;if(-1!==e.packageConfigKeys.indexOf(l))return n.packageKey=void 0,l;n.packageConfig=e.packages[n.packageKey]||(e.packages[n.packageKey]=Se());var f=l.substr(n.packageKey.length+1);return se(this,e,n.packageConfig,n.packageKey,f,n,i)}function te(e,t,r,n,o,i){var a=this;return St.then(function(){if(o&&o.packageConfig&&"./"!==t.substr(0,2)){var r=o.packageConfig.map,s=r&&J(r,t);if(s)return fe(a,e,o.packageConfig,o.packageKey,s,t,n,i)}return St}).then(function(o){if(o)return o;var s=V.call(a,e,t,r,!0,!0),u=pe(e,s);if(n.packageKey=u&&u.packageKey||J(e.packages,s),!n.packageKey)return Promise.resolve(s);if(-1!==e.packageConfigKeys.indexOf(s))return n.packageKey=void 0,n.load=re(),n.load.format="json",n.load.loader="",Promise.resolve(s);n.packageConfig=e.packages[n.packageKey]||(e.packages[n.packageKey]=Se());var l=u&&!n.packageConfig.configured;return(l?ge(a,e,u.configPath,n):St).then(function(){var t=s.substr(n.packageKey.length+1);return ce(a,e,n.packageConfig,n.packageKey,t,n,i)})})}function re(){return{extension:"",deps:void 0,format:void 0,loader:void 0,scriptLoad:void 0,globals:void 0,nonce:void 0,integrity:void 0,sourceMap:void 0,exports:void 0,encapsulateGlobal:!1,crossOrigin:void 0,cjsRequireDetection:!0,cjsDeferDepsExecute:!1,esModule:!1}}function ne(e,t,r){r.load=r.load||re();var n,o=0;for(var i in e.meta)if(n=i.indexOf("*"),-1!==n&&i.substr(0,n)===t.substr(0,n)&&i.substr(n+1)===t.substr(t.length-i.length+n+1)){var a=i.split("/").length;a>o&&(o=a),K(r.load,e.meta[i],o!==a)}if(e.meta[t]&&K(r.load,e.meta[t],!1),r.packageKey){var s=t.substr(r.packageKey.length+1),u={};if(r.packageConfig.meta){var o=0;he(r.packageConfig.meta,s,function(e,t,r){r>o&&(o=r),K(u,t,r&&o>r)}),K(r.load,u,!1)}!r.packageConfig.format||r.pluginKey||r.load.loader||(r.load.format=r.load.format||r.packageConfig.format)}}function oe(e,t){var r,n,o=e?t.indexOf("!"):t.lastIndexOf("!");return-1!==o?(e?(r=t.substr(o+1),n=t.substr(0,o)):(r=t.substr(0,o),n=t.substr(o+1)||r.substr(r.lastIndexOf(".")+1)),{argument:r,plugin:n}):void 0}function ie(e,t,r){return e?r+"!"+t:t+"!"+r}function ae(e,t,r,n,o){if(!n||!t.defaultExtension||"/"===n[n.length-1]||o)return n;var i=!1;if(t.meta&&he(t.meta,n,function(e,t,r){return 0===r||e.lastIndexOf("*")!==e.length-1?i=!0:void 0}),!i&&e.meta&&he(e.meta,r+"/"+n,function(e,t,r){return 0===r||e.lastIndexOf("*")!==e.length-1?i=!0:void 0}),i)return n;var a="."+t.defaultExtension;return n.substr(n.length-a.length)!==a?n+a:n}function se(e,t,r,n,o,i,a){if(!o){if(!r.main)return n;o="./"===r.main.substr(0,2)?r.main.substr(2):r.main}if(r.map){var s="./"+o,u=J(r.map,s);if(u||(s="./"+ae(e,r,n,o,a),s!=="./"+o&&(u=J(r.map,s))),u){var l=le(e,t,r,n,u,s,i,a);if(l)return l}}return n+"/"+ae(e,r,n,o,a)}function ue(e,t,r){return t.substr(0,e.length)===e&&r.length>e.length?!1:!0}function le(e,t,r,n,o,i,a,s){"/"===i[i.length-1]&&(i=i.substr(0,i.length-1));var u=r.map[o];if("object"==typeof u)throw new Error("Synchronous conditional normalization not supported sync normalizing "+o+" in "+n);if(ue(o,u,i)&&"string"==typeof u)return ee.call(this,t,u+i.substr(o.length),n+"/",a,a,s)}function ce(e,t,r,n,o,i,a){if(!o){if(!r.main)return Promise.resolve(n);o="./"===r.main.substr(0,2)?r.main.substr(2):r.main}var s,u;return r.map&&(s="./"+o,u=J(r.map,s),u||(s="./"+ae(e,r,n,o,a),s!=="./"+o&&(u=J(r.map,s)))),(u?fe(e,t,r,n,u,s,i,a):St).then(function(t){return t?Promise.resolve(t):Promise.resolve(n+"/"+ae(e,r,n,o,a))})}function fe(e,t,r,n,o,i,a,s){"/"===i[i.length-1]&&(i=i.substr(0,i.length-1));var u=r.map[o];if("string"==typeof u)return ue(o,u,i)?te.call(e,t,u+i.substr(o.length),n+"/",a,a,s).then(function(t){return ye.call(e,t,n+"/",a)}):St;var l=[],c=[];for(var d in u){var p=me(d);c.push({condition:p,map:u[d]}),l.push(f.prototype.import.call(e,p.module,n))}return Promise.all(l).then(function(e){for(var t=0;t<c.length;t++){var r=c[t].condition,n=U(r.prop,e[t].__useDefault?e[t].default:e[t]);if(!r.negate&&n||r.negate&&!n)return c[t].map}}).then(function(r){return r?ue(o,r,i)?te.call(e,t,r+i.substr(o.length),n+"/",a,a,s).then(function(t){return ye.call(e,t,n+"/",a)}):St:void 0})}function de(e){var t=e.lastIndexOf("*"),r=Math.max(t+1,e.lastIndexOf("/"));return{length:r,regEx:new RegExp("^("+e.substr(0,r).replace(/[.+?^${}()|[\]\\]/g,"\\$&").replace(/\*/g,"[^\\/]+")+")(\\/|$)"),wildcard:-1!==t}}function pe(e,t){for(var r,n,o=!1,i=0;i<e.packageConfigPaths.length;i++){var a=e.packageConfigPaths[i],s=qt[a]||(qt[a]=de(a));if(!(t.length<s.length)){var u=t.match(s.regEx);!u||r&&(o&&s.wildcard||!(r.length<u[1].length))||(r=u[1],o=!s.wildcard,n=r+a.substr(s.length))}}return r?{packageKey:r,configPath:n}:void 0}function ge(e,r,n,o,i){var a=e.pluginLoader||e;return-1===r.packageConfigKeys.indexOf(n)&&r.packageConfigKeys.push(n),a.import(n).then(function(e){Ee(o.packageConfig,e,o.packageKey,!0,r),o.packageConfig.configured=!0}).catch(function(e){throw t(e,"Unable to fetch package configuration file "+n)})}function he(e,t,r){var n;for(var o in e){var i="./"===o.substr(0,2)?"./":"";if(i&&(o=o.substr(2)),n=o.indexOf("*"),-1!==n&&o.substr(0,n)===t.substr(0,n)&&o.substr(n+1)===t.substr(t.length-o.length+n+1)&&r(o,e[i+o],o.split("/").length))return}var a=e[t]&&Object.hasOwnProperty.call(e,t)?e[t]:e["./"+t];a&&r(a,a,0)}function me(e){var t,r,n,n,o=e.lastIndexOf("|");return-1!==o?(t=e.substr(o+1),r=e.substr(0,o),"~"===t[0]&&(n=!0,t=t.substr(1))):(n="~"===e[0],t="default",r=e.substr(n),-1!==Tt.indexOf(r)&&(t=r,r=null)),{module:r||"@system-env",prop:t,negate:n}}function ve(e,t,r){return f.prototype.import.call(this,e.module,t).then(function(t){var n=U(e.prop,t);if(r&&"boolean"!=typeof n)throw new TypeError("Condition did not resolve to a boolean.");return e.negate?!n:n})}function ye(e,t,r){var n=e.match(Ut);if(!n)return Promise.resolve(e);var o=me.call(this,n[0].substr(2,n[0].length-3));return ve.call(this,o,t,!1).then(function(r){if("string"!=typeof r)throw new TypeError("The condition value for "+e+" doesn't resolve to a string.");if(-1!==r.indexOf("/"))throw new TypeError("Unabled to interpolate conditional "+e+(t?" in "+t:"")+"\n	The condition value "+r+' cannot contain a "/" separator.');return e.replace(Ut,r)})}function be(e,t,r){for(var n=0;n<zt.length;n++){var o=zt[n];t[o]&&Sr[o.substr(0,o.length-6)]&&r(t[o])}}function we(e,t){var r={};for(var n in e){var o=e[n];t>1?o instanceof Array?r[n]=[].concat(o):"object"==typeof o?r[n]=we(o,t-1):"packageConfig"!==n&&(r[n]=o):r[n]=o}return r}function ke(e,t){var r=e[t];return r instanceof Array?e[t].concat([]):"object"==typeof r?we(r,3):e[t]}function xe(e){if(e){if(-1!==Er.indexOf(e))return ke(this[jt],e);throw new Error('"'+e+'" is not a valid configuration name. Must be one of '+Er.join(", ")+".")}for(var t={},r=0;r<Er.length;r++){var n=Er[r],o=ke(this[jt],n);void 0!==o&&(t[n]=o)}return t}function Oe(e,t){var r=this,o=this[jt];if("warnings"in e&&(o.warnings=e.warnings),"wasm"in e&&(o.wasm="undefined"!=typeof WebAssembly&&e.wasm),("production"in e||"build"in e)&&rt.call(r,!!e.production,!!(e.build||Sr&&Sr.build)),!t){var i;be(r,e,function(e){i=i||e.baseURL}),i=i||e.baseURL,i&&(o.baseURL=n(i,ot)||n("./"+i,ot),"/"!==o.baseURL[o.baseURL.length-1]&&(o.baseURL+="/")),e.paths&&I(o.paths,e.paths),be(r,e,function(e){e.paths&&I(o.paths,e.paths)});for(var a in o.paths)-1!==o.paths[a].indexOf("*")&&(C.call(o,"Path config "+a+" -> "+o.paths[a]+" is no longer supported as wildcards are deprecated."),delete o.paths[a])}if(e.defaultJSExtensions&&C.call(o,"The defaultJSExtensions configuration option is deprecated.\n  Use packages defaultExtension instead.",!0),"boolean"==typeof e.pluginFirst&&(o.pluginFirst=e.pluginFirst),e.map)for(var a in e.map){var s=e.map[a];if("string"==typeof s){var u=V.call(r,o,s,void 0,!1,!1);"/"===u[u.length-1]&&":"!==a[a.length-1]&&"/"!==a[a.length-1]&&(u=u.substr(0,u.length-1)),o.map[a]=u}else{var l=V.call(r,o,"/"!==a[a.length-1]?a+"/":a,void 0,!0,!0);l=l.substr(0,l.length-1);var c=o.packages[l];c||(c=o.packages[l]=Se(),c.defaultExtension=""),Ee(c,{map:s},l,!1,o)}}if(e.packageConfigPaths){for(var f=[],d=0;d<e.packageConfigPaths.length;d++){var p=e.packageConfigPaths[d],g=Math.max(p.lastIndexOf("*")+1,p.lastIndexOf("/")),h=V.call(r,o,p.substr(0,g),void 0,!1,!1);f[d]=h+p.substr(g)}o.packageConfigPaths=f}if(e.bundles)for(var a in e.bundles){for(var m=[],d=0;d<e.bundles[a].length;d++)m.push(r.normalizeSync(e.bundles[a][d]));o.bundles[a]=m}if(e.packages)for(var a in e.packages){if(a.match(/^([^\/]+:)?\/\/$/))throw new TypeError('"'+a+'" is not a valid package name.');var l=V.call(r,o,"/"!==a[a.length-1]?a+"/":a,void 0,!0,!0);l=l.substr(0,l.length-1),Ee(o.packages[l]=o.packages[l]||Se(),e.packages[a],l,!1,o)}if(e.depCache)for(var a in e.depCache)o.depCache[r.normalizeSync(a)]=[].concat(e.depCache[a]);if(e.meta)for(var a in e.meta)if("*"===a[0])I(o.meta[a]=o.meta[a]||{},e.meta[a]);else{var v=V.call(r,o,a,void 0,!0,!0);I(o.meta[v]=o.meta[v]||{},e.meta[a])}"transpiler"in e&&(o.transpiler=e.transpiler);for(var y in e)-1===Er.indexOf(y)&&-1===zt.indexOf(y)&&(r[y]=e[y]);be(r,e,function(e){r.config(e,!0)})}function Se(){return{defaultExtension:void 0,main:void 0,format:void 0,meta:void 0,map:void 0,packageConfig:void 0,configured:!1}}function Ee(e,t,r,n,o){for(var i in t)"main"===i||"format"===i||"defaultExtension"===i||"configured"===i?n&&void 0!==e[i]||(e[i]=t[i]):"map"===i?(n?F:I)(e.map=e.map||{},t.map):"meta"===i?(n?F:I)(e.meta=e.meta||{},t.meta):Object.hasOwnProperty.call(t,i)&&C.call(o,'"'+i+'" is not a valid package configuration option in package '+r);return void 0===e.defaultExtension&&(e.defaultExtension="js"),void 0===e.main&&e.map&&e.map["."]?(e.main=e.map["."],delete e.map["."]):"object"==typeof e.main&&(e.map=e.map||{},e.map["./@main"]=e.main,e.main.default=e.main.default||"./",e.main="@main"),e}function je(e){return Nt?Gt+new Buffer(e).toString("base64"):"undefined"!=typeof btoa?Gt+btoa(unescape(encodeURIComponent(e))):""}function Pe(e,t,r,n){var o=e.lastIndexOf("\n");if(t){if("object"!=typeof t)throw new TypeError("load.metadata.sourceMap must be set to an object.");t=JSON.stringify(t)}return(n?"(function(System, SystemJS) {":"")+e+(n?"\n})(System, System);":"")+("\n//# sourceURL="!=e.substr(o,15)?"\n//# sourceURL="+r+(t?"!transpiled":""):"")+(t&&je(t)||"")}function Re(e,t,r,n,o){Jt||(Jt=document.head||document.body||document.documentElement);var i=document.createElement("script");i.text=Pe(t,r,n,!1);var a,s=window.onerror;return window.onerror=function(e){a=addToError(e,"Evaluating "+n),s&&s.apply(this,arguments)},Me(e),o&&i.setAttribute("nonce",o),Jt.appendChild(i),Jt.removeChild(i),_e(),window.onerror=s,a?a:void 0}function Me(e){0==Ht++&&(Wt=ut.System),ut.System=ut.SystemJS=e}function _e(){0==--Ht&&(ut.System=ut.SystemJS=Wt)}function Ce(e,t,r,n,o,i,a){if(t){if(i&&Zt)return Re(e,t,r,n,i);try{Me(e),!$t&&e._nodeRequire&&($t=e._nodeRequire("vm"),Bt=$t.runInThisContext("typeof System !== 'undefined' && System")===e),Bt?$t.runInThisContext(Pe(t,r,n,!a),{filename:n+(r?"!transpiled":"")}):(0,eval)(Pe(t,r,n,!a)),_e()}catch(e){return _e(),e}}}function Le(e){return"file:///"===e.substr(0,8)?e.substr(7+!!st):Xt&&e.substr(0,Xt.length)===Xt?e.substr(Xt.length):e}function Ae(e,t){return Le(this.normalizeSync(e,t))}function Ie(e){var t,r=e.lastIndexOf("!");t=-1!==r?e.substr(0,r):e;var n=t.split("/");return n.pop(),n=n.join("/"),{filename:Le(t),dirname:Le(n)}}function Fe(e){function t(e,t){for(var r=0;r<e.length;r++)if(e[r][0]<t.index&&e[r][1]>t.index)return!0;return!1}Ft.lastIndex=rr.lastIndex=nr.lastIndex=0;var r,n=[],o=[],i=[];if(e.length/e.split("\n").length<200){for(;r=nr.exec(e);)o.push([r.index,r.index+r[0].length]);for(;r=rr.exec(e);)t(o,r)||i.push([r.index+r[1].length,r.index+r[0].length-1])}for(;r=Ft.exec(e);)if(!t(o,r)&&!t(i,r)){var a=r[1].substr(1,r[1].length-2);if(a.match(/"|'/))continue;n.push(a)}return n}function Ke(e){if(-1===or.indexOf(e)){try{var t=ut[e]}catch(t){or.push(e)}this(e,t)}}function De(e){if("string"==typeof e)return U(e,ut);if(!(e instanceof Array))throw new Error("Global exports must be a string or array.");for(var t={},r=0;r<e.length;r++)t[e[r].split(".").pop()]=U(e[r],ut);return t}function qe(e,t,r,n){var o=ut.define;ut.define=void 0;var i;if(r){i={};for(var a in r)i[a]=ut[a],ut[a]=r[a]}return t||(Qt={},Object.keys(ut).forEach(Ke,function(e,t){Qt[e]=t})),function(){var e,r=t?De(t):{},a=!!t;if((!t||n)&&Object.keys(ut).forEach(Ke,function(o,i){Qt[o]!==i&&void 0!==i&&(n&&(ut[o]=void 0),t||(r[o]=i,void 0!==e?a||e===i||(a=!0):e=i))}),r=a?r:e,i)for(var s in i)ut[s]=i[s];return ut.define=o,r}}function Te(e,t){e=e.replace(rr,"");var r=e.match(sr),n=(r[1].split(",")[t]||"require").replace(ur,""),o=lr[n]||(lr[n]=new RegExp(ir+n+ar,"g"));o.lastIndex=0;for(var i,a=[];i=o.exec(e);)a.push(i[2]||i[3]);return a}function Ue(e){return function(t,r,n){e(t,r,n),r=n.exports,"object"!=typeof r&&"function"!=typeof r||"__esModule"in r||Object.defineProperty(n.exports,"__esModule",{value:!0})}}function ze(e,t){er=e,fr=t,Vt=void 0,cr=!1}function Ne(e){Vt?e.registerDynamic(er?Vt[0].concat(er):Vt[0],!1,fr?Ue(Vt[1]):Vt[1]):cr&&e.registerDynamic([],!1,R)}function Je(e,t){!e.load.esModule||"object"!=typeof t&&"function"!=typeof t||"__esModule"in t||Object.defineProperty(t,"__esModule",{value:!0})}function $e(e,t){var r=this,n=this[jt];return(We(n,this,e)||St).then(function(){if(!t()){var o=r[Pt][e];if("@node/"===e.substr(0,6)){if(!r._nodeRequire)throw new TypeError("Error loading "+e+". Can only load node core modules in Node.");return r.registerDynamic([],!1,function(){return A.call(r,e.substr(6),r.baseURL)}),void t()}return o.load.scriptLoad?(o.load.pluginKey||!dr)&&(o.load.scriptLoad=!1,C.call(n,'scriptLoad not supported for "'+e+'"')):o.load.scriptLoad!==!1&&!o.load.pluginKey&&dr&&(o.load.deps||o.load.globals||!("system"===o.load.format||"register"===o.load.format||"global"===o.load.format&&o.load.exports)||(o.load.scriptLoad=!0)),o.load.scriptLoad?new Promise(function(n,i){if("amd"===o.load.format&&ut.define!==r.amdDefine)throw new Error("Loading AMD with scriptLoad requires setting the global `"+gr+".define = SystemJS.amdDefine`");T(e,o.load.crossOrigin,o.load.integrity,function(){if(!t()){o.load.format="global";var e=o.load.exports&&De(o.load.exports);r.registerDynamic([],!1,function(){return Je(o,e),e}),t()}n()},i)}):Be(r,e,o).then(function(){return Ge(r,e,o,t,n.wasm)})}}).then(function(t){return delete r[Pt][e],t})}function Be(e,t,r){return r.pluginKey?e.import(r.pluginKey).then(function(e){r.pluginModule=e,r.pluginLoad={name:t,address:r.pluginArgument,source:void 0,metadata:r.load},r.load.deps=r.load.deps||[]}):St}function We(e,t,r){var n=e.depCache[r];if(n)for(var o=0;o<n.length;o++)t.normalize(n[o],r).then(D);else{var i=!1;for(var a in e.bundles){for(var o=0;o<e.bundles[a].length;o++){var s=e.bundles[a][o];if(s===r){i=!0;break}if(-1!==s.indexOf("*")){var u=s.split("*");if(2!==u.length){e.bundles[a].splice(o--,1);continue}if(r.substr(0,u[0].length)===u[0]&&r.substr(r.length-u[1].length,u[1].length)===u[1]){i=!0;break}}}if(i)return t.import(a)}}}function Ge(e,t,r,n,o){return r.load.exports&&!r.load.format&&(r.load.format="global"),St.then(function(){return r.pluginModule&&r.pluginModule.locate?Promise.resolve(r.pluginModule.locate.call(e,r.pluginLoad)).then(function(e){e&&(r.pluginLoad.address=e)}):void 0}).then(function(){return r.pluginModule?(o=!1,r.pluginModule.fetch?r.pluginModule.fetch.call(e,r.pluginLoad,function(e){return Dt(e.address,r.load.authorization,r.load.integrity,!1)}):Dt(r.pluginLoad.address,r.load.authorization,r.load.integrity,!1)):Dt(t,r.load.authorization,r.load.integrity,o)}).then(function(i){return o&&"string"!=typeof i?L(e,i,n).then(function(o){if(!o){var a=it?new TextDecoder("utf-8").decode(new Uint8Array(i)):i.toString();return He(e,t,a,r,n)}}):He(e,t,i,r,n)})}function He(e,t,r,n,o){return Promise.resolve(r).then(function(t){return"detect"===n.load.format&&(n.load.format=void 0),et(t,n),n.pluginModule&&n.pluginModule.translate?(n.pluginLoad.source=t,Promise.resolve(n.pluginModule.translate.call(e,n.pluginLoad,n.traceOpts)).then(function(e){if(n.load.sourceMap){if("object"!=typeof n.load.sourceMap)throw new Error("metadata.load.sourceMap must be set to an object.");Ye(n.pluginLoad.address,n.load.sourceMap)}return"string"==typeof e?e:n.pluginLoad.source})):t}).then(function(r){return n.load.format||'"bundle"'!==r.substring(0,8)?"register"===n.load.format||!n.load.format&&Ze(r)?(n.load.format="register",r):"esm"===n.load.format||!n.load.format&&r.match(hr)?(n.load.format="esm",Qe(e,r,t,n,o)):r:(n.load.format="system",r)}).then(function(t){if("string"!=typeof t||!n.pluginModule||!n.pluginModule.instantiate)return t;var r=!1;return n.pluginLoad.source=t,Promise.resolve(n.pluginModule.instantiate.call(e,n.pluginLoad,function(e){if(t=e.source,n.load=e.metadata,r)throw new Error("Instantiate must only be called once.");r=!0})).then(function(e){return r?t:M(e)})}).then(function(r){if("string"!=typeof r)return r;n.load.format||(n.load.format=Xe(r));var i=!1;switch(n.load.format){case"esm":case"register":case"system":var a=Ce(e,r,n.load.sourceMap,t,n.load.integrity,n.load.nonce,!1);if(a)throw a;if(!o())return Et;return;case"json":return e.newModule({
    default:JSON.parse(r),__useDefault:!0});case"amd":var s=ut.define;ut.define=e.amdDefine,ze(n.load.deps,n.load.esModule);var a=Ce(e,r,n.load.sourceMap,t,n.load.integrity,n.load.nonce,!1);if(i=o(),i||(Ne(e),i=o()),ut.define=s,a)throw a;break;case"cjs":var u=n.load.deps,l=(n.load.deps||[]).concat(n.load.cjsRequireDetection?Fe(r):[]);for(var c in n.load.globals)n.load.globals[c]&&l.push(n.load.globals[c]);e.registerDynamic(l,!0,function(o,i,a){if(o.resolve=function(t){return Ae.call(e,t,a.id)},a.paths=[],a.require=o,!n.load.cjsDeferDepsExecute&&u)for(var s=0;s<u.length;s++)o(u[s]);var l=Ie(a.id),c={exports:i,args:[o,i,a,l.filename,l.dirname,ut,ut]},f="(function (require, exports, module, __filename, __dirname, global, GLOBAL";if(n.load.globals)for(var d in n.load.globals)c.args.push(o(n.load.globals[d])),f+=", "+d;var p=ut.define;ut.define=void 0,ut.__cjsWrapper=c,r=f+") {"+r.replace(br,"")+"\n}).apply(__cjsWrapper.exports, __cjsWrapper.args);";var g=Ce(e,r,n.load.sourceMap,t,n.load.integrity,n.load.nonce,!1);if(g)throw g;Je(n,i),ut.__cjsWrapper=void 0,ut.define=p}),i=o();break;case"global":var l=n.load.deps||[];for(var c in n.load.globals){var f=n.load.globals[c];f&&l.push(f)}e.registerDynamic(l,!1,function(o,i,a){var s;if(n.load.globals){s={};for(var u in n.load.globals)n.load.globals[u]&&(s[u]=o(n.load.globals[u]))}var l=n.load.exports;l&&(r+="\n"+gr+'["'+l+'"] = '+l+";");var c=qe(a.id,l,s,n.load.encapsulateGlobal),f=Ce(e,r,n.load.sourceMap,t,n.load.integrity,n.load.nonce,!0);if(f)throw f;var d=c();return Je(n,d),d}),i=o();break;default:throw new TypeError('Unknown module format "'+n.load.format+'" for "'+t+'".'+("es6"===n.load.format?' Use "esm" instead here.':""))}if(!i)throw new Error("Module "+t+" detected as "+n.load.format+" but didn't execute correctly.")})}function Ze(e){var t=e.match(mr);return t&&"System.register"===e.substr(t[0].length,15)}function Xe(e){return e.match(vr)?"amd":(yr.lastIndex=0,Ft.lastIndex=0,Ft.exec(e)||yr.exec(e)?"cjs":"global")}function Ye(e,t){var r=e.split("!")[0];t.file&&t.file!=e||(t.file=r+"!transpiled"),(!t.sources||t.sources.length<=1&&(!t.sources[0]||t.sources[0]===e))&&(t.sources=[r])}function Qe(e,r,n,o,i){if(!e.transpiler)throw new TypeError("Unable to dynamically transpile ES module\n   A loader plugin needs to be configured via `SystemJS.config({ transpiler: 'transpiler-module' })`.");if(o.load.deps){for(var a="",s=0;s<o.load.deps.length;s++)a+='import "'+o.load.deps[s]+'"; ';r=a+r}return e.import.call(e,e.transpiler).then(function(t){if(t.__useDefault&&(t=t.default),!t.translate)throw new Error(e.transpiler+" is not a valid transpiler plugin.");return t===o.pluginModule?r:("string"==typeof o.load.sourceMap&&(o.load.sourceMap=JSON.parse(o.load.sourceMap)),o.pluginLoad=o.pluginLoad||{name:n,address:n,source:r,metadata:o.load},o.load.deps=o.load.deps||[],Promise.resolve(t.translate.call(e,o.pluginLoad,o.traceOpts)).then(function(e){var t=o.load.sourceMap;return t&&"object"==typeof t&&Ye(n,t),"esm"===o.load.format&&Ze(e)&&(o.load.format="register"),e}))},function(e){throw t(e,"Unable to load transpiler to transpile "+n)})}function Ve(e,t,r){for(var n,o=t.split(".");o.length>1;)n=o.shift(),e=e[n]=e[n]||{};n=o.shift(),void 0===e[n]&&(e[n]=r)}function et(e,t){var r=e.match(wr);if(r)for(var n=r[0].match(kr),o=0;o<n.length;o++){var i=n[o],a=i.length,s=i.substr(0,1);if(";"==i.substr(a-1,1)&&a--,'"'==s||"'"==s){var u=i.substr(1,i.length-3),l=u.substr(0,u.indexOf(" "));if(l){var c=u.substr(l.length+1,u.length-l.length-1);"deps"===l&&(l="deps[]"),"[]"===l.substr(l.length-2,2)?(l=l.substr(0,l.length-2),t.load[l]=t.load[l]||[],t.load[l].push(c)):"use"!==l&&Ve(t.load,l,c)}else t.load[u]=!0}}}function tt(){f.call(this),this._loader={},this[Pt]={},this[jt]={baseURL:ot,paths:{},packageConfigPaths:[],packageConfigKeys:[],map:{},packages:{},depCache:{},meta:{},bundles:{},production:!1,transpiler:void 0,loadedBundles:{},warnings:!1,pluginFirst:!1,wasm:!1},this.scriptSrc=pr,this._nodeRequire=tr,this.registry.set("@empty",Et),rt.call(this,!1,!1),Yt(this)}function rt(e,t){this[jt].production=e,this.registry.set("@system-env",Sr=this.newModule({browser:it,node:!!this._nodeRequire,production:!t&&e,dev:t||!e,build:t,default:!0}))}function nt(e,t){C.call(e[jt],"SystemJS."+t+" is deprecated for SystemJS.registry."+t)}var ot,it="undefined"!=typeof window&&"undefined"!=typeof document,at="undefined"!=typeof process&&process.versions&&process.versions.node,st="undefined"!=typeof process&&"string"==typeof process.platform&&process.platform.match(/^win/),ut="undefined"!=typeof self?self:global,lt="undefined"!=typeof Symbol;if("undefined"!=typeof document&&document.getElementsByTagName){if(ot=document.baseURI,!ot){var ct=document.getElementsByTagName("base");ot=ct[0]&&ct[0].href||window.location.href}}else"undefined"!=typeof location&&(ot=location.href);if(ot){ot=ot.split("#")[0].split("?")[0];var ft=ot.lastIndexOf("/");-1!==ft&&(ot=ot.substr(0,ft+1))}else{if("undefined"==typeof process||!process.cwd)throw new TypeError("No environment baseURI");ot="file://"+(st?"/":"")+process.cwd(),st&&(ot=ot.replace(/\\/g,"/"))}"/"!==ot[ot.length-1]&&(ot+="/");var dt="_"==new Error(0,"_").fileName,pt=Promise.resolve();i.prototype.constructor=i,i.prototype.import=function(e,r){if("string"!=typeof e)throw new TypeError("Loader import method must be passed a module key string");var n=this;return pt.then(function(){return n[ht](e,r)}).then(a).catch(function(n){throw t(n,"Loading "+e+(r?" from "+r:""))})};var gt=i.resolve=e("resolve"),ht=i.resolveInstantiate=e("resolveInstantiate");i.prototype[ht]=function(e,t){var r=this;return r.resolve(e,t).then(function(e){return r.registry.get(e)})},i.prototype.resolve=function(e,r){var n=this;return pt.then(function(){return n[gt](e,r)}).then(s).catch(function(n){throw t(n,"Resolving "+e+(r?" to "+r:""))})};var mt="undefined"!=typeof Symbol&&Symbol.iterator,vt=e("registry");mt&&(u.prototype[Symbol.iterator]=function(){return this.entries()[Symbol.iterator]()},u.prototype.entries=function(){var e=this[vt];return o(Object.keys(e).map(function(t){return[t,e[t]]}))}),u.prototype.keys=function(){return o(Object.keys(this[vt]))},u.prototype.values=function(){var e=this[vt];return o(Object.keys(e).map(function(t){return e[t]}))},u.prototype.get=function(e){return this[vt][e]},u.prototype.set=function(e,t){if(!(t instanceof l))throw new Error("Registry must be set with an instance of Module Namespace");return this[vt][e]=t,this},u.prototype.has=function(e){return Object.hasOwnProperty.call(this[vt],e)},u.prototype.delete=function(e){return Object.hasOwnProperty.call(this[vt],e)?(delete this[vt][e],!0):!1};var yt=e("baseObject");l.prototype=Object.create(null),"undefined"!=typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(l.prototype,Symbol.toStringTag,{value:"Module"});var bt=e("register-internal");f.prototype=Object.create(i.prototype),f.prototype.constructor=f;var wt=f.instantiate=e("instantiate");f.prototype[f.resolve=i.resolve]=function(e,t){return n(e,t||ot)},f.prototype[wt]=function(e,t){},f.prototype[i.resolveInstantiate]=function(e,t){var r=this,n=this[bt],o=r.registry[r.registry._registry];return p(r,e,t,o,n).then(function(e){return e instanceof l?e:e.module?e.module:e.linkRecord.linked?O(r,e,e.linkRecord,o,n,void 0):w(r,e,e.linkRecord,o,n,[e]).then(function(){return O(r,e,e.linkRecord,o,n,void 0)}).catch(function(t){throw k(r,e),t})})},f.prototype.register=function(e,t,r){var n=this[bt];if(void 0===r)n.lastRegister=[e,t,void 0];else{var o=n.records[e]||d(n,e,void 0);o.registration=[t,r,void 0]}},f.prototype.registerDynamic=function(e,t,r,n){var o=this[bt];if("string"!=typeof e)o.lastRegister=[e,t,r];else{var i=o.records[e]||d(o,e,void 0);i.registration=[t,r,n]}},x.prototype.import=function(e){return this.loader.trace&&y(this.loader,this.key,e),this.loader.import(e,this.key)};var kt={};Object.freeze&&Object.freeze(kt);var xt,Ot,St=Promise.resolve(),Et=new l({}),jt=e("loader-config"),Pt=e("metadata"),Rt="undefined"==typeof window&&"undefined"!=typeof self&&"undefined"!=typeof importScripts,Mt=!1,_t=!1;if(it&&function(){var e=document.createElement("link").relList;if(e&&e.supports){_t=!0;try{Mt=e.supports("preload")}catch(e){}}}(),it){var Ct=[],Lt=window.onerror;window.onerror=function(e,t){for(var r=0;r<Ct.length;r++)if(Ct[r].src===t)return void Ct[r].err(e);Lt&&Lt.apply(this,arguments)}}var At,It,Ft=/(?:^\uFEFF?|[^$_a-zA-Z\xA0-\uFFFF."'])require\s*\(\s*("[^"\\]*(?:\\.[^"\\]*)*"|'[^'\\]*(?:\\.[^'\\]*)*')\s*\)/g,Kt="undefined"!=typeof XMLHttpRequest;It="undefined"!=typeof self&&"undefined"!=typeof self.fetch?$:Kt?B:"undefined"!=typeof require&&"undefined"!=typeof process?W:G;var Dt=It,qt={},Tt=["browser","node","dev","build","production","default"],Ut=/#\{[^\}]+\}/,zt=["browserConfig","nodeConfig","devConfig","buildConfig","productionConfig"],Nt="undefined"!=typeof Buffer;try{Nt&&"YQ=="!==new Buffer("a").toString("base64")&&(Nt=!1)}catch(e){Nt=!1}var Jt,$t,Bt,Wt,Gt="\n//# sourceMappingURL=data:application/json;base64,",Ht=0,Zt=!1;it&&"undefined"!=typeof document&&document.getElementsByTagName&&(window.chrome&&window.chrome.extension||navigator.userAgent.match(/^Node\.js/)||(Zt=!0));var Xt,Yt=function(e){function t(r,n,o,i){if("object"==typeof r&&!(r instanceof Array))return t.apply(null,Array.prototype.splice.call(arguments,1,arguments.length-1));if("string"==typeof r&&"function"==typeof n&&(r=[r]),!(r instanceof Array)){if("string"==typeof r){var a=e.decanonicalize(r,i),s=e.get(a);if(!s)throw new Error('Module not already loaded loading "'+r+'" as '+a+(i?' from "'+i+'".':"."));return s.__useDefault?s.default:s}throw new TypeError("Invalid require")}for(var u=[],l=0;l<r.length;l++)u.push(e.import(r[l],i));Promise.all(u).then(function(e){for(var t=0;t<e.length;t++)e[t]=e[t].__useDefault?e[t].default:e[t];n&&n.apply(null,e)},o)}function r(r,n,o){function i(r,i,l){for(var c=[],f=0;f<n.length;f++)c.push(r(n[f]));if(l.uri=l.id,l.config=R,-1!==u&&c.splice(u,0,l),-1!==s&&c.splice(s,0,i),-1!==a){var d=function(n,o,i){return"string"==typeof n&&"function"!=typeof o?r(n):t.call(e,n,o,i,l.id)};d.toUrl=function(t){return e.normalizeSync(t,l.id)},c.splice(a,0,d)}var p=ut.require;ut.require=t;var g=o.apply(-1===s?ut:i,c);ut.require=p,"undefined"!=typeof g&&(l.exports=g)}"string"!=typeof r&&(o=n,n=r,r=null),n instanceof Array||(o=n,n=["require","exports","module"].splice(0,o.length)),"function"!=typeof o&&(o=function(e){return function(){return e}}(o)),r||er&&(n=n.concat(er),er=void 0);var a,s,u;-1!==(a=n.indexOf("require"))&&(n.splice(a,1),r||(n=n.concat(Te(o.toString(),a)))),-1!==(s=n.indexOf("exports"))&&n.splice(s,1),-1!==(u=n.indexOf("module"))&&n.splice(u,1),r?(e.registerDynamic(r,n,!1,i),Vt?(Vt=void 0,cr=!0):cr||(Vt=[n,i])):e.registerDynamic(n,!1,fr?Ue(i):i)}e.set("@@cjs-helpers",e.newModule({requireResolve:Ae.bind(e),getPathVars:Ie})),e.set("@@global-helpers",e.newModule({prepareGlobal:qe})),r.amd={},e.amdDefine=r,e.amdRequire=t};"undefined"!=typeof window&&"undefined"!=typeof document&&window.location&&(Xt=location.protocol+"//"+location.hostname+(location.port?":"+location.port:""));var Qt,Vt,er,tr,rr=/(^|[^\\])(\/\*([\s\S]*?)\*\/|([^:]|^)\/\/(.*)$)/gm,nr=/("[^"\\\n\r]*(\\.[^"\\\n\r]*)*"|'[^'\\\n\r]*(\\.[^'\\\n\r]*)*')/g,or=["_g","sessionStorage","localStorage","clipboardData","frames","frameElement","external","mozAnimationStartTime","webkitStorageInfo","webkitIndexedDB","mozInnerScreenY","mozInnerScreenX"],ir="(?:^|[^$_a-zA-Z\\xA0-\\uFFFF.])",ar="\\s*\\(\\s*(\"([^\"]+)\"|'([^']+)')\\s*\\)",sr=/\(([^\)]*)\)/,ur=/^\s+|\s+$/g,lr={},cr=!1,fr=!1,dr=(it||Rt)&&"undefined"!=typeof navigator&&navigator.userAgent&&!navigator.userAgent.match(/MSIE (9|10).0/);"undefined"==typeof require||"undefined"==typeof process||process.browser||(tr=require);var pr,gr="undefined"!=typeof self?"self":"global",hr=/(^\s*|[}\);\n]\s*)(import\s*(['"]|(\*\s+as\s+)?[^"'\(\)\n;]+\s*from\s*['"]|\{)|export\s+\*\s+from\s+["']|export\s*(\{|default|function|class|var|const|let|async\s+function))/,mr=/^(\s*\/\*[^\*]*(\*(?!\/)[^\*]*)*\*\/|\s*\/\/[^\n]*|\s*"[^"]+"\s*;?|\s*'[^']+'\s*;?)*\s*/,vr=/(?:^\uFEFF?|[^$_a-zA-Z\xA0-\uFFFF.])define\s*\(\s*("[^"]+"\s*,\s*|'[^']+'\s*,\s*)?\s*(\[(\s*(("[^"]+"|'[^']+')\s*,|\/\/.*\r?\n|\/\*(.|\s)*?\*\/))*(\s*("[^"]+"|'[^']+')\s*,?)?(\s*(\/\/.*\r?\n|\/\*(.|\s)*?\*\/))*\s*\]|function\s*|{|[_$a-zA-Z\xA0-\uFFFF][_$a-zA-Z0-9\xA0-\uFFFF]*\))/,yr=/(?:^\uFEFF?|[^$_a-zA-Z\xA0-\uFFFF.])(exports\s*(\[['"]|\.)|module(\.exports|\['exports'\]|\["exports"\])\s*(\[['"]|[=,\.]))/,br=/^\#\!.*/,wr=/^(\s*\/\*[^\*]*(\*(?!\/)[^\*]*)*\*\/|\s*\/\/[^\n]*|\s*"[^"]+"\s*;?|\s*'[^']+'\s*;?)+/,kr=/\/\*[^\*]*(\*(?!\/)[^\*]*)*\*\/|\/\/[^\n]*|"[^"]+"\s*;?|'[^']+'\s*;?/g;if("undefined"==typeof Promise)throw new Error("SystemJS needs a Promise polyfill.");if("undefined"!=typeof document){var xr=document.getElementsByTagName("script"),Or=xr[xr.length-1];document.currentScript&&(Or.defer||Or.async)&&(Or=document.currentScript),pr=Or&&Or.src}else if("undefined"!=typeof importScripts)try{throw new Error("_")}catch(e){e.stack.replace(/(?:at|@).*(http.+):[\d]+:[\d]+/,function(e,t){pr=t})}else"undefined"!=typeof __filename&&(pr=__filename);var Sr;tt.prototype=Object.create(f.prototype),tt.prototype.constructor=tt,tt.prototype[tt.resolve=f.resolve]=tt.prototype.normalize=X,tt.prototype.load=function(e,t){return C.call(this[jt],"System.load is deprecated."),this.import(e,t)},tt.prototype.decanonicalize=tt.prototype.normalizeSync=tt.prototype.resolveSync=Q,tt.prototype[tt.instantiate=f.instantiate]=$e,tt.prototype.config=Oe,tt.prototype.getConfig=xe,tt.prototype.global=ut,tt.prototype.import=function(){return f.prototype.import.apply(this,arguments).then(function(e){return e.__useDefault?e.default:e})};for(var Er=["baseURL","map","paths","packages","packageConfigPaths","depCache","meta","bundles","transpiler","warnings","pluginFirst","production","wasm"],jr="undefined"!=typeof Proxy,Pr=0;Pr<Er.length;Pr++)(function(e){Object.defineProperty(tt.prototype,e,{get:function(){var t=ke(this[jt],e);return jr&&"object"==typeof t&&(t=new Proxy(t,{set:function(t,r){throw new Error("Cannot set SystemJS."+e+'["'+r+'"] directly. Use SystemJS.config({ '+e+': { "'+r+'": ... } }) rather.')}})),t},set:function(t){throw new Error("Setting `SystemJS."+e+"` directly is no longer supported. Use `SystemJS.config({ "+e+": ... })`.")}})})(Er[Pr]);tt.prototype.delete=function(e){nt(this,"delete"),this.registry.delete(e)},tt.prototype.get=function(e){return nt(this,"get"),this.registry.get(e)},tt.prototype.has=function(e){return nt(this,"has"),this.registry.has(e)},tt.prototype.set=function(e,t){return nt(this,"set"),this.registry.set(e,t)},tt.prototype.newModule=function(e){return new l(e)},tt.prototype.isModule=_,tt.prototype.register=function(e,t,r){return"string"==typeof e&&(e=Y.call(this,this[jt],e)),f.prototype.register.call(this,e,t,r)},tt.prototype.registerDynamic=function(e,t,r,n){return"string"==typeof e&&(e=Y.call(this,this[jt],e)),f.prototype.registerDynamic.call(this,e,t,r,n)},tt.prototype.version="0.20.12 Dev";var Rr=new tt;(it||Rt)&&(ut.SystemJS=ut.System=Rr),"undefined"!=typeof module&&module.exports&&(module.exports=Rr)}();
//# sourceMappingURL=system.js.map
