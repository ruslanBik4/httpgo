
const classNames = {
  'radio'     : 'c-app-radio',
  'checkbox'  : 'c-app-checkbox'
};

export class Input {

  /*
   *   create inputs
   */

  static createList(component, list) {

    const typeComponent = component.getAttribute('type');
    const className = classNames[typeComponent];

    if (className) {
      component = Native.findAncestor(component, className);

      if (component) {

        for (let item in list) {
          if (typeof list[item] === 'string') {
            debugger;
            component = this._appendDomToComponent(component, component, list[item]);
          }
        }

      } else {
        console.log(`Not found component with className: ${ className }`);
      }
    }
  }


  /*
   *   selected active item
   */

  static selectedItem(component, ) {

  }



  static _appendDomToComponent(component, parent, textContent = '') {
    if (component.children) {
      for (let child of component.children) {
        this._recursionAppendDomToComponent(child, parent, textContent);
      }
    }
    if (component.hasAttribute(Variables.paramsJSONSetText)) {
      component.textContent = textContent;
    }
    parent.appendChild(component);
    return parent;
  }

}