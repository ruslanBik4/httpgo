
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
            callback(component, data[id],  id.replace(new RegExp('^' + Variables.paramsJSONTable), ''));
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
      const func = this.components[component.tagName];
      if (func && func.addAttrToComponent) {
        return func.addAttrToComponent(component, attr);
      } else {
        component.textContent = attr;
        console.log(`Not found: ${ component.tagName }`);
      }
    };

    function setNewAttrIdAndName(component, id, index) {
      component.setAttribute('name', `${ id }[${ index }]`);
      component.setAttribute('id', `${ id }-${ index }`);
    }

    if (attr != null) {

      if (strForTable.length !== 0 &&  Object.prototype.toString.call(attr) === '[object Array]' && attr.length !== 0) {

        let isFirstComponent = true;
        let newComponents;
        let parent;

        attr.forEach((element, index) => {

          let addComponent;

          if (newComponents) {
            addComponent = newComponents.cloneNode(true);
          }

          for (let id in element) {

            let component;
            let idComponent = `${ strForTable }:${ id }`;

            if (isFirstComponent) {
              component = document.getElementById(idComponent);

              if (component) {
                if (!parent) {
                  parent = Native.findAncestorByClass(component, Variables.paramsJSONIdForTable);
                  const temp = document.createElement('template');
                  temp.innerHTML = parent.innerHTML;
                  newComponents = temp;
                } else {
                  insertValueCurrentComponent(component, element[id]);
                  setNewAttrIdAndName(component, idComponent, index);
                }
              }

            } else if (newComponents && parent) {
              try {
                // debugger;
                const newComponent = addComponent.content.querySelector(`[name="${ strForTable }:${ id }"]`);
                if (newComponent) {
                  insertValueCurrentComponent(newComponent, element[id]);
                  setNewAttrIdAndName(newComponent, idComponent, index);
                }
              } catch(e) {
                console.log(e);
                alert(e);
              }
            }

          }

          if (parent && addComponent) {
            parent.appendChild(addComponent.content);
          }

          isFirstComponent = false;

        });
        // const component = document.getElementById(`${ strForTable }:${ attr[id] }`);

      } else if (Native.isElement(component) && attr.length !== 0) {
        insertValueCurrentComponent(component, attr);
      }

    }
  }


  static insertDataToAttrSetText(component, textContent = '') {
    if (component && component.children && component.children.length !== 0) {
      for (let child of component.children) {
        this.insertDataToAttrSetText(child, textContent);
      }
    }
    if (component.hasAttribute(Variables.paramsJSONSetText)) {
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