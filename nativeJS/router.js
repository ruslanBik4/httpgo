
let currentLink = '';

export class Router {

  /*
  *   Init Router and Native
  */
  static start() {

    this.urls = urls;
    Parse.start(document.body);
    this.routing(window.location.pathname, true);

    window.onpopstate = (obj) => {
      Parse.setStateHistoryComponents();
      if (obj && obj.state) {
        Parse.setComponent(obj.state.component);
      } else {
        Parse.setComponent('', true);
      }
      this.routing(document.location.pathname, true);
      Observer.emit(Variables.documentIsReady);
    };

  }


  /*
  *   Routing, change history and toolbar
  */
  static routing(url, isHistoryBack = false) {

    if (url !== currentLink) {

      currentLink = url;
      const curURL = this.urls[url];
      if (curURL) {
        document.title = curURL.title;
        console.log(curURL);

      } else {
        console.log('/404');
      }

      if (!isHistoryBack) {
        history.pushState({ url: url, component: Parse.mainContent.innerHTML }, '', url);
      }

    }
  }


}
