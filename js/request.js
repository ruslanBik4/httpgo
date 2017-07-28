/**
 * Created by ruslan on 7/27/17.
 * TODO: сделать системной функцией в нативе - пусть получает имя элемента, куда писать респонс или пишет в боди
 */

/*
*   for queue of requests
*/

function queueRequests(arr) {

    request(arr.shift(), queue);

    function queue(response) {
        if (response.status) {
            var url = arr.shift();
            waitMessage(url);
            createDivInBody(response.responseText);
            request(url, queue);
        } else {
            console.error(response);
            createDivInBody(response.errors);
        }
    }
}

function waitMessage(msg) {
  var p = document.createElement('p');
  p.style.marginTop = '20px';
  p.innerHTML = 'Request to server by address: ' + msg + ', please wait ...';
  document.body.append(p);
}

function createDivInBody(text) {
    var d = document.createElement('div');
    d.style.marginTop = '20px';
    d.innerHTML = text;
    document.body.append(d);
}

function request(url, callback) {

    var body = ['\r\n'];

    var XHR = 'onload' in new XMLHttpRequest() ? XMLHttpRequest : XDomainRequest;
    var xhr = new XHR();

    xhr.open('POST', url, true);

    var boundary = String(Math.random()).slice(2);
    var boundaryMiddle = '--' + boundary + '\r\n';
    var boundaryLast = '--' + boundary + '--\r\n';

    body = body.join(boundaryMiddle) + boundaryLast;
    xhr.setRequestHeader('Content-Type', 'multipart/form-data; boundary=' + boundary);


    xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");
    xhr.send(body);

    xhr.onload = function (response) {
        callback(response.currentTarget, url);
    };

    xhr.onerror = function () {
        console.error('Error API to url ' + url + ' : ' + this);
    };
}
