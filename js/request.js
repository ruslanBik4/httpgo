/**
 * Created by ruslan on 7/27/17.
 * TODO: сделать системной функцией в нативе - пусть получает имя элемента, куда писать респонс или пишет в боди
 */

function request(url) {

    var method = 'GET';
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
        var d = document.createElement('div');
        d.style.marginTop = '20px';
        d.innerHTML = response.currentTarget.responseText;
        document.body.append(d);
    }
}
