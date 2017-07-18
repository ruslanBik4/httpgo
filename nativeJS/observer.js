//
// observer.js
//

let isFunction = function(obj) {
  return typeof obj == 'function' || false;
};

let listeners = new Map();

export class Observer {

  constructor() {
  }

  static get listeners() {
    return listeners;
  }

  static addListener(eventName, callback) {

    const listeners = this.listeners.get(eventName);
    const hashCallback = this.hashCode(callback);
    let isListener = false;

    if (listeners && listeners.length) {
      for (let listener of listeners) {
        if (listener.hash === hashCallback) {
          isListener = true;
          break;
        }
      }
    }

    if (!isListener) {
      this.listeners.has(eventName) || this.listeners.set(eventName, []);
      this.listeners.get(eventName).push({ func:callback, hash: this.hashCode(callback) });
    }

  }

  static removeListener(eventName, callback) {
    let listeners = this.listeners.get(eventName);
    let index;
      
    if (listeners && listeners.length) {
      index = listeners.func.reduce((i, listener, index) => {
        return (isFunction(listener) && listener === callback) ? i = index : i;
      }, -1);
      if (index > -1) {
        listeners.func.splice(index, 1);
        this.listeners.set(eventName, listeners.func);
        return true;
      }
    }
    return false;
  }
  
  static emit(eventName, ...args) {
    let result = false;
    let listeners = this.listeners.get(eventName);

    if (listeners && listeners.length) {
      for (let listener of listeners) {
        const isRemove = listener.func(...args);
        if (isRemove) {
          this.removeListener(eventName, listener.func);
        }
      }
      result = true;
    }

    return result;
  }

  static hashCode(str) {
    let hash = 0;
    str = (typeof str === 'string') ? str : str.toString();

    if (str.length == 0) return hash;
    for (let i = 0; i < str.length; i++) {
      let char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32bit integer
    }
    return hash;
  }
}