$(document).ready(function() { 
    
    var conn;
    var log = $("#log");
    var pixelX = [];
    var pixelY = [];
    var drawCounter = 0;
    var width = 600;
    var height = 600;
    var layer = new Konva.Layer();
    var polyLot =[];
    var stage = new Konva.Stage({
      container: 'viewportMid',
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
        document.getElementById("form_x").innerHTML = pixelX;
        document.getElementById("form_y").innerHTML = pixelY;
        drawCounter = 0;
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

            console.log(msgServer);

            var msgType = msgServer.messageprocessedtype;
            var msgOriginal;
            var msgOriginal1DArray=[];
            var msgOriginal2DArray=[];
            var msgProcessed;
            if (msgType==0){
                msgOriginal = msgServer.lot2client[0].messageprocessedoriginal;
                msgProcessed = msgServer.lot2client[0].messageprocessedvectors;
                   appendLog($("<div/>").text(">  Pixel Coor: "+
                       JSON.stringify(msgOriginal)));
                    appendLog($("<div/>").text(" Latitude: "+
                           JSON.stringify(msgProcessed.Latitude).concat(" ,Longtitude: "+
                           JSON.stringify(msgProcessed.Longtitude).concat(" ,Elevation: "+
                           JSON.stringify(msgProcessed.Elevation)))));

            }else if(msgType==1)
            {
                for (i = 0; i < msgServer.lot2client.length; i++){
                    msgOriginal = msgServer.lot2client[i].messageprocessedoriginal;
                    msgProcessed = msgServer.lot2client[i].messageprocessedvectors;

                    msgOriginal2DArray.push(msgOriginal);

                    appendLog($("<div/>").text(">  Pixel Coor: "+
                           JSON.stringify(msgOriginal)));
                        appendLog($("<div/>").text(" Latitude: "+
                               JSON.stringify(msgProcessed.Latitude).concat(" ,Longtitude: "+
                               JSON.stringify(msgProcessed.Longtitude).concat(" ,Elevation: "+
                               JSON.stringify(msgProcessed.Elevation)))));
                        // console.log(msgOriginal)
                }
            }

            for(var i = 0; i < msgOriginal2DArray.length; i++){
               msgOriginal1DArray = msgOriginal1DArray.concat(msgOriginal2DArray[i]);
            }
            drawLots(msgOriginal1DArray);
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

    function getMousePos(canvas, evt) {
        canvas = document.getElementById('viewportTop');
        var rect = canvas.getBoundingClientRect();
        return {
          x: evt.clientX - rect.left,
          y: evt.clientY - rect.top
        };
    }

    function draw(e) {
        if (drawCounter <4){
            var canvas = document.getElementById('viewportTop');
            context = canvas.getContext('2d');
            var pos = getMousePos(canvas, e);
            pixelX.push(pos.x);
            pixelY.push(pos.y);
            document.getElementById("form_x").innerHTML = pixelX;
            document.getElementById("form_y").innerHTML = pixelY;
            context.fillStyle = "white";
            context.beginPath();
            context.arc(pos.x, pos.y, 2, 0, 2*Math.PI);
            context.fill();
            drawCounter++;
        }
        else{
            appendLog($("<div>Please send the pixel array for processing.<\/div>"));
        }
    }

    drawLots();
    function drawLots(msgOriginal1DArray) {
    polyLot = new Konva.Line({
        points: msgOriginal1DArray,
        stroke: '#edeaea',
        strokeWidth: 5,
        closed : true
      });

    polyLot.on('mouseover', function () {
        appendLog($("<div/>").text("in"));
      });

    layer.add(polyLot);
    stage.add(layer);
    }
    
    window.draw = draw;


});
