'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.CurrentAlerts = undefined;

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

var _native = require('native');

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var Admin = exports.CurrentAlerts = function () {
  function Admin() {
    _classCallCheck(this, Admin);

    console.log('Admin create!');

    this.url = '';
    this.body = '324';
  }

  _createClass(Admin, [{
    key: 'func',
    value: function func(e) {
      e.preventDefault();
      var dom = e.target;
      _native.Native.request({
        url: dom.href,
        processData: false,
        success: (respo, url) => {
          this.url = url;
          this.body = respo;
        }
      });
    }
  }]);

  return Admin;
}();