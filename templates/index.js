$(document).ready(function() { 
    var conn;
    var log = $("#log");
    var pixelX = [];
    var pixelY = [];
    var drawCounter = 0;

    appendLog($("<div><b>Parking coordinates saved on server.<\/b><\/div>"))
    make_base();

    function appendLog(message) {
        var d = log[0]
        var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
        message.appendTo(log)
        if (doScroll) {
            d.scrollTop = d.scrollHeight - d.clientHeight;
        }
    }

    $("#form").submit(function() {
        
        if (!pixelX[0] || !pixelX[0]) {
            return false;
        }
        var message4Server = {
            pixelX1: parseInt(pixelX[0]),
            pixelY1: parseInt(pixelY[0])
        }
        message4Server = JSON.stringify(message4Server);
        pixelX = [];
        pixelY = [];
        document.getElementById("form_x").innerHTML = pixelX;
        document.getElementById("form_y").innerHTML = pixelY;
        drawCounter = 0;
        if (!conn) {
            return false;
        }

        conn.send(message4Server);
        return false
    });

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://localhost:8080/ws");
        
        conn.addEventListener('message', function(e){
            var msgServer=JSON.parse(e.data);
            appendLog($("<div/>").text(msgServer.messageprocessed))
            make_base();
        });

        conn.onclose = function(evt) {
            appendLog($("<div><b>Connection closed.<\/b><\/div>"))
        }
    } else {
        appendLog($("<div><b>Your browser does not support WebSockets.<\/b><\/div>"))
    }

    function make_base() {
        var canvas = document.getElementById('viewportBottom');
        context = canvas.getContext('2d');
    base_image = new Image();
    base_image.src = 'templates/fBW-loc43_4516288_-80_4961367-fov80-heading115-pitch-10.jpg';
        base_image.onload = function(){
            context.drawImage(base_image, 0, 0,600,600);
        }
    }

    function getMousePos(canvas, evt) {
        var rect = canvas.getBoundingClientRect();
        return {
          x: evt.clientX - rect.left,
          y: evt.clientY - rect.top
        };
    }

    function draw(e) {
        if (drawCounter <1){
            var canvas = document.getElementById('viewportBottom');
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
            appendLog($("<div>Please send the pixel array for processing.<\/div>"))
        }
    }
    window.draw = draw;

});


