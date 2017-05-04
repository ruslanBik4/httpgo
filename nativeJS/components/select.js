
const className = 'c-app-select';

export class Select {

  /*
  *   create <option></option> list
  */

  static createList(component, list) {

    if (component) {
      for (let key in list) {

        let option = document.createElement('option');

        option.setAttribute(Variables.paramsJSONForPost, key);
        option.text = text;

        component.appendChild(option);

      }
    } else {
      console.log(`Not found component with className: ${ className }`);
    }

  }

  /*
  *   selected active item
  */

  static selectedItem(component, key) {

    for (let option of component.children) {
      if (option.getAttribute(Variables.paramsJSONForPost) === key) {
        option.setAttribute('selected', '');
        break;
      }
    }

  }

}