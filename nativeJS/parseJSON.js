
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

  static parseDataGet(data, callback, strForTable = '', isDataTable = false) {

    for (let id in data) {

      if (data[id] !== null) {
        let component;

        if (strForTable.length === 0) {
          component = document.getElementById(id);
        } else {
          component = document.getElementById(`${ strForTable }:${ id }`);
        }

        if (Native.isElement(component)) {
          callback(component, data[id]);
        }

        // has prefix "tableid_" for recursion
        else if (id.startsWith(Variables.paramsJSONTable)) {
          if (isDataTable) {
            callback(component, data[id], id.replace(new RegExp('^' + Variables.paramsJSONTable), ''));
          } else {
            this.parseDataGet(data[id][Variables.paramsJSONList], callback, id.replace(new RegExp(`^${ Variables.paramsJSONTable }`), ''), isDataTable);
          }
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

    try {
      if (func) {

        // if has attr in params 'list'
        if (typeof params[Variables.paramsJSONList] === 'object' && func.createList) {
          func.createList(component, params[Variables.paramsJSONList], (params[Variables.paramsJSONType] === Variables.paramsJSONSet));
        }

        // if has attr in params 'title'
        else if (params[Variables.paramsJSONTitle] && func.setDefaultAttr) {
          func.setDefaultAttr(component, params[Variables.paramsJSONTitle]);
        }

      } else {
        console.log(`Not found in frame: ${ component.tagName }`);
      }
    }
    catch (e) {
      console.log(e, component, params);
    }

  }


  /*
   *   Insert data after create component
   */

  static insertValueToComponent(component, attr, strForTable = '') {

    const insertValueCurrentComponent = (component, attr) => {

      let func = this.components[component.tagName];
      if (func && func.addAttrToComponent) {
        func.addAttrToComponent(component, attr);
      } else {
        if (Object.prototype.toString.call(attr) === '[object Array]') {
          for (let value of attr) {
            const currentComponent = component.querySelector(`[${ Variables.paramsJSONIdData }="${ value.id }"]`);
            if (!currentComponent) {
              console.log(`component data-id not found for set value: ${ component.id }, ${ value }`);
              continue;
            }
            if (!func) {
              func = this.components[currentComponent.tagName];
            }
            if (func.addAttrToComponent) {
              func.addAttrToComponent(currentComponent, "1");
            }
          }
        } else {
          component.textContent = attr;
          console.log(`Not found in frame: ${ component }`);
        }
      }

    };

    function setNewAttrIdAndName(component, index) {
      const nameAttr = component.getAttribute('name');
      const idAttr = component.getAttribute('id');
      component.setAttribute('name', `${ nameAttr }[${ index }]`);
      component.setAttribute('id', `${ idAttr }-${ index }`);
    }

    const tableIdParse = (data, strForTable) => {

      /* first, get parent and default component */
      let parent;
      let defaultComponent;
      let index = 0;

      for (let id in data[index]) {
        const component = document.getElementById(`${ strForTable }:${ id }`) || document.getElementById(`${ strForTable }:${ id }[]`);
        if (component) {
          if (!parent) {
            parent = Native.findAncestorByClass(component, Variables.paramsJSONIdForTable);
            const temp = document.createElement('template');
            temp.innerHTML = parent.innerHTML;
            defaultComponent = temp;
          }
          if (data[index][id].length !== 0) {
            insertValueCurrentComponent(component, data[index][id]);
          }
          setNewAttrIdAndName(component, index);
        }
      }

      if (!parent) {
        return;
      }

      /* secondary components */

      for (index++; index < data.length; index++) {
        const newComponent = defaultComponent.cloneNode(true);

        for (let id in data[index]) {

          const component = document.querySelector(`[name="${ strForTable }:${ id }"]`) || document.querySelector(`[name="${ strForTable }:${ id }[]"]`);

          if (component && data[index][id].length !== 0) {
            insertValueCurrentComponent(component, data[index][id]);
            setNewAttrIdAndName(component, index);
          }

        }

        parent.appendChild(newComponent.content);
      }

    };

    if (attr !== null) {

      if (strForTable.length !== 0 && Object.prototype.toString.call(attr) === '[object Array]' && attr.length !== 0) {
        tableIdParse(attr, strForTable);
      } else if (attr.length !== 0) {
        insertValueCurrentComponent(component, attr);
      }
    }

  }


  static insertDataToAttrSetText(component, textContent = '') {
    if (component.children.length !== 0) {
      for (let i = 0; i < component.children.length; i++) {
        this.insertDataToAttrSetText(component.children[i], textContent);
      }
    }
    if (component && component.hasAttribute(Variables.paramsJSONSetText)) {
      component.textContent = textContent;
    }
  }



  static setValue(component, attr, callback, str = '', isDefault = false, strTable = '') {

    for (let name in attr) {
      let nameField;

      if (strTable.length !== 0) {
        nameField = (isDefault) ? `${ strTable }:${ name }` : `${ strTable }:${ name }${ str }`;
      } else {
        nameField = (isDefault) ? `${ name }` : `${ name }${ str }`;
      }

      const dom = component.querySelector(`[name="${ nameField }"]`);

      if (name.startsWith(Variables.paramsJSONTable)) {
        this.setValue(component, attr[name], callback, str, isDefault, name.replace(new RegExp('^' + Variables.paramsJSONTable)));
      } else if (dom) {
        if (isDefault) {
          const intArray = str.match(/\d+/g);
          dom.setAttribute('name', `${ nameField }${ str }`);
          dom.setAttribute('id', `${ nameField }-${ (intArray) ? intArray.join('') : '' }`);
        }
        callback(dom, attr[name]);
      }

    }
  }

}