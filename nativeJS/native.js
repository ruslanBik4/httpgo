//
// native.js
//

let bufData;

const codeStatusServer = {
  '401' : {
    url: '/customer/login-user/'
  },
  '403' : {

  }

};


export class Native {

  /*
  *   get HTML bu template string
  */
  static getHTMLDom(component, data, parent, isRemove = false) {
    console.log(document)
    let temp = document.createElement('template');
    let result;

    if (temp.content && this.isElement(component)) {

      try {
        temp.innerHTML = eval('`' + component.innerHTML + '`');
      } catch (e) {
        console.log('%c error in Native.getHTMLDom()! ', 'color: #F44336');
        console.error(component, '\n', data, '\n', e);
      }

      if (this.isElement(parent)) {
        parent.appendChild(temp.content);
        result = parent.lastElementChild;
      } else {
        component.parentElement.appendChild(temp.content);
        result = component.parentElement.lastElementChild;
      }

      if (isRemove) {
        component.parentNode.removeChild(component);
      }

    } else {
      console.error(`It's not dom component: ${ component }`);
    }
    return result;
  }


  /*
  *   go to new component
  */
  static goToLink(url) {
    if (typeof url === 'string') {
      Parse.getComponentByRoute(url);
    } else {
      console.error('Url is don`t string: ', url);
    }
  }


  /*
  *   add parse to Dynamic Component
  */
  static reChangeDomDynamically(component) {
    if (this.isElement(component)) {
      Parse.parsComponents(component);
    } else {
      console.error(`This component is not dom: `, component);
    }
  }


  /*
  *   current Id dynamically page
  */
  static get getIdCurrentPage() {
    return Parse.idCurrentPage;
  }


  /*
  *   reset current id dynamically page
  */
  static resetIdCurrentPage() {
    Parse.idCurrentPage = null;
  }


  /*
  *   parse JSON is safely
  */
  static jsonParse(response) {
    try {
      return JSON.parse(response);
    } catch(e) {
      console.error(e, response);
      // alert(e); // error in the above string (in this case, yes)!
    }
  }


  /*
   *  get and post request with callback
   */
  static request(url, callback, data) {

    let method = 'GET';
    let body = ['\r\n'];

    const XHR = ('onload' in new XMLHttpRequest()) ? XMLHttpRequest : XDomainRequest;
    const xhr = new XHR();

    if (data) {
      method = 'POST';
    }

    xhr.open(method, url, true);

    if (data) {
      let boundary = String(Math.random()).slice(2);
      let boundaryMiddle = '--' + boundary + '\r\n';
      let boundaryLast = '--' + boundary + '--\r\n';

      for (let key in data) {
        body.push('Content-Disposition: form-data; name="' + key + '"\r\n\r\n' + data[key] + '\r\n');
      }

      body = body.join(boundaryMiddle) + boundaryLast;
      xhr.setRequestHeader('Content-Type', 'multipart/form-data; boundary=' + boundary);

    }

    xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
    xhr.send(body);

    xhr.onload = (response) => {

      const codeStatus = codeStatusServer[response.currentTarget.status];

      if (codeStatus) {
        Observer.emit(`Server: ${ response.currentTarget.status }`, response, url, callback, data);
        return;
      }


      if (callback) {
        callback(response.currentTarget.responseText, url);
      } else {
        Observer.emit(Variables.responseToRequest, response.currentTarget.responseText, url);
      }
    };

    xhr.onerror = function () {
      console.error(`Error API to url ${ url } : ${ this }`);
    };

  }


  /*
  *   Set Value Data By Attribute to Dom
  */
  static setValueDataByAttr(data = {}) {

    ParseJSON.parseDataGet(data['fields'], ParseJSON.setAttrToComponent.bind(ParseJSON));

    let obj = data['form'];
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
   *   set default data for Fields
   */
  static setDefaultFields(component, fields, str = '', isOnlyClass = false) {
    if (this.isElement(component) && fields) {
      ParseJSON.setValue(component, fields, ParseJSON.setAttrToComponent.bind(ParseJSON), (typeof str === 'string') ? str : str.toString(), true, isOnlyClass);
    }
  }


  /*
  *     set form attributes
  */
  static setForm(componentName, attr) {
    const component = document.getElementById(componentName);
    if (component) {
      delete attr.id;
      for (let key in attr) {
        component.setAttribute(key, attr[key]);
      }
    }
  }


  /*
   *   insert data for data
   */
  static insertData(component, data, str = '', isOnlyClass = false) {
    if (this.isElement(component) && data) {
      ParseJSON.setValue(component, data, ParseJSON.insertValueToComponent.bind(ParseJSON), (typeof str === 'string') ? str : str.toString(), false, isOnlyClass);
    }
  }
  
  
  /*
  *     get data after submit form 
  */
  static getDataAfterForm() {
    return Parse.getDataAfterForm;
  }


  /*
  *     custom handler
  */
  static customHadlerAfterForm(func) {
    Parse.customHadlerAfterForm(func);
  }


  /*
  *     buf variables
  */
  static bufVariables(data) {
    if (data) {
      bufData = data;
    } else {
      return bufData;
    }
  }



}
