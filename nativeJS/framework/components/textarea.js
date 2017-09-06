
export class TextArea {

  /*
  *   Getters
  */
  static get className() {
    return 'c-app-textarea';
  }


  /*
   *   set default value
   */
  static setDefaultAttr(component, attr) {

    component = Native.findAncestorByClass(component, this.className);

    if (component) {
      ParseJSON.insertDataToAttrSetText(component, attr);
    } else {
      throw new SyntaxError(`Данные некорректны, поле textarea`);
    }

  }


  /*
   *    set value
   */
  static addAttrToComponent(component, value) {
    component.textContent = value;
  }

}