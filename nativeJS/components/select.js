
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
      console.log(`Not found component with className: ${ this.className }`);
    }

  }

  /*
  *   selected active item
  */

  static addAttrToComponent(component, attr) {

    for (let option of component.children) {
      if (option.getAttribute(Variables.paramsJSONForPost) === attr) {
        option.setAttribute('selected', '');
        break;
      }
    }

  }

}