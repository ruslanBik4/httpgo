//
// native.js
//

export class Native {


  /*
  *   Set Value Data By Attribute to Dom
  */

  static setValueDataByAttr(data = {}) {

    let obj = data['fields'];
    ParseJSON.parseData(obj, ParseJSON.setAttrToComponent);

    obj = data['form'];
    const element = document.getElementById(obj['id']);
    if (this.isElement(element)) {
      for (let key in obj) {
        element.setAttribute(key, obj[key]);
      }
    }

    obj = data['data'];
    for (let key in obj) {
      ParseJSON.parseData(obj[key], ParseJSON.insertValueToComponent);
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
