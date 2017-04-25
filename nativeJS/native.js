//
// native.js
//

export class Native {


  //
  // getHTML() {
  //   let temp = document.createElement('template');
  //   if (temp.content) {
  //     // for (let i = 0; i < this.dom.length - 2; i++) {
  //     //   if (this.dom[i] === '#') {
  //     //     let buf = this.dom[i];
  //     //     for (let j = i + 1; j < this.dom.length - 2; j++) {
  //     //       buf += this.dom[j];
  //
  //     //       //call func for lexer/syntax analyze
  //
  //     //       if (utilities[buf])
  //     //         debugger;
  //     //     }
  //     //   }
  //     // }
  //
  //     temp.innerHTML = eval('`' + this.dom + '`');
  //
  //     /*
  //      * Call func after parse Dom
  //     */
  //
  //     if (this.functions) {
  //       this.functions.map( (func) => {
  //         func.call(this, temp.content);
  //       });
  //     }
  //
  //   }
  //   return temp.content;
  // }

  /*
   *  Call name func after parse Dom
  */

  // callFuncAfterParseDom(...functions) {
  //   this.functions = functions;
  // }


  /*
   *  Func for render HTML
  */

  _reChangedDomHTML() {
    Observer.emit(Variables.responseToRequest, this);
  }


  /*
   * set Value Data By Attribute to Dom
  */

  static _setValueDataByAttr(dom, data = {}) {
    debugger;
    for (let key in data) {

      // json have key fields
      if (key === Variables.paramsJSON.fields.name) {

        const fieldsConst = Variables.paramsJSON.fields;

        // unique field in component
        for (let fieldsID in data[key]) {

          const uniqueField = data[key][fieldsID];
          const element = dom.getAttribute(fieldsID);
          debugger;
          // all params in unique field
          for (let param in uniqueField) {

            switch (uniqueField[param]) {
              case fieldsConst.text:
                element.innerHTML = fields[fieldsID];
                break;

              default:
                console.log(`Don't key in Native: ${ fieldsID }`);
                break;
            }

          }

        }
      }

      element.setAttribute(key, data[key]);
    }
  }



  static setValueDataByAttr(data = {}) {

    // if (data[Variables.paramsJSONId] === 'test') {
    //   const element = document.getElementById(data[Variables.paramsJSONId]);
    //   if (element) {
    //
    //     this._setValueAttrByComponent(element, data);
    //   }
    // } else {

      for (let key in data) {
        const element = document.getElementById(key);

        if (element && !element.hasAttribute(Variables.paramsJSONAdded)) {
          this._setValueAttrByComponent(element, data[key]);
          break;
        } else if (typeof data[key] === 'object') {
          this.setValueDataByAttr(data[key]);
        }

      }
    // }

  }

  static _setValueAttrByComponent(element, params) {
    for (let attr in params) {
      if (attr !== Variables.paramsJSONChildren) {
        element.setAttribute(attr, params[attr]);
      }
    }
    if (params[Variables.paramsJSONChildren]) {
      this._addChildrenToComponent(element, params[Variables.paramsJSONChildren]);
    }
    element.setAttribute(Variables.paramsJSONAdded, '');
  }

  static _addChildrenToComponent(component, params) {
    let template = document.createElement('template');

    if (component.tagName === 'SELECT') {
      const child = component.firstElementChild;
      const tagName = child.tagName;
      const attributes = child.attributes;
      for (let key in params) {
        const tag = document.createElement(tagName);
        const text = params[key];
        for (let attr of attributes) {
          tag.setAttribute(attr.name, text);
          tag.text = text;
        }
        template.content.appendChild(tag);
      }
      component.innerHTML = template.innerHTML;
    } else if (component.tagName === 'INPUT') {

    }
  }




  /*
   * get Value Data By Attribute to Dom
  */

  static getValueDataByAttr(dom, attr = '', data = {}) {
    const elements = dom.getAttribute(attr);
    for (let element of elements) {
      for (let key in data) {
        element.setAttribute(key, data[key]);
      }
    }
  }




  /* 
   *  get and post request with callback
  */

  static request(url, method = 'GET', params = {}) {

    const XHR = ('onload' in new XMLHttpRequest()) ? XMLHttpRequest : XDomainRequest;
    const xhr = new XHR();
    xhr.open(method, url, true);

    if (method === 'GET') {
      xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
    } else if (method === 'POST') {
      xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded')
    }

    xhr.send(params);

    this.requestOn = true;

    xhr.onload = (response) => {
      Observer.emit(Variables.responseToRequest, response.currentTarget.responseText);
    };

    xhr.onerror = function () {
      console.log(`Error API to url ${ url } : ${ this }`);
    };

  }

}
