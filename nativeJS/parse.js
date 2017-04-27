
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


  _changeComponentDom(component) {
    this.mainContent.innerHTML = component;
    this._parsComponents(this.mainContent);
    if (!isRequestAPIGET) {
      Observer.emit(Variables.documentIsReady, this.mainContent);
    }
  };


  _parsComponents(componentDom) {

    // if tag have a link to router
    componentDom.querySelectorAll(`[${ Variables.routerAttr }]`).forEach((component) => {
      const self = this;
      component.onclick = function () {
        const urlAPIGET = this.getAttribute(Variables.routerHref);
        Router.routing(urlAPIGET);
        Native.request(urlAPIGET);
        Observer.addListener(Variables.responseToRequest, (component, url) => {
          if (urlAPIGET === url) {
            self._changeComponentDom(component);
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
      component.onclick = function () {
        Native.request(component.getAttribute(Variables.routerAPIGET), Native.getValueDataByAttr());
        return false;
      };
    });

  }


}