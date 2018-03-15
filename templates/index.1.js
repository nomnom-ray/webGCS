var conn;
var pixelX = [];
var pixelY = [];
var drawCounter = 0;
var width = 600;
var height = 600;
var layer = new Konva.Layer();
var drawLine = [];
var polyLot = [];
var infoLots = [];
var featureIDLot = 0;

var log = document.getElementById('log');
var stage = new Konva.Stage({
    container: 'viewportTop',
    width: width,
    height: height
});

var mapG;
var leafMaps = L.map('leafMaps', {
    minZoom: 1,
    maxZoom: 1,
    center: [0, 0],
    zoom: 1,
    crs: L.CRS.Simple
});


leafInit();

appendLog("<div><b>Please mark the 4 corners of a parking spot and send.</b></div>");
// make_base();
// $(window).focus(function() {
//     make_base();
//     var pixelX = [];
//     var pixelY = [];
//     var drawCounter = 0;
// });


function leafInit() {
    var northEast = leafMaps.unproject([width, 0], leafMaps.getMaxZoom());
    var southWest = leafMaps.unproject([0, height], leafMaps.getMaxZoom());
    var imageBounds = new L.LatLngBounds(southWest, northEast);
    var imageUrl = 'http://localhost:8080/templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';

    leafMaps.setMaxBounds(new L.LatLngBounds(southWest, northEast));
    L.imageOverlay(imageUrl, imageBounds).addTo(leafMaps);
    leafMaps.setMaxBounds(imageBounds);
}

// define toolbar options
var options = {
    position: 'topleft', // toolbar position, options are 'topleft', 'topright', 'bottomleft', 'bottomright'
    drawMarker: true, // adds button to draw markers
    drawPolyline: false, // adds button to draw a polyline
    drawRectangle: false, // adds button to draw a rectangle
    drawPolygon: true, // adds button to draw a polygon
    drawCircle: false, // adds button to draw a cricle
    cutPolygon: false, // adds button to cut a hole in a polygon
    editMode: false, // adds button to toggle edit mode for all layers
    removalMode: true, // adds a button to remove layers
        // the lines between coordinates/markers
        templineStyle: {
            color: 'red',
        },
    
        // the line from the last marker to the mouse cursor
        hintlineStyle: {
            color: 'red',
            dashArray: [5, 5],
        },
        pathOptions: {
            // add leaflet options for polylines/polygons
            color: 'orange',
            fillColor: 'green',
            fillOpacity: 0.4,
        },
};

// add leaflet.pm controls to the map
leafMaps.pm.addControls(options);

// enable drawing mode for shape to change change
leafMaps.pm.enableDraw('Poly', options);

// // listen to when drawing mode gets enabled
// leafMaps.on('pm:drawstart', function(e) {
//     console.log(e.shape); // the name of the shape being drawn (i.e. 'Circle')
//     console.log(e.workingLayer); // the leaflet layer displayed while drawing
// });

// listen to when a new layer is created
leafMaps.on('pm:create', function(e) {
    e.shape; // the name of the shape being drawn (i.e. 'Circle')
    console.log(e.layer); // the leaflet layer created

    for (var name in e.layer) {
        if (e.layer.hasOwnProperty(name)) {
            console.log(name);
        }
      }

});

// toggle global removal mode
leafMaps.pm.toggleGlobalRemovalMode();

// listen to removal of layers that are NOT ignored and NOT helpers by leaflet.pm
leafMaps.on('pm:remove', function(e) {});




function initMap() {
    mapG = new google.maps.Map(document.getElementById('googleMaps'), {
        center: {
            lat: 43.451791,
            lng: -80.496825
        },
        // mapTypeId: 'satellite',
        zoom: 18
    });
}



// TODO:leave google maps until after leafmaps can send geojson to server

