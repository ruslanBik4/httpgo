//
// variables.js
//

const nameApp = 'app';
const routerAPIGET = 'data-api-get';
const routerAPIPOST = 'data-api-post';
const routerHref = 'href';
const routerAttr = 'data-link';
const nameMainContent = 'app-main';
const paramsJSONAttr = 'attr';


// JSON const
const paramsJSONId = 'id';
const paramsJSONChildren = 'children';
const paramsJSONAdded = 'data-api-added';


// Observer
const responseToRequest = 'responseToRequest';
const documentIsReady = 'documentIsReady';

export class Variables {

  static get nameApp() {
    return nameApp;
  }

  static get routerAPIGET() {
    return routerAPIGET;
  }

  static get routerAPIPOST() {
    return routerAPIPOST;
  }

  static get routerHref() {
    return routerHref;
  }

  static get routerAttr() {
    return routerAttr;
  }

  static get nameMainContent() {
    return nameMainContent;
  }

  static get paramsJSONId() {
    return paramsJSONId;
  }

  static get paramsJSONChildren() {
    return paramsJSONChildren;
  }

  static get paramsJSONAdded() {
    return paramsJSONAdded;
  }


  static get responseToRequest() {
    return responseToRequest;
  }

  static get documentIsReady() {
    return documentIsReady;
  }
  
}