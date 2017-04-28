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
   * set Value Data By Attribute to Dom
  */

  static setValueDataByAttr(data = {}) {

    // if (data[Variables.paramsJSONId] === 'test') {
    //   const element = document.getElementById(data[Variables.paramsJSONId]);
    //   if (element) {
    //
    //     this._setValueAttrByComponent(element, data);
    //   }
    // } else {

    //TODO: refactoring hardcode
      for (let key in data['fields']) {
        const element = document.getElementById(key);
        if (element)
          this._setValueAttrByComponent(element, data['fields'][key]);
      }

    const element = document.getElementById(data['form']['id']);
    if (element) {
      for (let key in data['form']) {
        element.setAttribute(key, data['form'][key]);
      }
    }

    for (let key in data['data']) {
      for (let id in data['data'][key]) {
        const element = document.getElementById(id);
        if (element && data['data'][key][id]) {
          switch (element.tagName) {
            case 'INPUT':
              if (element.getAttribute('type') === 'checkbox') {
                if (data['data'][key][id] !== "0") {
                  element.setAttribute('checked', 'true');
                }
              } else if (element.getAttribute('type') === 'text') {
                element.setAttribute('value', data['data'][key][id]);
              }
              break;
            case 'SELECT':
              if (parseInt(data['data'][key][id]) < element.children.length) {
                element.children[data['data'][key][id]].setAttribute('selected', '');
              }
            case 'TEXTAREA':
              element.setAttribute('value', data['data'][key][id]);
              break;
            default:
              break;
          }
        }
      }
    }
    // }

  }

  static _setValueAttrByComponent(element, params = {}) {
    for (let attr in params) {
      if (attr !== Variables.paramsJSONChildren) {
        element.setAttribute(attr, params[attr]);
      }
    }
    if (params[Variables.paramsJSONChildren]) {
      this._addChildrenToComponent(element, params[Variables.paramsJSONChildren]);
    } else {
      this._addChildrenToComponent(element, params);
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

      let title = component.getAttribute('title');

      switch (component.getAttribute('type')) {
        case 'search':
        case 'text':
        case 'number':
        case 'checkbox':
          if (title) {
            component.parentElement.lastElementChild.textContent = title;
          }

          break;
        case 'radio':
          debugger;
          break;
        default:
      }

    } else if (component.tagName === 'TEXTAREA') {
      component.parentElement.lastElementChild.textContent = component.getAttribute('title');
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

    xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");

    xhr.send(params);

    this.requestOn = true;

    xhr.onload = (response) => {
      Observer.emit(Variables.responseToRequest, response.currentTarget.responseText, url);
    };

    xhr.onerror = function () {
      console.log(`Error API to url ${ url } : ${ this }`);
    };

  }

}
