
const components = {
  'SELECT'  : Select,
  'INPUT'   : Input
};


export class ParseJSON {


  /*
   *   When need recursion for table
   */

  static parseData(data, callback, strForTable = '') {

    for (let id in data) {

      let dom;

      if (strForTable.length !== 0) {
        dom = document.getElementById(id);
      } else {
        dom = document.getElementById(`${ strForTable }:${ id }`);
      }

      if (dom) {
        callback(dom, data[id]);
      }

      // has prefix "tableid_" for recursion
      else if (typeof id === 'string' && id.startsWith(Variables.paramsJSONTable)) {
        debugger;
        this._insertDataToDom(data[id], callback, id.replace(/^Variables.paramsJSONTable/, ''));
      }

    }

  }


  static setAttrToComponent(component, params = {}) {

    for (let attr in params) {
      if (attr !== Variables.paramsJSONList && attr !== Variables.paramsJSONUnknown) {
        component.setAttribute(attr, params[attr]);
      }
    }

    if (params[Variables.paramsJSONList]) {
      const func = components[component.tagName];
      func.createList(component, params[Variables.paramsJSONList]);
    } else {
      // this._addListToComponent(component, params);
    }

  }


  /*
   *   func for add list to Component
   */

  static addListToComponent(component, params) {
    let template = document.createElement('template');

    if (component.tagName === 'INPUT') {

      let title = component.getAttribute('title');

      switch (component.getAttribute('type')) {
        case 'search':
        case 'text':
        case 'time':
        case 'number':
        case 'checkbox':
          if (title) {
            component.parentElement.lastElementChild.textContent = title;
          }

          break;
        case 'radio':
          debugger;
          break;
        default:
      }

    } else if (component.tagName === 'TEXTAREA') {
      component.parentElement.lastElementChild.textContent = component.getAttribute('title');
    }
  }



  /*
   *   Insert data after create component
   */

  static insertValueToComponent(element, attr = '') {

    if (this.isElement(element) && attr != null) {
      switch (element.tagName) {

        // input

        case 'INPUT':
          switch (element.getAttribute('type')) {
            case 'text':
            case 'number':
              element.setAttribute('value', attr);
              break;
            case 'checkbox':
              if (attr !== "0") {
                element.setAttribute('checked', 'true');
              }
              break;
            default:
              break;
          }
          break;


        // select

        case 'SELECT':
          let number = parseInt(attr) - 1;
          if (0 <= number && number < element.children.length) {
            element.children[number].setAttribute('selected', '');
          }
          break;


        // textarea

        case 'TEXTAREA':
          element.value = attr;
          break;

        default:
          break;
      }
    }
  }

}