"use strict";

function AddItem(props, event)
{
    var  typeCanvas = [
            'linear-gauge',
            'radial-gauge',
            'linear-gauge',
            'radial-gauge'
        ],
        propsAttr = {
            renderTo: document.createElement('canvas'),
            width: 160,
            height: 400,
            borderRadius: 50,
            borders: 0,
            barStrokeWidth: 20,
            minorTicks: 10,
            majorTicks: [0,10,20,30,40,50,60,70,80,90,100],
            value: 22.3,
            valueBox: true,
            animationRule: "linear",
            animationDuration: 5000,
            animatedValue: "true",
            needleType: "arrow",
            tickSide: "left",
            needleSide: "right",
            numberSide: "left",
            units: "°C",
            colorValueBoxShadow: true,
            highlights: [
                { "from": 0, "to": 15, "color": "blue" },
                { "from": 35, "to": 100, "color": "red" },
                { "from": 100, "to": 150, "color": "rgba(255,30,0,.25)" },
                { "from": 150, "to": 200, "color": "rgba(255,0,225,.25)" },
                { "from": 200, "to": 220, "color": "rgba(0,0,255,.25)" }
            ]
        },
        dataText = '',
        propText = '',
        elem,
//     настройки для показа на экране
        css = event == undefined ? {left:250, top:250} : {left: event.pageX, top: event.pageY};
    
    for (var i in props ) {
        var text = props[i];

        if ( typeof props[i] == 'object') {
            text = '';
            for (var j in props[i]) {
                text += props[i][j]
            }
        }
        switch (i) {
             case 'left':
             case 'top': 
                 css[i] = text;
                 break;
             default: 
                propsAttr[i] = text;
                dataText += 'data-' + i + '="' + props[i] + '" ';
                 propText += (i == 'title' ? '<b>' + props[i]  + '</b>' : props[i] ) + ', ';
        }
    }
    try {

    // создаем прибор для показа
    if (props.type == 'cam') {
        elem = $('<img id="' + propsAttr['id'] + '" src="img/cam.jpg" title="' + propsAttr['title'] + '" ' + dataText + 'ondblclick="return showVideo(this);"/>');
        propText = '<img src="img/cam.jpg" class="mini-photo"/>' + propText;
    } else if (props.type == 'lamp')  {
        elem = $('<svg id="' + propsAttr['id'] + '" class="draggable lamp" title="' + propsAttr['title'] + '" ' + dataText + '> <circle cx="50%" cy="80%"  r="40%" style="fill:yellow; " /></svg>');
        propText = '<svg class="lamp"> <circle cx="50%" cy="80%"  r="40%" style="fill:yellow; " /></svg>' + propText;
    } else {
        
        var gauge = new LinearGauge( propsAttr ).draw();

        elem = $(gauge.options.renderTo).data(props);
        // new RadialGauge();
        propText = $(gauge.options.renderTo).html() + propText;
    }

    elem.addClass("draggable").attr('title', propsAttr['rflog']).css( css ).appendTo('body');
    elem.mousedown( readyForDragAndDrop ).slideUp('fast').slideDown('slow').slideUp().slideDown();

    } catch (e) {
        console.log(e);
    }

    $('#dRoomtools').append('<div>' + propText + '</div>');
    return false;
}
function getNameFromID(id) {
    var names = {title: "название", value: "значение", password:"пароль"};
    return ( names[id] === undefined ? id : names[id] );
}
function showVideo(thisElem) {
    
    //     Radiant Media Player
    // First we specify bitrates to feed to the player
    var bitrates = { 
          mp4: [ 'http://192.168.2.182' ]
          // Optional WebM fallback - may be useful for older browsers
//           hls:  'rtsp://192.168.2.182/onvif1' 
        };
    // Then we set our player settings 
    var settings = { 
//       licenseKey: '12121212',
      bitrates: bitrates, 
      delayToFade: 3000, 
      width: '99%', 
      height: '100%', 
      skin: 's1', 
      initialBitrate: 1, 
      sharing: true, 
      isLive: true,
      poster: 'https://www.radiantmediaplayer.com/images/poster-rmp-showcase.jpg',
      displayStreams: true 
    };
    // Reference to the wrapper div (unique id) 
    var id = /* thisElem.id + */ 'rmpPlayer1',
        offset = getCoords(thisElem),
//         element = $('<div class="rmpPlayer"><a title="Close" class="fancybox-close" href="javascript:;" onclick="$(this).parent().detach(); return false;"></a><span>' + thisElem.title + '</span><div id="' + id + '"></div></div>').appendTo('body').css({left: offset.left + 20, top: offset.top + 10 }),
    // Create an object based on RadiantMP constructor 
         rmp = new RadiantMP(id);
         
    // Initialization ... test your pages and done!
//     element.mousedown( readyForDragAndDrop );
    rmp.init(settings);
//     rmp.play();
        
    return false;
}
function ExpandUL(thisElem) {
    thisElem.toggleClass('expanded');

    return false;
}

function readyForDragAndDrop(e) {

  var elem = e.target,
      coords = getCoords(elem),
      shiftX = e.pageX - coords.left,
      shiftY = e.pageY - coords.top;

  moveAt(e);

  elem.style.zIndex = 1000; // над другими элементами

  function moveAt(e) {
    elem.style.left = e.pageX - shiftX + 'px';
    elem.style.top = e.pageY - shiftY + 'px';
  }

  document.onmousemove = function(e) {
    moveAt(e);
  };

  document.onmouseup = function() {
    document.onmousemove = null;
    document.onmouseup = null;
      if ($(elem).hasClass('dragging draggable')) {
          $(elem).toggleClass('dragging draggable');

      }
  };
    if ($(elem).hasClass('dragging draggable')) {
        $(elem).toggleClass('dragging draggable');

    }

  elem.ondragstart = function() {
      // $(elem).toggleClass('dragging draggable');
      return false;
  };    
  elem.ondragover = function() {
      // $(elem).toggleClass('dragging draggable');
      return false;
  };  
}

function getCoords(elem) { // кроме IE8-
  var box = elem.getBoundingClientRect();

  return {
    top: box.top + pageYOffset,
    left: box.left + pageXOffset
  };

}
