
let isRequestAPI = 0;
let scriptIncludes = [];
let idCurrentPage;
let stateHistoryComponents = [];
let firstComponent;

let dataAfterForm;
let customHadlerAfterForm;

export class Parse {

  /*
  *   Getters
  */
  static get idCurrentPage() { return idCurrentPage; }
  static set idCurrentPage(id) { idCurrentPage = id; }
  static get getDataAfterForm() { return dataAfterForm; }
  static customHadlerAfterForm(func) { customHadlerAfterForm = func; }
  static setStateHistoryComponents() { stateHistoryComponents.push(this.mainContent.innerHTML); }


  /*
  *   Set main content to page
  */
  static setMainContent(index = 0) {
    if (index < stateHistoryComponents.length) {
      this.mainContent.innerHTML = stateHistoryComponents[index];
    } else {
      console.log(`don't find component in stateHistory`);
    }
  }

  /*
  *   Set component to dynamic component (main)
  */
  static setComponent(component, isFirst = false) {
    if (isFirst) {
      this.mainContent.innerHTML = firstComponent;
    } else {
      this.mainContent.innerHTML = component;
    }
  }


  /*
  *   Init Parse framework
  */
  static start(page) {
    this.page = page;
    this.mainContent = page.getElementsByTagName(Variables.nameMainContent)[0];
    firstComponent = this.mainContent.innerHTML;
    this.parsComponents(this.page);
    this._documentIsReady(this.page);
  }


  /*
  *   parsing component
  */
  static parsComponents(componentDom) {

    // if tag have a link to router
    this._routerLink(componentDom);

    // for form TODO: need refactoring
    componentDom.querySelectorAll(`${ Variables.nameForm }`).forEach((component) => {
      const self = this;
      component.onsubmit = function() {

        saveForm(this, (data, form) => {
          let result = true;
          dataAfterForm = data;

          if (customHadlerAfterForm) {
            result = customHadlerAfterForm(data, form);
          }

          if (result) {
            const url = form.getAttribute(Variables.routerHref);

            if (url) {
              Native.request(url, (component) => {
                self._changeComponentDom(component);
                Router.routing(url);
              });
            }
          }

        }, () => { alert('Произошла ошибка, повторите попытку') });

        return false;
      };
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
  *   get component by route
  */
  static getComponentByRoute(url) {

    if (url.startsWith('/')) {
      Native.request(url, (component) => {
        this._changeComponentDom(component);
        Router.routing(url);
      });
    }
  }


  /*
  *    change Component dynamically
  */
  static _changeComponentDom(component) {
    isRequestAPI = 0;
    this.mainContent.innerHTML = component;
    // stateHistoryComponents.push(this.mainContent.innerHTML);
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

        if (this.hasAttribute(Variables.paramsJSONId)) {
          idCurrentPage = this.getAttribute(Variables.paramsJSONId);
        }

        self.getComponentByRoute(this.getAttribute(Variables.routerHref));

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


  /*
  *   Is document Ready?
  */
  static _documentIsReady(component) {
    if (isRequestAPI === 0) {
      Observer.emit(Variables.documentIsReady, component);
      this._importScript(component);
    }
  }
}