
export class ParseJSON {

  static get components() {
    return {
      'SELECT'    : Select,
      'INPUT'     : Input,
      'TEXTAREA'  : TextArea
    };
  }


  /*
   *   When need recursion for table
   */

  static parseData(data, callback, strForTable = '') {

    for (let id in data) {

      if (typeof id === 'string') {
        let component;

        if (strForTable.length === 0) {
          component = document.getElementById(id);
        } else {
          component = document.getElementById(`${ strForTable }:${ id }`);
        }

        if (component) {
          callback(component, data[id]);
        }

        // has prefix "tableid_" for recursion
        else if (id.startsWith(Variables.paramsJSONTable)) {
          this.parseData(data[id][Variables.paramsJSONList], callback, id.replace(new RegExp(`^${ Variables.paramsJSONTable }`), ''));
        }
      }

    }

  }


  /*
  *   set attributes to component
  */

  static setAttrToComponent(component, params = {}) {

    for (let attr in params) {

      // if attr === type
      if (attr === Variables.paramsJSONType) {
        if (params[attr] !== Variables.paramsJSONSet
          && params[attr] !== Variables.paramsJSONEnum) {
          component.setAttribute(attr, params[attr]);
        }
      }

      // if attr !== list
      else if (attr !== Variables.paramsJSONList
        && attr !== Variables.paramsJSONTitle) {
        component.setAttribute(attr, params[attr]);
      }

    }

    const func = this.components[component.tagName];

    if (func) {

      // if has attr in params 'list'
      if (params[Variables.paramsJSONList] && func.createList) {
        func.createList(component, params[Variables.paramsJSONList], (params[Variables.paramsJSONType] === Variables.paramsJSONSet));
      }

      // if has attr in params 'title'
      else if (params[Variables.paramsJSONTitle] && func.setDefaultAttr) {
        func.setDefaultAttr(component, params[Variables.paramsJSONTitle]);
      }

    } else {
      console.log(`Not found: ${ component.tagName }`);
    }

  }


  /*
   *   Insert data after create component
   */

  static insertValueToComponent(component, attr = '') {
    if (typeof attr === 'string' && attr.length !== 0) {
      const func = this.components[component.tagName];
      if (func && func.addAttrToComponent) {
        func.addAttrToComponent(component, attr);
      } else {
        console.log(`Not found: ${ component.tagName }`);
      }
    }
  }


  static insertDataToAttrSetText(component, textContent = '') {
    if (component.children.length !== 0) {
      for (let child of component.children) {
        this.insertDataToAttrSetText(child, textContent);
      }
    }
    if (component.hasAttribute(Variables.paramsJSONSetText)) {
      component.textContent = textContent;
    }
  }

}