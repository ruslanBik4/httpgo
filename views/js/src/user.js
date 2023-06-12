/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
var token = '';
var lang = 'en'
var userStruct;
var urlAfterLogin = '';

function getUser() {
    let user = localStorage.getItem("USER");
    if (user > '') {
        userStruct = JSON.parse(user);
        saveUser(userStruct);
    }
}

// handling response login form, save users data & render some properties
function afterLogin(userStruct, thisForm) {
    if (!userStruct) {
        alert("Need users data!");
        console.log(thisForm)
        return false;
    }

    localStorage.setItem("USER", JSON.stringify(userStruct));
    saveUser(userStruct);
    $('input[autofocus]:last').focus();

    return true;
}

function changeLang(newLang) {
    if (lang === newLang) {
        return false
    }

    lang = newLang;
    $.ajaxSetup({
        'headers': {'Authorization': 'Bearer ' + token, "Accept-Language": newLang}
    });
    $('.topline-navbar').load('/top_menu');
    $('.footer-mnu').load('/foot_menu');
    loadContent(document.location.href.replace(/lang=\d+/, ``));
    return false
}


function saveUser(userStruct) {
    var userSuffix = userStruct.lang ? `(${userStruct.lang})` : '';
    console.log(userStruct);
    document.getElementById('bLogin').textContent = userStruct.name + userSuffix;
    token = userStruct.token || userStruct.access_token || userStruct.bearer_token || userStruct.auth_token;

    $('#bLogin').text(userStruct.name + userSuffix);
    $('.auth').removeClass("auth");
    changeLang(userStruct.lang);

    if (ChangeTheme !== undefined) {
        ChangeTheme(userStruct.theme);
    }

    if (urlAfterLogin === '') {
        if (userStruct.formActions !== undefined) {
            urlAfterLogin = userStruct.formActions[0].url;
        } else if (urlAfterLogin.onsubmit !== undefined) {
            urlAfterLogin.onsubmit();
            urlAfterLogin = "";
            return;
        }
    }
    if (urlAfterLogin > '') {
        loadContent(urlAfterLogin);
    }
}
