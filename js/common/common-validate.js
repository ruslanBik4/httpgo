'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var classNameError = 'error-validate';
var attr = 'data-validate';

var attrSelect = 'data-validate-select'; // data-validate-select="first, 1, 2, ..., last" => value attr to option
var attrCheckbox = 'data-validate-checkbox';

var Validate = exports.Validate = function () {
  function Validate() {
    _classCallCheck(this, Validate);
  }

  _createClass(Validate, null, [{
    key: 'validate',
    value: function validate(data) {
      var isTest = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : false;

      var isError = false;

      data.querySelectorAll('[' + attr + ']').forEach(function (component) {

        var componentChild = component.querySelector('select');

        // select

        if (componentChild) {
          var validateSelect = component.getAttribute(attrSelect);

          if (validateSelect) {

            var result = false;
            var indexSelected = componentChild.selectedIndex;
            var valueSelect = componentChild.options[indexSelected].value;

            validateSelect.split(', ').forEach(function (value) {
              if (value === 'first') {
                result = indexSelected == 0;
              } else if (value === 'last') {
                result = indexSelected == componentChild.options.length - 1;
              } else {
                result = value == valueSelect;
              }
              if (result) {
                return false;
              }
            });

            if (result) {
              if (!isTest) componentChild.classList.add(classNameError);
              isError = true;
            } else {
              componentChild.classList.remove(classNameError);
            }
          }
          return;
        }

        // input

        if (component.hasAttribute(attrCheckbox)) {
          var _result = false;
          component.querySelectorAll('input[type="checkbox"]').forEach(function (input) {
            if (input.checked) {
              _result = true;
            }
          });
          isError = !_result;
          return;
        }

        componentChild = component.querySelector('input');

        if (componentChild) {
          if (componentChild.value.length === 0) {
            if (!isTest) componentChild.classList.add(classNameError);
            isError = true;
          } else {
            componentChild.classList.remove(classNameError);
          }
          return;
        }

        //textarea

        componentChild = component.querySelector('textarea');

        if (componentChild) {
          if (componentChild.textLength === 0) {
            if (!isTest) componentChild.classList.add(classNameError);
            isError = true;
          } else {
            componentChild.classList.remove(classNameError);
          }
        }
      });

      return isError;
    }
  }]);

  return Validate;
}();