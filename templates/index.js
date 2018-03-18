var conn;

window.onbeforeunload = function () {
    conn.close();
};

if (window.WebSocket) {
    appendLog("<div><b>" + '\xa0\xa0' + "Connection Established.</b></div>");
    appendLog("<div>" + '\xa0\xa0' +
        "> Use the palette tools to beginning annotation. " +
        "Deselect a tool by clicking it again. " +
        "Delete an annotation by selecting the Bin tool and clicking on the annoation.</div>");

    conn = new WebSocket("ws://localhost:8080/ws");
    conn.addEventListener('message', function (e) {
        var msgServer = JSON.parse(e.data);
        if(!msgServer.features){
            // It doesn't exist, do nothing
        } else {
            console.log(msgServer.features.properties.prop0[0][0]);
        }


    });
    conn.onclose = function (evt) {
        appendLog("<div><b>" + '\xa0\xa0' + "Connection closed.</b></div>");
    };
} else {
    appendLog("<div><b>" + '\xa0\xa0' + "Please Use a browser that supports WebSockets.</b></div>");
}

leafInit();

function appendLog(message) {
    var d = document.getElementById('log');
    var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
    d.innerHTML += message;
    if (doScroll) {
        d.scrollTop = d.scrollHeight - d.clientHeight;
    }
}

function leafInit() {
    var width = 600;
    var height = 600;
    var leafMaps = L.map('leafMaps', {
        // minZoom: 1,
        // maxZoom: 1,
        zoom: 0,
        zoomControl: false,
        center: [300, 300],
        crs: L.CRS.Simple
    });
    var northEast = leafMaps.unproject([0, width]);
    var southWest = leafMaps.unproject([height, 0]);
    var imageBounds = new L.LatLngBounds(southWest, northEast);
    var imageUrl = 'http://localhost:8080/templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';

    leafMaps.setMaxBounds(imageBounds);
    L.imageOverlay(imageUrl, imageBounds).addTo(leafMaps);

    leafMaps.dragging.disable();

    leafDraw(leafMaps);
}

function leafDraw(leafMaps) {
    // define toolbar and polygon options, adn initialize
    var options = {
        position: 'topleft',
        drawMarker: true,
        drawPolyline: false,
        drawRectangle: false,
        drawPolygon: true,
        drawCircle: false,
        cutPolygon: false,
        editMode: false,
        removalMode: true,
        templineStyle: {
            color: 'red',
        },
        hintlineStyle: {
            color: 'red',
            dashArray: [5, 5],
        },
        pathOptions: {
            color: 'orange',
            fillColor: 'green',
            fillOpacity: 0.4,
        },
    };
    leafMaps.pm.addControls(options);
    leafMaps.pm.enableDraw('Poly', options);
    leafMaps.pm.disableDraw('Poly');

    leafMaps.on('pm:drawstart', function (e) {
        if (e.shape === "Marker") {
            appendLog("<div>" + '\xa0\xa0' + "> Place marker to get a location coordinate.</div>");
        } else if (e.shape === "Poly") {
            appendLog("<div>" + '\xa0\xa0' + "> Draw polygon to get coordinates encapsulating a feature.</div>");
        }
    });

    leafMaps.on('pm:create', function (e) {

        var geometryType;
        var pixCoordinates;
        var annotationType;

        if (e.shape === "Marker") {
            geometryType = e.shape;
            pixCoordinates = e.layer._latlng;
            annotationType = "Point";
        } else if (e.shape === "Poly") {
            geometryType = e.shape;
            pixCoordinates = e.layer._latlngs;
            annotationType = "LotParking";
        }
        var message4Server = createGEOJSON(geometryType, pixCoordinates, annotationType);
        message4Server = JSON.stringify(message4Server);

        console.log(e.layer);
        for (var name in e.layer) {
            if (e.layer.hasOwnProperty(name)) {
                console.log(name);
            }
          }

        var sent = toServer(message4Server);
        if (!sent) {
            appendLog("<div><b>" + '\xa0\xa0' + "Message not sent.</b></div>");
        }
    });
}

function toServer(message4Server) {
    if (!conn) {
        return false;
    }
    conn.send(message4Server);
    return true;
}

function createGEOJSON(geometryType, pixCoordinates, annotationType) {

    var annotation;
    var pixCoordinates1Array = [];

    if (geometryType === "Marker") {

        var pixelCoordinates = [pixCoordinates.lng, pixCoordinates.lat]; 

        annotation = turf.point([0,0], {
            annotationType: annotationType,
            pixelCoordinates: pixelCoordinates 
            });
    } else if (geometryType === "Poly") {
        // for (i = 0; i < coordinates.length; i++) {
        var ringCloser = pixCoordinates[0][0];
        pixCoordinates[0].push(ringCloser);

        for (i = 0; i < pixCoordinates[0].length; i++) {
            pixCoordinates1Array.push([pixCoordinates[0][i].lng, pixCoordinates[0][i].lat]);
        }
        annotation = turf.polygon([[[0,0],[0,0],[0,0],[0,0]]], {
            annotationType: annotationType,
            pixelCoordinates: [pixCoordinates1Array]
        });
        // }    
    }
    return annotation;
}


// TODO:leave google maps until after leafmaps can send geojson to server
function initMap() {
    var mapG;
    mapG = new google.maps.Map(document.getElementById('googleMaps'), {
        center: {
            lat: 43.451791,
            lng: -80.496825
        },
        // mapTypeId: 'satellite',
        zoom: 18
    });
}