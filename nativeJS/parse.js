
let isRequestAPIGET = false;

export class Parse {

  constructor(page) {
    this.page = page;
    this.mainContent = page.getElementsByTagName(Variables.nameMainContent)[0];

    this._parsComponents(this.page);

    Observer.addListener(Variables.reChangeDomDynamically, (component) => this._parsComponents(component));

    if (!isRequestAPIGET) {
      Observer.emit(Variables.documentIsReady, this.page);
    }

  }


  _changeComponentDom(component, jsonID = "0") {
    this.mainContent.innerHTML = component;
    this._parsComponents(this.mainContent, jsonID);
    if (!isRequestAPIGET) {
      Observer.emit(Variables.documentIsReady, this.mainContent);
    }
  };


  _parsComponents(componentDom, jsonID = "0") {

    // if tag have a link to router
    componentDom.querySelectorAll(`[${ Variables.routerAttr }]`).forEach((component) => {
      const self = this;
      component.onclick = function () {
        const urlAPIGET = this.getAttribute(Variables.routerHref);
        jsonID = this.getAttribute(Variables.paramsJSONId);
        Router.routing(urlAPIGET);
        Native.request(urlAPIGET);
        Observer.addListener(Variables.responseToRequest, (component, url) => {
          if (urlAPIGET === url) {
            self._changeComponentDom(component, jsonID);
          }
        });
        return false;
      };
    });


    // if tag have a link to API GET request
    componentDom.querySelectorAll(`[${ Variables.routerAPIGET }]`).forEach((component) => {

      isRequestAPIGET = true;

      const urlAPIGET = component.getAttribute(Variables.routerAPIGET);

      Native.request(urlAPIGET);
      component.removeAttribute(Variables.routerAPIGET);

      Observer.addListener(Variables.responseToRequest, (response, url) => {

        //TODO: need try catch response => JSON.parse()
        if (urlAPIGET === url) {
          if(response) {
            try {
              Native.setValueDataByAttr(JSON.parse(response));
            } catch(e) {
              alert(e); // error in the above string (in this case, yes)!
            }
          }
          Observer.emit(Variables.documentIsReady, component);
          return true;
        }
      });

    });

    // if tag have a link to API POST request
    componentDom.querySelectorAll(`[${ Variables.routerAPIPOST }]`).forEach((component) => {

      isRequestAPIGET = true;

      // TODO: refactoring to POST request
      const urlAPIPOST = component.getAttribute(Variables.routerAPIPOST) + '?id=' + jsonID;

      Native.request(urlAPIPOST);
      component.removeAttribute(Variables.routerAPIPOST);

      Observer.addListener(Variables.responseToRequest, (response, url) => {

        //TODO: need try catch response => JSON.parse()
        if (urlAPIPOST === url) {
          debugger;
          if(response) {
            try {
              Native.setValueDataByAttr(JSON.parse(response));
            } catch(e) {
              alert(e); // error in the above string (in this case, yes)!
            }
          }
          Observer.emit(Variables.documentIsReady, component);
          return true;
        }
      });
    });

  }


}