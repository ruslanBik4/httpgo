//
// variables.js
//


// common
const routerAPIGET = 'data-api-get';
const routerAPIPOST = 'data-api-post';
const routerHref = 'href';
const routerAttr = 'data-link';
const routerStopAttr = 'data-link';
const nameForm = 'form';
const nameMainContent = 'app-main';
const dynamicallyScript = 'app-script';


// JSON
const paramsJSONId = 'data-api-post-id';
const paramsJSONSetText = 'data-set-text';
const paramsJSONTable = 'tableid_';
const paramsJSONPhotos = 'photoid_photos';
const paramsJSONForPost = 'value';
const paramsJSONIdForTable = 'native-table-id';
const paramsJSONIdData = 'data-id';
const paramsForm = 'data-form-id';
const paramsFormChildren = 'form';
const paramsForClick = 'native-click-button';
const paramsChangeId = 'native-change-id';
const paramsNotInsertData = 'data-not-insert';


// JSON reserved words
const paramsJSONList = 'list';
const paramsJSONTitle = 'title';
const paramsJSONSet = 'set';
const paramsJSONEnum = 'enum';
const paramsJSONType = 'type';
const paramsJSONDefault = 'default';


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

  static get nameForm() {
    return nameForm;
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

  static get paramsJSONPhotos() {
    return paramsJSONPhotos;
  }

  static get paramsJSONForPost() {
    return paramsJSONForPost;
  }

  static get paramsJSONIdForTable() {
    return paramsJSONIdForTable;
  }

  static get paramsJSONIdData() {
    return paramsJSONIdData;
  }

  static get paramsForm() {
    return paramsForm;
  }

  static get paramsFormChildren() {
    return paramsFormChildren;
  }

  static get paramsForClick() {
    return paramsForClick;
  }

  static get paramsChangeId() {
    return paramsChangeId;
  }

  static get paramsNotInsertData() {
    return paramsNotInsertData;
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

  static get paramsJSONDefault() {
    return paramsJSONDefault;
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