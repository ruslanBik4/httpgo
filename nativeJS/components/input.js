

export class Input {

  static get classNames() {
    return {
      'radio'     : 'c-app-radio',
      'checkbox'  : 'c-app-checkbox'
    };
  }

  static get defaultClassName() {
    return 'c-app-input';
  }

  static classNameByTag(tag) {
    return (this.classNames[tag] ? this.classNames[tag] : this.defaultClassName);
  }


  /*
   *   set default value
   */

  static setDefaultAttr(component, attr) {

    const className = this.classNameByTag(component.type);
    let result;

    if (className) {
      result = Native.findAncestorByClass(component, className);
    }

    if (result) {
      ParseJSON.insertDataToAttrSetText(result, attr);
    } else {
      throw new SyntaxError(`Данные некорректны, поле input`);
    }

  }


  /*
   *   create inputs
   */

  static createList(component, list, isSet = false, isReturnDomString = false) {

    this.isSet = isSet;

    const typeComponent = component.getAttribute('type');
    const className = this.classNameByTag(typeComponent);

    if (className) {

      const idComponent = component.id;
      component = Native.findAncestorByClass(component, className);

      if (component) {

        let template = document.createElement('template');

        for (let item in list) {
          if (typeof list[item] === 'string') {
            const newComponent = component.firstElementChild.cloneNode(true);
            this._appendDomToComponent(newComponent, template.content, item, list[item]);
            template.content.appendChild(newComponent);
          }
        }

        if (isReturnDomString) {
          return template.innerHTML;
        }

        component.id = idComponent;
        component.innerHTML = template.innerHTML;

      }
    }
  }


  /*
   *   selected active item for input and set value
   */

  static addAttrToComponent(component, value = '') {

    switch (component.getAttribute('type')) {

      // radio || checkbox

      case 'radio':
      case 'checkbox':
        component.checked = !(value === '0');
        break;


      // other types for input

      default:
        component.value = value;
        break;
    }

  }



  static _appendDomToComponent(component, parent, dataId, textContent = '') {

    if (component.children.length !== 0) {
      for (let i = 0; i < component.children.length; i++) {
        this._appendDomToComponent(component.children[i], parent, dataId, textContent);
      }
    }

    if (component.tagName === 'INPUT') {
      if (this.isSet) {
        component.name += `[]`;
        component.setAttribute(Variables.paramsJSONIdData, dataId);
      }
      component.id += `-${ parent.children.length }`;
    }
    else if (component.tagName === 'LABEL') {
      component.htmlFor += `-${ parent.children.length }`;
    }
    else if (component.hasAttribute(Variables.paramsJSONSetText)) {
      component.textContent = textContent;
    }

  }

}