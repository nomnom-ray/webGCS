var conn;
var pixelX = [];
var pixelY = [];
var drawCounter = 0;
var width = 600;
var height = 600;
var layer = new Konva.Layer();
var drawLine = [];
var polyLot =[];
var infoLots =[];
var featureIDLot =0;

var mapG;
function initMap() {
    mapG = new google.maps.Map(document.getElementById('googleMaps'), {
    center: {lat: 43.451791, lng: -80.496825},
    // mapTypeId: 'satellite',
    zoom: 18
  });

}

function drawMapLots(mapG,featureIDLot){

    for (i = 0; i < featureIDLot; i++){
        var mapLotsCoor = [
            {lat: infoLots[i][0], lng:  infoLots[i][1]}, // north west
            {lat: infoLots[i][2], lng:  infoLots[i][3]}, // south west
            {lat: infoLots[i][4], lng:  infoLots[i][5]}, // south east
            {lat: infoLots[i][6], lng: infoLots[i][7]}  // north east
        ];
        // mapG.data.add({geometry: new google.maps.Data.Polygon([mapLotsCoor])});

        var mapLots = new google.maps.Polygon({
            paths: mapLotsCoor,
            strokeColor: '#FF0000',
            strokeOpacity: 0.8,
            strokeWeight: 3,
            fillColor: '#FF0000',
            fillOpacity: 0.35
          });
          mapLots.setMap(mapG);

        }
}


 

    var log = document.getElementById('log');
    var stage = new Konva.Stage({
      container: 'viewportTop',
      width: width,
      height: height
    });

    appendLog("<div><b>Please mark the 4 corners of a parking spot and send.</b></div>");
    make_base();
    // $(window).focus(function() {
    //     make_base();
    //     var pixelX = [];
    //     var pixelY = [];
    //     var drawCounter = 0;

    // });
    function appendLog(message) {
        var d = log;
        var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
        log.innerHTML += message;
        if (doScroll) {
            d.scrollTop = d.scrollHeight - d.clientHeight;
        }
    }

    document.getElementById('form').onsubmit = function() {
        if (!pixelX[0] || !pixelX[0]) {
            return false;
        }
        if (pixelX.length == 2 || pixelX.length == 3) {
            appendLog("<div><b>Please either mark 1 or 4 points.</b></div>");
            pixelX = [];
            pixelY = [];
            document.getElementById("form_x").innerHTML = pixelX;
            document.getElementById("form_y").innerHTML = pixelY;
            drawCounter = 0;
            return false;
        }
        var message4Server = {
            lotX0: parseInt(pixelX[0]),
            lotY0: parseInt(pixelY[0]),
            lotX1: parseInt(pixelX[1]),
            lotY1: parseInt(pixelY[1]),
            lotX2: parseInt(pixelX[2]),
            lotY2: parseInt(pixelY[2]),
            lotX3: parseInt(pixelX[3]),
            lotY3: parseInt(pixelY[3])
        };
        message4Server = JSON.stringify(message4Server);
        make_base();
        pixelX = [];
        pixelY = [];
        drawCounter = 0;
        drawLine = [];
        document.getElementById("form_x").innerHTML = pixelX;
        document.getElementById("form_y").innerHTML = pixelY;

        if (!conn) {
            return false;
        }
        conn.send(message4Server);
        return false;
    };



    if (window.WebSocket) {
        conn = new WebSocket("ws://localhost:8080/ws");
        
        conn.addEventListener('message', function(e){
            var msgServer=JSON.parse(e.data);

            // console.log(msgServer);

            var msgType = msgServer.messageprocessedtype;
            var msgOriginal;
            var msgOriginal1DArray=[];
            var msgOriginal2DArray=[];
            var msgProcessed;
            var infoLot=[];
            if (msgType==0){
                msgOriginal = msgServer.lot2client[0].messageprocessedoriginal;
                msgProcessed = msgServer.lot2client[0].messageprocessedvectors;
                   appendLog("<div>> Pixel Coor: "+
                       JSON.stringify(msgOriginal) +
                      " Latitude: "+
                           JSON.stringify(msgProcessed.Latitude) + ", Longtitude: "+
                           JSON.stringify(msgProcessed.Longtitude) + ", Elevation: "+
                           JSON.stringify(msgProcessed.Elevation + '</div>'));

            }else if(msgType==1){
                for (i = 0; i < msgServer.lot2client.length; i++){
                    msgOriginal = msgServer.lot2client[i].messageprocessedoriginal;
                    msgProcessed = msgServer.lot2client[i].messageprocessedvectors;

                    msgOriginal2DArray.push(msgOriginal);
                    infoLot.push(msgProcessed.Latitude,msgProcessed.Longtitude); 
                }
                for(var x = 0; x < msgOriginal2DArray.length; x++){
                    msgOriginal1DArray = msgOriginal1DArray.concat(msgOriginal2DArray[x]);
                }
                infoLots.push(infoLot);
                drawLots(msgOriginal1DArray,featureIDLot,infoLots[featureIDLot]);
                featureIDLot++;
            } else if(msgType==9){
                appendLog("<div><b>Please Click within the vertex map.</b></div>");
            }

            make_base();
            drawMapLots(mapG,featureIDLot);
        });

        conn.onclose = function(evt) {
            appendLog("<div><b>Connection closed.</b></div>");
        };
    } else {
        appendLog("<div><b>Your browser does not support WebSockets.</b></div>");
    }
    
    function make_base() {
        var canvas = document.getElementById('viewportBottom');
        context = canvas.getContext('2d');
        context.clearRect(0, 0, width, height);
        base_image = new Image();
        base_image.src = 'templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';
        base_image.onload = function(){
        context.drawImage(base_image, 0, 0,width,height);
        };
    }

    document.getElementById("viewportTop").onclick = function(e){
 
        var relX = e.clientX;
        var relY = e.clientY;
        
        
        if (drawCounter <4){
            pixelX.push(relX);
            pixelY.push(relY);
            document.getElementById("form_x").innerHTML = pixelX;
            document.getElementById("form_y").innerHTML = pixelY;
            drawCounter++;
            drawClick(relX,relY);

            drawLine.push(relX,relY);
            var featureIDLot=0;
            var infoLots =0;
            drawLots(drawLine,featureIDLot,infoLots);
            

        }
        else{
            appendLog("<div><b>Please send the pixel array for processing.</b></div>");
            drawLine = [];
        }
     };


    function drawClick(relX,relY) {
        var circle = new Konva.Circle({
            x: relX,
            y: relY,
            radius: 3,
            fill: 'white',
            stroke: 'white',
            strokeWidth: 1
          });

          layer.add(circle);
          stage.add(layer);
    }

    function drawLots(msgOriginal1DArray,featureIDLot, infoLot) {

    var closeLot = false;
    if (msgOriginal1DArray.length >=8){
        closeLot = true;
    }
    polyLot = new Konva.Line({
        points: msgOriginal1DArray,
        stroke: '#edeaea',
        strokeWidth: 5,
        closed : closeLot
      });

    polyLot.on('click', function () {

    appendLog("<div>> Latitude: "+
       JSON.stringify(infoLot[0]) + ", Longtitude: "+
       JSON.stringify(infoLot[1]) + '</div>');
    appendLog("<div>> Latitude: "+
       JSON.stringify(infoLot[2]) + ", Longtitude: "+
       JSON.stringify(infoLot[3]) + '</div>');
    appendLog("<div>> Latitude: "+
       JSON.stringify(infoLot[4]) + ", Longtitude: "+
       JSON.stringify(infoLot[5]) + '</div>');
    appendLog("<div>> Latitude: "+
       JSON.stringify(infoLot[6]) + ", Longtitude: "+
       JSON.stringify(infoLot[7]) + '</div>');
    });

    layer.add(polyLot);
    stage.add(layer);
    }
