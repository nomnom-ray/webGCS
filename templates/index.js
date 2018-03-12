$(document).ready(function() { 

    var conn;
    var log = $("#log");
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
    var stage = new Konva.Stage({
      container: 'viewportTop',
      width: width,
      height: height
    });

    appendLog($("<div><b>Please mark the 4 corners of a parking spot and send.<\/b><\/div>"));
    make_base();
    $(window).focus(function() {
        make_base();
        var pixelX = [];
        var pixelY = [];
        var drawCounter = 0;

    });
    function appendLog(message) {
        var d = log[0];
        var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
        message.appendTo(log);
        if (doScroll) {
            d.scrollTop = d.scrollHeight - d.clientHeight;
        }
    }

    $("#form").submit(function() {
        if (!pixelX[0] || !pixelX[0]) {
            return false;
        }
        if (pixelX.length == 2 || pixelX.length == 3) {
            appendLog($("<div><b>Please either mark 1 or 4 points.<\/b><\/div>"));
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
    });

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
                   appendLog($("<div/>").text(">  Pixel Coor: "+
                       JSON.stringify(msgOriginal)));
                    appendLog($("<div/>").text(" Latitude: "+
                           JSON.stringify(msgProcessed.Latitude).concat(" ,Longtitude: "+
                           JSON.stringify(msgProcessed.Longtitude).concat(" ,Elevation: "+
                           JSON.stringify(msgProcessed.Elevation)))));

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
                appendLog($("<div><b>Please Click within the vertex map.<\/b><\/div>"));
            }

            make_base();
        });

        conn.onclose = function(evt) {
            appendLog($("<div><b>Connection closed.<\/b><\/div>"));
        };
    } else {
        appendLog($("<div><b>Your browser does not support WebSockets.<\/b><\/div>"));
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

    $("#viewportTop").click(function(e){
 
        var parentOffset = $(this).offset(); 
        var relX = e.pageX - parentOffset.left;
        var relY = e.pageY - parentOffset.top;
        
        
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
            appendLog($("<div>Please send the pixel array for processing.<\/div>"));
            drawLine = [];
        }
     });


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

    polyLot.on('mouseover', function () {

    appendLog($("<div/>").text("> Coordinates of Lot Corners: "));
    appendLog($("<div/>").text(" Point 1 <==> Latitude: "+
        JSON.stringify(infoLot[0]).concat(" ,Longtitude: "+
        JSON.stringify(infoLot[1]))));
    appendLog($("<div/>").text(" Point 2 <==> Latitude: "+
        JSON.stringify(infoLot[2]).concat(" ,Longtitude: "+
        JSON.stringify(infoLot[3]))));
    appendLog($("<div/>").text(" Point 3 <==> Latitude: "+
        JSON.stringify(infoLot[4]).concat(" ,Longtitude: "+
        JSON.stringify(infoLot[5]))));
    appendLog($("<div/>").text(" Point 4 <==> Latitude: "+
        JSON.stringify(infoLot[6]).concat(" ,Longtitude: "+
        JSON.stringify(infoLot[7]))));

    });

    layer.add(polyLot);
    stage.add(layer);
    }
});
