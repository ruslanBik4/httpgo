//
// native.js
//

export class Native {


  /*
  *   Set Value Data By Attribute to Dom
  */

  static setValueDataByAttr(data = {}) {

    // TODO: refactoring hardcode

    let obj = data['fields'];
    this._insertDataToDom(obj, this._setValueAttrByComponent);

    obj = data['form'];
    const element = document.getElementById(obj['id']);
    if (this.isElement(element)) {
      for (let key in obj) {
        element.setAttribute(key, obj[key]);
      }
    }

    obj = data['data'];
    for (let key in obj) {
      this._insertDataToDom(obj[key], this._insertAttributeToDom);
    }

  }


  static _insertAttributeToDom(element, attr = '') {

    if (this.isElement(element) && attr != null) {
      switch (element.tagName) {

        // input

        case 'INPUT':
          switch (element.getAttribute('type')) {
            case 'text':
            case 'number':
              element.setAttribute('value', attr);
              break;
            case 'checkbox':
              if (attr !== "0") {
                element.setAttribute('checked', 'true');
              }
              break;
            default:
              break;
          }
          break;


        // select

        case 'SELECT':
          let number = parseInt(attr) - 1;
          if (0 <= number && number < element.children.length) {
            element.children[number].setAttribute('selected', '');
          }
          break;


        // textarea

        case 'TEXTAREA':
          element.value = attr;
          break;

        default:
          break;
      }
    }
  }


  /*
   *   When need recursion for table
   */

  static _insertDataToDom(data, callback) {

    for (let id in data) {

      const dom = document.getElementById(id);

      if (dom) {

        // has prefix "tableid_" for recursion
        if (data[id].startsWith(Variables.paramsJSONTable)) {
          this._insertDataToDom(data[id], callback);
        } else {
          callback(dom, data[id]);
        }

      }

    }
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

  }


  /*
  *   func for add Children to Component
  */

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
        case 'time':
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
   *  returns true if it is a DOM element
  */
  static isElement(obj) {
  return (
    typeof HTMLElement === "object" ? obj instanceof HTMLElement : //DOM2
      obj && typeof obj === "object" && obj !== null && obj.nodeType === 1 && typeof obj.nodeName==="string"
  );
}


  /*
   *  get Value Data By Attribute to Dom
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
   *  When type="table" => recursion get id (name)
  */





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
