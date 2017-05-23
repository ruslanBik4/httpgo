
let isRequestAPI = 0;
let scriptIncludes = [];
let idCurrentPage;
let stateHistoryComponents = [];
let firstComponent;

export class Parse {

  static get idCurrentPage() {
    return idCurrentPage;
  }

  static setMainContent(index = 0) {
    if (index < stateHistoryComponents.length) {
      this.mainContent.innerHTML = stateHistoryComponents[index];
    } else {
      console.log(`don't find component in stateHistory`);
    }
  }

  static setComponent(component, isFirst = false) {
    if (isFirst) {
      this.mainContent.innerHTML = firstComponent;
    } else {
      this.mainContent.innerHTML = component;
    }
  }

  static start(page) {
    this.page = page;
    this.mainContent = page.getElementsByTagName(Variables.nameMainContent)[0];
    firstComponent = this.mainContent.innerHTML;
    this.parsComponents(this.page);
    this._documentIsReady(this.page);
  }


  static parsComponents(componentDom) {

    // if tag have a link to router
    this._routerLink(componentDom);

    // for form TODO: need refactoring
    componentDom.querySelectorAll(`${ Variables.nameForm }`).forEach((component) => {
      const elements = component.querySelectorAll(`button, input[type=button]`);
      for (let element of elements) {
        element.onclick = function () {
          saveForm(component, () => { alert('Ваша форма сохранена') }, () => { alert('Произошла ошибка, повторите попытку') });
        };
      }
    });

    // if tag have a app-script
    componentDom.querySelectorAll(Variables.dynamicallyScript).forEach((scriptComponent) => {
      scriptIncludes.push(scriptComponent.getAttribute('src'));
    });


    // if tag have a link to API GET request
    this._APIGetRequest(componentDom);

    // if tag have a link to API POST request
    this._APIPostRequest(componentDom);

  }

  /*
  *    change Component dynamically
  */

  static _changeComponentDom(component) {
    isRequestAPI = 0;
    this.mainContent.innerHTML = component;
    stateHistoryComponents.push(this.mainContent.innerHTML);
    this.parsComponents(this.mainContent);
    this._documentIsReady(this.mainContent);
  };


  /*
  *   parse Router link for dynamically component
  */

  static _routerLink(componentDom) {
    componentDom.querySelectorAll(`[${ Variables.routerAttr }]`).forEach((component) => {
      const self = this;
      component.onclick = function () {
        idCurrentPage = this.getAttribute(Variables.paramsJSONId);
        let url = this.getAttribute(Variables.routerHref);

        Native.request(url, (component) => {
          self._changeComponentDom(component);
          Router.routing(url);
        });

        return false;
      };
    });
  }


  /*
   *   parse API get request by tag 'data-api-get'
   */

  static _APIGetRequest(componentDom) {
    componentDom.querySelectorAll(`[${ Variables.routerAPIGET }]`).forEach((component) => {

      ++isRequestAPI;

      let src = component.getAttribute(Variables.routerAPIGET);

      if (idCurrentPage) {
        src += '?id=' + idCurrentPage;
      }

      Native.request(src, (response) => {
        Native.setValueDataByAttr(Native.jsonParse(response));
        --isRequestAPI;
        this._documentIsReady(component);
      });

      component.removeAttribute(Variables.routerAPIGET);

    });
  }



  /*
  *   parse API post request by tag 'data-api-post'
  */

  static _APIPostRequest(componentDom) {
    componentDom.querySelectorAll(`[${ Variables.routerAPIPOST }]`).forEach((component) => {

      ++isRequestAPI;

      let data = {};

      if (idCurrentPage) {
        data.id = idCurrentPage
      }

      Native.request(component.getAttribute(Variables.routerAPIPOST), (response) => {
        Native.setValueDataByAttr(Native.jsonParse(response));
        --isRequestAPI;
        this._documentIsReady(component);
      }, data);

      component.removeAttribute(Variables.routerAPIPOST);

    });
  }


  /*
  *   import script dynamically
  */

  static _importScript(component) {
    let scriptsComponent = [];

    component.querySelectorAll(Variables.dynamicallyScript).forEach((dom) => {
      scriptsComponent.push(dom.getAttribute('src'));
    });

    for (let script of scriptIncludes) {
      const normalized = System.normalizeSync(script);
      if (System.has(normalized) && scriptsComponent.includes(script)) {
        System.delete(normalized);
      }
      System.import(script);
    }
  }

  static _documentIsReady(component) {
    if (isRequestAPI === 0) {
      Observer.emit(Variables.documentIsReady, component);
      this._importScript(component);
    }
  }


}