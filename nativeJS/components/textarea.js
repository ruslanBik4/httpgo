
export class TextArea {

  static get className() {
    return 'c-app-textarea';
  }


  /*
   *   set default value
   */

  static setDefaultAttr(component, attr) {

    component = Native.findAncestorByClass(component, this.className);
    ParseJSON.insertDataToAttrSetText(component, attr);

  }


  /*
   *    set value
   */

  static addAttrToComponent(component, value) {
    component.textContent = value;
  }

}