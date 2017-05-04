//
// variables.js
//


// common
const routerAPIGET = 'data-api-get';
const routerAPIPOST = 'data-api-post';
const routerHref = 'href';
const routerAttr = 'data-link';
const nameMainContent = 'app-main';


// JSON
const paramsJSONId = 'data-api-post-id';
const paramsJSONSetText = 'data-set-text';
const paramsJSONList = 'list';
const paramsJSONTable = 'tableid_';
const paramsJSONUnknown = 'set';
const paramsJSONForPost = 'value';


// Observer
const responseToRequest = 'responseToRequest';
const documentIsReady = 'documentIsReady';
const reChangeDomDynamically = 'reChangeDomDynamically';

export class Variables {

  /*
   *  Common
  */

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


  /*
   * JSON
  */

  static get paramsJSONId() {
    return paramsJSONId;
  }

  static get paramsJSONSetText() {
    return paramsJSONSetText;
  }

  static get paramsJSONList() {
    return paramsJSONList;
  }

  static get paramsJSONTable() {
    return paramsJSONTable;
  }

  static get paramsJSONUnknown() {
    return paramsJSONUnknown;
  }

  static get paramsJSONForPost() {
    return paramsJSONForPost;
  }


  /*
   * Observer
  */

  static get responseToRequest() {
    return responseToRequest;
  }

  static get documentIsReady() {
    return documentIsReady;
  }

  static get reChangeDomDynamically() {
    return reChangeDomDynamically;
  }
  
}