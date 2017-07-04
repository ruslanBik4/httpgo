
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
          delete params[Variables.paramsJSONList].count;
          func.createList(component, params[Variables.paramsJSONList], (params[Variables.paramsJSONType] === Variables.paramsJSONSet));
        }

        // if has attr in params 'title'
        else if (params[Variables.paramsJSONTitle] && func.setDefaultAttr) {
          func.setDefaultAttr(component, params[Variables.paramsJSONTitle]);
        }

        // set default value TODO: refactor need default set value for select
        if (component.tagName !== 'SELECT' && params[Variables.paramsJSONDefault] && func.addAttrToComponent) {
          func.addAttrToComponent(component, params[Variables.paramsJSONDefault]);
        }

      } else {
        // console.log(`Not found in frame: ${ component.tagName }`);
      }
    }
    catch (e) {
      console.log(e, component, params);
    }

  }



  static insertValueCurrentComponent(component, attr) {

    let func = this.components[component.tagName];
    if (func && func.addAttrToComponent) {
      func.addAttrToComponent(component, attr);
    } else {
      if (Object.prototype.toString.call(attr) === '[object Array]') {
      } else {
        component.textContent = attr;
        // console.log(`Not found in frame: ${ component }`);
      }
    }

  };

  static setNewAttrIdAndName(component, index) {
    const nameAttr = component.getAttribute('name');
    const idAttr = component.getAttribute('id');
    component.setAttribute('name', `${ nameAttr }[${ index }]`);
    component.setAttribute('id', `${ idAttr }-${ index }`);
  }



  /*
   *   Insert data after create component
   */

  static insertValueToComponent(component, attr, strForTable = '') {

    function getDefaultComponent(parent) {
      if (parent) {
        const temp = document.createElement('template');
        temp.innerHTML = parent.innerHTML;
        // debugger;
        //
        // document.querySelectorAll(`[${ Variables.paramsForClick }="${ parent.getAttribute(Variables.paramsJSONIdForTable) }"]`).forEach((component) => {
        //   component.onclick = function() {
        //     debugger;
        //     const newComponent = temp.cloneNode(true);
        //     const parent = document.querySelector(`[${ Varibales.paramsJSONIdForTable }="${ this.getAttribute(Variables.paramsForClick) }"]`)
        //     parent.appendChild(newComponent);
        //   };
        // });

        return temp;
      }
    }

    const tableIdParse = (curComponent, data, strForTable) => {

      /* first, get parent and default component */
      let parent;
      let defaultComponent;
      let index = 0;

      for (let id in data[index]) {

        const [component] = this._getDom(curComponent, id, strForTable);

        if (component) {
          if (!parent) {
            parent = Native.findAncestorByClass(component, Variables.paramsJSONIdForTable);
            defaultComponent = getDefaultComponent(parent);
          }
          if (data[index][id].length !== 0) {
            this.insertValueCurrentComponent(component, data[index][id]);
          }
          this.setNewAttrIdAndName(component, index);
        }
      }

      if (!parent) {
        // let dom;
        // if (component.hasAttribute(Variables.paramsForm)) {
        //   dom = document.querySelector(`[name^="${ strForTable }"]`);
        // } else {
        //   dom = component.querySelector(`[name^="${ strForTable }"]`);
        // }
        // if (dom) {
        //   getDefaultComponent(Native.findAncestorByClass(dom, Variables.paramsJSONIdForTable));
        // }
        return;
      }

      /* secondary components */

      for (index++; index < data.length; index++) {
        const newComponent = defaultComponent.cloneNode(true);

        for (let id in data[index]) {

          const [component] = this._getDom(newComponent.content.firstElementChild, id, strForTable);

          if (component && data[index][id].length !== 0) {
            this.insertValueCurrentComponent(component, data[index][id]);
            this.setNewAttrIdAndName(component, index);
          }

        }

        parent.appendChild(newComponent.content);
      }

    };

    if (attr !== null) {
      if (strForTable.length !== 0 && Object.prototype.toString.call(attr) === '[object Array]') {
        tableIdParse(component, attr, strForTable);
      } else if (component && attr.length !== 0) {
        this.insertValueCurrentComponent(component, attr);
      }
      /* else if (Object.prototype.toString.call(attr) === '[object Array]') {
        debugger;
        const parent = Native.findAncestorByClass(component, Variables.paramsJSONIdForTable);
        if (parent) {
          getDefaultComponent(parent);
        }
      } */

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



  static setValue(component, attr, callback, str = '', isDefault = false, isOnlyClass = false, strTable = '') {

    if (strTable) {
      attr = attr[Variables.paramsJSONList];
    }

    for (let name in attr) {

      const [doms, nameField] = this._getDom(component, name, strTable, (isDefault || isOnlyClass) ? '' : str);

      for (let dom of doms) {

        if (name.startsWith(Variables.paramsJSONTable)) {
          if (isDefault) {
            this.setValue(component, attr[name], callback, str, isDefault, isOnlyClass, name.replace(new RegExp('^' + Variables.paramsJSONTable), ''));
          } else {
            callback(component, attr[name], name.replace(new RegExp('^' + Variables.paramsJSONTable), ''));
          }
        } else if (dom) {
          if (isDefault) {

            if (strTable.length !== 0) {

              const parent = Native.findAncestorByClass(dom, Variables.paramsJSONIdForTable);

              if (parent && parent.getAttribute(Variables.paramsJSONIdForTable).length === 0) {

                const idParent = dom.getAttribute(Variables.paramsJSONIdData);
                parent.setAttribute(Variables.paramsJSONIdForTable, idParent);

                const temp = document.createElement('template');
                temp.innerHTML = parent.innerHTML;

                document.querySelectorAll(`[${ Variables.paramsForClick }="${ parent.getAttribute(Variables.paramsJSONIdForTable) }"]`).forEach((component) => {
                  component.onclick = () => {

                    const newComponent = temp.cloneNode(true);
                    const index = parent.children.length;

                    newComponent.content.querySelectorAll(`[${ Variables.paramsChangeId }]`).forEach(function () {
                      this.setAttribute('id', this.getAttribute('id') + '-' + index);
                    });

                    for (let id in attr) {
                      const component = newComponent.content.querySelector(`[name="${ strTable }:${ id }"]`) || newComponent.content.querySelector(`[name="${ strTable }:${ id }[]"]`);

                      if (component) {
                        component.setAttribute('id', component.getAttribute('id') + '-' + idParent);
                        this.setAttrToComponent(component, attr[id]);
                        this.setNewAttrIdAndName(component, index);
                      }
                    }

                    parent.appendChild(newComponent.content);

                    return false;
                  };
                });
              }
            }

            /*
             *   change name
             */

            const intArray = str.match(/\d+/g);
            if (!isOnlyClass) dom.setAttribute('name', `${ nameField }${ str }`);
            if (intArray) dom.setAttribute('id', `${ nameField }-${ (intArray) ? intArray.join('') : '' }`);
          }
          callback(dom, attr[name]);
        } else if (component && !isDefault && Object.prototype.toString.call(attr[name]) === '[object Array]') {

          for (let value of attr[name]) {

            let domArray;
            if (component.hasAttribute(Variables.paramsForm)) {
              domArray = document.querySelector(`[name="${ nameField }[]"][${ Variables.paramsJSONIdData }="${ value.id }"][${ Variables.paramsFormChildren }="${ component.getAttribute('id') }"]`);
            } else {
              domArray = component.querySelector(`[name="${ nameField }[]"][${ Variables.paramsJSONIdData }="${ value.id }"]`);
            }

            if (domArray) {
              callback(domArray, value);
            }
          }

        }
      }

    }
  }


  /*
  *   get dom
  */

  static _getDom(component, name, strTable, str = '') {
    let dom;

    const nameField = (strTable.length !== 0) ? `${ strTable }:${ name }${ str }` : `${ name }${ str }`;

    if (component && component.hasAttribute(Variables.paramsForm)) {
      dom = document.querySelectorAll(`[name="${ nameField }"][${ Variables.paramsFormChildren }="${ component.getAttribute('id') }"]`);
    } else if (component) {
      dom = component.querySelectorAll(`[name="${ nameField }"]`);
    }
    return [dom, nameField];
  }

}