function drawMapLots(mapG, featureIDLot) {

    for (i = 0; i < featureIDLot; i++) {
        var mapLotsCoor = [{
                lat: infoLots[i][0],
                lng: infoLots[i][1]
            }, // north west
            {
                lat: infoLots[i][2],
                lng: infoLots[i][3]
            }, // south west
            {
                lat: infoLots[i][4],
                lng: infoLots[i][5]
            }, // south east
            {
                lat: infoLots[i][6],
                lng: infoLots[i][7]
            } // north east
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

        //TODO: make this just for connection, instruction and error messages

function appendLog(message) {
    var d = log;
    var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
    log.innerHTML += message;
    if (doScroll) {
        d.scrollTop = d.scrollHeight - d.clientHeight;
    }
}


        //TODO: copy geojson.io

document.getElementById('form').onsubmit = function () {
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

    conn.addEventListener('message', function (e) {
        var msgServer = JSON.parse(e.data);

        // console.log(msgServer);
        //TODO: replace with geojson
        var msgType = msgServer.messageprocessedtype;
        var msgOriginal;
        var msgOriginal1DArray = [];
        var msgOriginal2DArray = [];
        var msgProcessed;
        var infoLot = [];
        if (msgType == 0) {
            msgOriginal = msgServer.lot2client[0].messageprocessedoriginal;
            msgProcessed = msgServer.lot2client[0].messageprocessedvectors;
            appendLog("<div>> Pixel Coor: " +
                JSON.stringify(msgOriginal) +
                " Latitude: " +
                JSON.stringify(msgProcessed.Latitude) + ", Longtitude: " +
                JSON.stringify(msgProcessed.Longtitude) + ", Elevation: " +
                JSON.stringify(msgProcessed.Elevation + '</div>'));

        } else if (msgType == 1) {
            for (i = 0; i < msgServer.lot2client.length; i++) {
                msgOriginal = msgServer.lot2client[i].messageprocessedoriginal;
                msgProcessed = msgServer.lot2client[i].messageprocessedvectors;

                msgOriginal2DArray.push(msgOriginal);
                infoLot.push(msgProcessed.Latitude, msgProcessed.Longtitude);
            }
            for (var x = 0; x < msgOriginal2DArray.length; x++) {
                msgOriginal1DArray = msgOriginal1DArray.concat(msgOriginal2DArray[x]);
            }
            infoLots.push(infoLot);
            drawLots(msgOriginal1DArray, featureIDLot, infoLots[featureIDLot]);
            featureIDLot++;
        } else if (msgType == 9) {
            appendLog("<div><b>Please Click within the vertex map.</b></div>");
        }

        //TODO: delete
        // make_base();
        //TODO:reactive google maps when leafmaps is done; maybe consider its new implementation
        // drawMapLots(mapG, featureIDLot);
    });

    conn.onclose = function (evt) {
        appendLog("<div><b>Connection closed.</b></div>");
    };
} else {
    appendLog("<div><b>Your browser does not support WebSockets.</b></div>");
}

// #########garbage#######################
// #########garbage#######################

function make_base() {
    var canvas = document.getElementById('viewportBottom');
    context = canvas.getContext('2d');
    context.clearRect(0, 0, width, height);
    base_image = new Image();
    base_image.src = 'templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';
    base_image.onload = function () {
        context.drawImage(base_image, 0, 0, width, height);
    };
}

//TODO:use gEOJSON instead of push arrays

document.getElementById("viewportTop").onclick = function (e) {

    var relX = e.clientX;
    var relY = e.clientY;


    if (drawCounter < 4) {
        pixelX.push(relX);
        pixelY.push(relY);
        document.getElementById("form_x").innerHTML = pixelX;
        document.getElementById("form_y").innerHTML = pixelY;
        drawCounter++;
        drawClick(relX, relY);

        drawLine.push(relX, relY);
        var featureIDLot = 0;
        var infoLots = 0;
        drawLots(drawLine, featureIDLot, infoLots);


    } else {
        appendLog("<div><b>Please send the pixel array for processing.</b></div>");
        drawLine = [];
    }
};


function drawClick(relX, relY) {
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

function drawLots(msgOriginal1DArray, featureIDLot, infoLot) {

    var closeLot = false;
    if (msgOriginal1DArray.length >= 8) {
        closeLot = true;
    }
    polyLot = new Konva.Line({
        points: msgOriginal1DArray,
        stroke: '#edeaea',
        strokeWidth: 5,
        closed: closeLot
    });

    polyLot.on('click', function () {

        appendLog("<div>> Latitude: " +
            JSON.stringify(infoLot[0]) + ", Longtitude: " +
            JSON.stringify(infoLot[1]) + '</div>');
        appendLog("<div>> Latitude: " +
            JSON.stringify(infoLot[2]) + ", Longtitude: " +
            JSON.stringify(infoLot[3]) + '</div>');
        appendLog("<div>> Latitude: " +
            JSON.stringify(infoLot[4]) + ", Longtitude: " +
            JSON.stringify(infoLot[5]) + '</div>');
        appendLog("<div>> Latitude: " +
            JSON.stringify(infoLot[6]) + ", Longtitude: " +
            JSON.stringify(infoLot[7]) + '</div>');
    });

    layer.add(polyLot);
    stage.add(layer);
}