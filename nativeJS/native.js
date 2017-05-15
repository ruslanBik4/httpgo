//
// native.js
//

export class Native {

  static getHTMLDom(component, data, remove = false, parent) {
    let temp = document.createElement('template');
    if (temp.content && this.isElement(component)) {
      temp.innerHTML = eval('`' + component.innerHTML + '`');
      if (this.isElement(parent)) {
        parent.appendChild(temp.content);
      } else {
        component.parentElement.appendChild(temp.content);
      }
      if (remove) {
        component.parentNode.removeChild(component);
      }
    } else {
      console.log(`It's not dom component: ${ component }`);
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


  /*
  *   Set Value Data By Attribute to Dom
  */

  static setValueDataByAttr(data = {}) {

    let obj = data['fields'];
    ParseJSON.parseDataGet(obj, ParseJSON.setAttrToComponent.bind(ParseJSON));

    obj = data['form'];
    const element = document.getElementById(obj['id']);

    if (this.isElement(element)) {
      for (let key in obj) {
        element.setAttribute(key, obj[key]);
      }
    }

    obj = data['data'];
    for (let key in obj) {
      ParseJSON.parseDataGet(obj[key], ParseJSON.insertValueToComponent.bind(ParseJSON), '', true);
    }

  }


  /*
   *  get Value Data By Attributes from Dom
   */

  static getValueDataByAttributes(dom, attr = '', data = {}) {
    const elements = dom.getAttribute(attr);
    for (let element of elements) {
      for (let key in data) {
        element.setAttribute(key, data[key]);
      }
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
  *   Find first ancestor by class
  */

  static findAncestorByClass(element, className) {
    if (this.isElement(element) && typeof className === 'string') {
      while (!element.classList.contains(className) && (element = element.parentElement));
    }
    return element;
  }



}
