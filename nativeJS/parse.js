
let isRequestAPIGET = false;

export class Parse {

  constructor(page) {
    this.page = page;
    this.mainContent = page.getElementsByTagName(Variables.nameMainContent)[0];

    //TODO: remove later
    Native.request('/extranet/test/');

    this.response = (response) => {this._showList(JSON.parse(response))};
    Observer.addListener(Variables.responseToRequest, this.response, true);

  }

  //TODO: remove later, this block for show list-hotels in asside
  _showList(data) {
    // for (let users of data) {
      let counterForBusiness = 0;
      let users = {
        business : {
          "id"		: 0,
          "name"	: "ЧАО «Президент-отель»",
        },
        hotels: data
      };
      // for (let itemBusiness of users['business']) {
        $('.c-app-aside .aside_right .business').append(
          '<h4 class="panel-title">' +
          '<a data-toggle="collapse" data-parent="#parent_collapse' + counterForBusiness + '" href="#parent_collapse' + counterForBusiness + '">' + users.business.name + '<b class="caret"></b></a>' +
          '</h4>'
        );
        $('.c-app-aside .aside_right .businessContent').append(
          `
            <div id="parent_collapse${ counterForBusiness }" class="panel-collapse collapse">
              <div class="panel-body no_padding ">
                <div class="panel-group" id="accordion_aside_${ counterForBusiness }"></div>
              </div>
            </div>
          `
        );

        let counterForHotels = 0;
        for (let hotel in users.hotels) {
          $(`#accordion_aside_${ counterForBusiness }`).append(
            `
              <div class="panel panel-default">
                <div class="panel-heading">
                  <h4 class="panel-title">
                    <a data-toggle="collapse" data-parent="#accordion_aside_${ counterForBusiness }" class="collapsed" id="${ users.hotels[hotel]['id'] }_" href="#collapse${ counterForHotels }"><img src="/images/aside_left_1.png">${ users.hotels[hotel]['title'] }</a>
                  </h4>
                </div>
                <div id="collapse${ counterForHotels }" class="panel-collapse collapse">
                  <div class="panel-body">
                    <ul class="aside_submenu">
                      <li><a href="/extranet/dashboard/" data-link><span class="travel-icon-1001"></span>Dashboard</a></li>
                      <li><a class="moderated" href="/extranet/objects/" data-link><span class="travel-icon-1379"></span>Настройка Объекта</a></li>
                      <li><a class="ready" href="/extranet/payment/" data-link><span class="travel-icon-087"></span>Ценообразование</a></li>
                      <li><a href="/extranet/booking/" data-link><span class="travel-icon-007"></span>Бронирование</a></li>
                      <li><a href="/extranet/documents/" data-link><span class="travel-icon-443"></span>Документы</a></li>
                      <li><a href="/extranet/analazy/" data-link><span class="travel-icon-1364"></span>Аналитика</a></li>
                      <li><a href="/extranet/users/" data-link><span class="travel-icon-1027"></span>Сотрудники</a></li>
                    </ul>
                  </div>
                </div>
              </div>
          `
          );
          counterForHotels++;
        }
        counterForBusiness++;
      // }
    // }
    
    this._parsComponents(this.page);

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
        Router.routing(this.getAttribute(Variables.routerHref));
        Native.request(this.getAttribute(Variables.routerHref));
        Observer.addListener(Variables.responseToRequest, (component) => self._changeComponentDom(component), true);
        return false;
      };
    });

    // if tag have a link to API GET request
    componentDom.querySelectorAll(`[${ Variables.routerAPIGET }]`).forEach((component) => {

      isRequestAPIGET = true;

      Native.request(component.getAttribute(Variables.routerAPIGET));
      component.removeAttribute(Variables.routerAPIGET);

      Observer.addListener(Variables.responseToRequest, (response) => {
        Native.setValueDataByAttr(JSON.parse(response));
        Observer.emit(Variables.documentIsReady, component);
      }, true);

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