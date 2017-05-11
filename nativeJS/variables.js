//
// variables.js
//


// common
const routerAPIGET = 'data-api-get';
const routerAPIPOST = 'data-api-post';
const routerHref = 'href';
const routerAttr = 'data-link';
const nameMainContent = 'app-main';
const dynamicallyScript = 'app-script';


// JSON
const paramsJSONId = 'data-api-post-id';
const paramsJSONSetText = 'data-set-text';
const paramsJSONTable = 'tableid_';
const paramsJSONForPost = 'value';


// JSON reserved words
const paramsJSONList = 'list';
const paramsJSONTitle = 'title';
const paramsJSONSet = 'set';
const paramsJSONEnum = 'enum';
const paramsJSONType = 'type';


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

  static get dynamicallyScript() {
    return dynamicallyScript;
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

  static get paramsJSONTable() {
    return paramsJSONTable;
  }

  static get paramsJSONForPost() {
    return paramsJSONForPost;
  }


  /*
   *  JSON reserved words
  */

  static get paramsJSONList() {
    return paramsJSONList;
  }

  static get paramsJSONTitle() {
    return paramsJSONTitle;
  }

  static get paramsJSONSet() {
    return paramsJSONSet;
  }

  static get paramsJSONEnum() {
    return paramsJSONEnum;
  }

  static get paramsJSONType() {
    return paramsJSONType;

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