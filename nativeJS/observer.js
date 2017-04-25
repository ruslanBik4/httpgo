//
// observer.js
//

let isFunction = function(obj) {
  return typeof obj == 'function' || false;
};

let listeners = new Map();
let listenersRemoving = new Map();

export class Observer {

  constructor() {
  }

  static get listeners() {
    return listeners;
  }

  static get listenersRemoving() {
    return listenersRemoving;
  }

  static addListener(eventName, callback, isRemoveListenerAfterComplete = false) {
    this.listeners.has(eventName) || this.listeners.set(eventName, []);
    this.listeners.get(eventName).push(callback);
    if (isRemoveListenerAfterComplete) {
      this.listenersRemoving.has(eventName) || this.listenersRemoving.set(eventName, []);
      this.listenersRemoving.get(eventName).push(callback);
    }
  }

  static removeListener(eventName, callback) {
    let listeners = this.listeners.get(eventName);
    let index;
      
    if (listeners && listeners.length) {
      index = listeners.reduce((i, listener, index) => {
        return (isFunction(listener) && listener === callback) ? i = index : i;
      }, -1);
      if (index > -1) {
        listeners.splice(index, 1);
        this.listeners.set(eventName, listeners);
        return true;
      }
    }
    return false;
  }
  
  static emit(eventName, ...args) {
    let result = false;
    let listeners = this.listeners.get(eventName);
    let listenersRemoving = this.listenersRemoving.get(eventName);

    if (listeners && listeners.length) {
      for (let listener of listeners) {
        listener(...args);
        if (listenersRemoving) {
          for (let i = 0; i < listenersRemoving.length; i++) {
            if (listener === listenersRemoving[i]) {
              this.removeListener(eventName, listenersRemoving[i]);
              listenersRemoving.splice(i, 1);
              this.listenersRemoving.set(eventName, listenersRemoving);
              break;
            }
          }
        }
      }
      result = true;
    }

    return result;
  }
}