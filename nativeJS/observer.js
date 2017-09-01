//
// observer.js
//

let isFunction = function(obj) {
  return typeof obj == 'function' || false;
};

let listeners = new Map();

export class Observer {

  /*
  *   Get all listener
  *   return Map()
  */
  static get listeners() {
    return listeners;
  }

  /*
  *   Add listener to Observer
  */
  static addListener(eventName, callback) {

    const listeners = this.listeners.get(eventName);
    const hashCallback = this.hashCode(callback);


    if (listeners && listeners.length) {
      for (let i = 0; i < listeners.length; i++) {
        if (listeners[i].hash === hashCallback) {
          listeners.splice(i, 1);
          break;
        }
      }
    }

    this.listeners.get(eventName) || this.listeners.set(eventName, []);
    this.listeners.get(eventName).push({ func:callback, hash: this.hashCode(callback) });

  }


  /*
   *   Remove listener from Observer
   */
  static removeListener(eventName, callback) {
    let listeners = this.listeners.get(eventName);
    let index;
      
    if (listeners && listeners.length) {
      const hashCallback = this.hashCode(callback);
      index = listeners.reduce((i, listener, index) => {
        return (isFunction(listener.func) && listener.hash === hashCallback) ? i = index : i;
      }, -1);
      if (index > -1) {
        listeners.splice(index, 1);
        this.listeners.set(eventName, listeners.func);
        return true;
      }
    }
    return false;
  }


  /*
   *   Emit listener from Observer
   */
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


  /*
   *   Create hash code for listener
   */
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