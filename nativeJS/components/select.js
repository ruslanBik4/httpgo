
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

      const isArray = (Object.prototype.toString.call(list) === '[object Array]');

      for (let key in list) {

        let option = document.createElement('option');

        option.setAttribute(Variables.paramsJSONForPost, (isArray) ? list[key] : key );
        option.textContent = list[key];

        component.appendChild(option);

      }
      // debugger

      // component.selectedIndex = 0;
    } else {
      throw new SyntaxError(`Данные некорректны, поле select`);
    }

  }

  /*
  *   selected active item
  */

  static addAttrToComponent(component, attr) {

    if (component.children.length !== 0) {

      if (component.selectedIndex > -1) {
        const index = component.selectedIndex;
        const option = component.children[index];
        component.selectedIndex = -1;
        option.selected = false;
        option.removeAttribute('selected');
      }

      for (let i = 0; i < component.children.length; i++) {
        let option = component.children[i];
        if (option.getAttribute(Variables.paramsJSONForPost) == attr || option.text == attr) {
          component.selectedIndex = i;
          option.selected = true;
          option.setAttribute('selected', true);
          break;
        }
      }
    }

  }

}