/*
 * Copyright (c) 2023. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */
"use strict";

var token = '';
var lang = 'en'
var userStruct = getUser();
var urlAfterLogin = '';

function getUser() {
    let user = localStorage.getItem("USER");
    if (user > '') {
        let userData = JSON.parse(user);
        saveUser(userData);
        console.log(userData)
        return userData;
    }
}

// handling response login form, save users data & render some properties
function afterLogin(userData, thisForm) {
    if (!userData) {
        alert("Need users data!");
        console.log(thisForm)
        return false;
    }

    localStorage.setItem("USER", JSON.stringify(userData));
    saveUser(userData);
    userStruct = userData;
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


function saveUser(userData) {
    var userSuffix = userData.lang ? `(${userData.lang})` : '';
    console.log(userData);
    document.getElementById('bLogin').textContent = userData.name + userSuffix;
    token = userData.token || userData.access_token || userData.bearer_token || userData.auth_token;

    $('#bLogin').text(userData.name + userSuffix);
    $('.auth').removeClass("auth");
    changeLang(userData.lang);

    if (ChangeTheme !== undefined) {
        ChangeTheme(userData.theme);
    }

    if (urlAfterLogin === '') {
        if (userData.formActions !== undefined) {
            urlAfterLogin = userData.formActions[0].url;
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
