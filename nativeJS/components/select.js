
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

    if (component && component.children) {
      for (let option in component.children) {
        if (component.children[option].getAttribute(Variables.paramsJSONForPost) === attr) {
          component.children[option].setAttribute('selected', '');
          break;
        }
      }
    }

  }

}