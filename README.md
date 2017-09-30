# httpgo <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://camo.githubusercontent.com/5dcb57e59f46a4ed65fafc343ab810e35086e21d/68747470733a2f2f696d672e736869656c64732e696f2f3a6c6963656e73652d6170616368652d626c75652e737667" alt="License" data-canonical-src="https://img.shields.io/:license-apache-blue.svg" style="max-width:100%;"></a> <a href="https://godoc.org/github.com/ruslanBik4/httpgo"><img src="https://godoc.org/github.com/ruslanBik4/httpgo?status.svg" alt="GoDoc"></a> <a href="https://goreportcard.com/report/github.com/ruslanBik4/httpgo"><img src="https://goreportcard.com/badge/github.com/ruslanBik4/httpgo" alt="GoReport"/></a> <a href="https://travis-ci.org/ruslanBik4/httpgo.svg?branch=master" > <img src="https://travis-ci.org/ruslanBik4/httpgo.svg?branch=master" /> </a>
Веб-сервер без претензий на сложность
Я создал этот веб-сервер для облегчения перехода со старых языков веб-программирования на Go
и приглашаю всех желающий сделать его идеальным.

 Умеет исполнять php скрипты, взаимодействую через сокет Unix с php-fpm

Настройка
 Папка config содержит настройки системы и конфигурационные файлы для сервера:
 - httpgo.service, настройка демона (юрит сервиса systemd) для управления работой httpgo (запуск, перезапуск, восстановление после сбоев), подробнее читайте
<a href="https://wiki.archlinux.org/index.php/Systemd_(Русский)"> Основы использования systemd </a>
для использования достаточно перенести его в директорию юнитов systemd
 - php-fpm.conf, настройка php-fpm (для запуска скриптов PHP), для использования достаточно перенести его в директорию php-fpm
 - db.yml.sample, настройка соединения с сервером MySQL (следует ввести нужные значения и удалить суффикс .sample)
 - mongo.yml.sample, настройка для работы базы MongoDB (следует ввести нужные значения и удалить суффикс .sample)