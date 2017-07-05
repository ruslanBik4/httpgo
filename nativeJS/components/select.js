
export class Select {

  static get className() {
    return 'c-app-select';
  }

  /*
  *   create <option></option> list
  */

  static createList(component, list) {

    component = Native.findAncestorByClass(component, this.className);

    if (component) {

      for (let key in list) {

        let option = document.createElement('option');

        option.setAttribute(Variables.paramsJSONForPost, key);
        option.textContent = list[key];

        component.appendChild(option);

      }
    } else {
      throw new SyntaxError(`Данные некорректны, поле select`);
    }

  }

  /*
  *   selected active item
  */

  static addAttrToComponent(component, attr) {

    if (component.children.length !== 0) {
      component.selectedIndex = -1;
      for (let i = 0; i < component.children.length; i++) {
        let option = component.children[i];
        if (option.getAttribute(Variables.paramsJSONForPost) == attr) {
          option.setAttribute('selected', '');
          break;
        } else if (option.text === attr) {
          option.setAttribute('selected', '');
          break;
        }
      }
    }

  }

}