
let currentLink = '';

export class Router {
  constructor() {
  }

  static start() {

    this.urls = urls;
    this.parse = new Parse(document.body);
    this.routing(window.location.pathname, true);

    window.onpopstate = (obj) => {
      console.log(obj.state);
      this.routing(document.location.pathname, true);
    };

  }

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
        history.pushState({ url: url }, '', url);
      }

    }
  }



  /*
   * TODO: need for bicycle, later it's removed
  */

  static _bicycle(url) {
    $('div.autoload').each( function () {
    });
  }


}
