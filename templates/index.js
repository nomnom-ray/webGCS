var gMaps;
// var gDeleteAnnotation;
function initMap() {
    gMaps = new google.maps.Map(document.getElementById('googleMaps'), {
        center: {
            lat: 43.451774,
            lng: -80.496678
        },
        zoom: 20,
        streetViewControl: false,
        fullscreenControl: false,
        zoomControl: false,
        rotateControl: true,
        tilt: 45,
        mapTypeId: 'terrain',
        mapTypeControl: true,
        mapTypeControlOptions: {
            style: google.maps.MapTypeControlStyle.HORIZONTAL_BAR,
            position: google.maps.ControlPosition.TOP_RIGHT
        }
    });

    var infowindow = new google.maps.InfoWindow();

    gMaps.data.setStyle({
        fillColor: '#3794ff',
        fillOpacity: 0.3,
        strokeWeight: 1,
        strokeColor: '#3794ff',
        // icon: 'https://cdn.rawgit.com/pointhi/leaflet-color-markers/master/img/marker-icon-blue.png'
    });

    var selectionToggle = false;
    gMaps.data.addListener('click', function (e) {

        gMaps.data.revertStyle();

        selectionToggle = !selectionToggle;
        if (selectionToggle == true) {
            gMaps.data.overrideStyle(e.feature, {
                fillColor: '#3fdbff',
                fillOpacity: 0.3,
                strokeWeight: 3,
                strokeColor: '#3fdbff'
            });
            gDeleteAnnotation = e.feature;
            var annotationType = e.feature.getProperty("annotationType");
            if (annotationType == "Point") {
                e.feature.getGeometry().forEachLatLng(function (path) {
                    appendLog("<div>" + '\xa0\xa0' +
                        "> Latitude: " + path.lat() + "<--> Longtitude: " + path.lng() + "</div>");
                });
                infowindow.setPosition(e.feature.getGeometry().get());
                infowindow.setOptions({
                    pixelOffset: new google.maps.Size(0, -30)
                });
            } else if (annotationType == "LotParking") {
                e.feature.getGeometry().forEachLatLng(function (path) {
                    appendLog("<div>" + '\xa0\xa0' +
                        "> Latitude: " + path.lat() + "<--> Longtitude: " + path.lng() + "</div>");
                });
                infowindow.setPosition(e.latLng);
                infowindow.setOptions({
                    pixelOffset: new google.maps.Size(0, -5)
                });
            }

            infowindow.setContent(
                '<div style="width:150px; text-align: left;">' + "This feature is a " +
                annotationType + '. Coordinates in textbox.</div><br/>' +
                '<button id="deleteButton" onclick="deleteAnnotation(gDeleteAnnotation);">Delete</button>' +
                '<button id="navButton" onclick="navigation()">Navigate</button>'
            );

            infowindow.open(gMaps);
        } else {
            gMaps.data.overrideStyle(e.feature, {
                fillColor: '#3794ff',
                fillOpacity: 0.3,
                strokeWeight: 1,
                strokeColor: '#3794ff'
            });
            infowindow.close(gMaps);
        }
    });
    gMaps.data.addListener('removefeature', function () {
        infowindow.close(gMaps);
        selectionToggle = false;
    });

}

function deleteAnnotation(gDeleteAnnotation) {

    var featureUUID = gDeleteAnnotation.getProperty("annotationID");
    geojson.eachLayer(function (layer) {
        // layer.feature is the original geojson feature
        if (layer.feature.properties.annotationID === featureUUID) {
            geojson.removeLayer(layer);
        }
    });

    gMaps.data.remove(gDeleteAnnotation);
    gDeleteAnnotation = 0;
}

if (window.WebSocket) {

    var conn;
    window.onbeforeunload = function () {
        conn.close();
    };

    appendLog("<div><b>" + '\xa0\xa0' + "Connection Established.</b></div>");
    appendLog("<div>" + '\xa0\xa0' +
        "> Use the palette tools to beginning annotation. " +
        "Deselect a tool by selecting it again. " +
        "Delete an annotation by selecting the Bin tool and clicking on the annoation.</div>");

    conn = new WebSocket("ws://localhost:8080/ws");
    conn.addEventListener('message', function (e) {
        var msgServer = JSON.parse(e.data);
        if (!msgServer.features) {
            // It doesn't exist, do nothing
        } else {
            if (msgServer.features.properties.annotationStatus == "no error"){
                gMaps.data.addGeoJson(msgServer.features);
            }
        }
    });

    leafInit(conn);

    conn.onclose = function (evt) {
        appendLog("<div><b>" + '\xa0\xa0' + "Connection closed.</b></div>");
    };
} else {
    appendLog("<div><b>" + '\xa0\xa0' + "Please Use a browser that supports WebSockets.</b></div>");
}

function leafInit(conn) {
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

    var blueIcon = new L.Icon({
        iconUrl: 'https://cdn.rawgit.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-red.png',
        shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png',
        iconSize: [25, 41],
        iconAnchor: [12, 41],
        popupAnchor: [1, -34],
        shadowSize: [41, 41]
    });

    leafDrawInit(leafMaps, blueIcon);
    leafDraw(leafMaps, conn, blueIcon);
}

function leafDrawInit(leafMaps, blueIcon) {
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
        markerStyle: {
            opacity: 0.5,
            draggable: true,
            // icon: blueIcon,
        },
        templineStyle: {
            color: '#3794ff',
        },
        hintlineStyle: {
            color: '#3794ff',
            dashArray: [5, 5],
        },
        pathOptions: {
            color: '#3794ff',
            fillColor: '#3794ff',
            fillOpacity: 0.3,
        },
    };
    leafMaps.pm.addControls(options);
    leafMaps.pm.enableDraw('Poly', options);
    leafMaps.pm.enableDraw('Marker', options);
    leafMaps.pm.disableDraw('Poly');
    leafMaps.pm.disableDraw('Marker');
}

var selectionToggle = false;

function leafDraw(leafMaps, conn, blueIcon) {

    var polygonStyle = {
        "stroke": true,
        "color": "#3794ff",
        "fillColor": "#3794ff",
        "weight": 3,
        "fillOpacity": 0.3
    };

    geojson = L.geoJSON(null, {
        pointToLayer: function (feature, latlng) {
            return L.marker(latlng, {
                // icon: blueIcon
            });
        },
        style: polygonStyle,
        onEachFeature: onEachFeature
    }).addTo(leafMaps);


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

        conn.addEventListener('message', function (evt) {
            var msgServer = JSON.parse(evt.data);
            if (!msgServer.features) {
                // It doesn't exist, do nothing
            } else {
                if (msgServer.features.properties.annotationStatus == "no error"){
                    var swapCoordinates = msgServer.features.geometry.coordinates;

                    if (msgServer.features.geometry.type == "Point") {
                        msgServer.features.geometry.coordinates = msgServer.features.properties.pixelCoordinates.Vertex1Array;
                        msgServer.features.properties.pixelCoordinates.Vertex1Array = swapCoordinates;
                    } else if (msgServer.features.geometry.type == "Polygon") {
                        msgServer.features.geometry.coordinates = msgServer.features.properties.pixelCoordinates.Vertex3Array;
                        msgServer.features.properties.pixelCoordinates.Vertex3Array = swapCoordinates;
                    }
                    geojson.addData(msgServer.features);
                }else if (msgServer.features.properties.annotationStatus == "no selection"){
                    appendLog("<div>" + '\xa0\xa0' + "> Please place marker on ground surfaces.</div>");
                }

                leafMaps.removeLayer(e.layer);
                this.removeEventListener('message', arguments.callee, false);
            }
        });

        var sent = toServer(message4Server, conn);
        if (!sent) {
            appendLog("<div><b>" + '\xa0\xa0' + "Message not sent.</b></div>");
        }
    });
    geojson.on("click", onFeatureGroupClick);
}

function onFeatureGroupClick(e) {
    var group = e.target;
    var layer = e.layer;

    var polygonStyle = {
        "stroke": true,
        "color": "#3794ff",
        "fillColor": "#3794ff",
        "weight": 3,
        "fillOpacity": 0.3
    };
    var polygonHLight = {
        "stroke": true,
        "color": "#3fdbff",
        "fillColor": "#3fdbff",
        "weight": 3,
        "fillOpacity": 0.5
    };
    if (layer._latlngs) {
        group.setStyle(polygonStyle);
        selectionToggle = !selectionToggle;

        if (selectionToggle == true) {
            layer.setStyle(polygonHLight);
        } else {
            layer.setStyle(polygonStyle);
            layer.closePopup();
        }
    }
}



function onEachFeature(feature, layer) {

    var annotationType = feature.properties.annotationType;
    layer.bindPopup('<div style="width:150px; text-align: left;">' + "This feature is a " +
    annotationType + '. Coordinates are in textbox. Click on the Google Maps Icon for More Options.</div>');

    var featureUUID = feature.properties.annotationID;
    layer.on('remove', function () {
        gMaps.data.forEach(function (feature) {
            if (feature.getProperty('annotationID') == featureUUID) {
                gMaps.data.remove(feature);
            }
        });
    });

    layer.on('popupopen', function () {
        
        gMaps.data.revertStyle();
        
        if (selectionToggle == false) {
            gMaps.data.forEach(function (feature) {
                if (feature.getProperty('annotationID') == featureUUID) {
                    gMaps.data.overrideStyle(feature, {
                        fillColor: '#3fdbff',
                        fillOpacity: 0.3,
                        strokeWeight: 3,
                        strokeColor: '#3fdbff'
                    });
                }
            });
        }
    });

    var clicked = false;
    layer.on('click', function(){
        clicked = !clicked;
        if (clicked){
            if(feature.geometry.type == "Point"){
                appendLog("<div>" + '\xa0\xa0' +
                "> Latitude: "+ feature.properties.pixelCoordinates.Vertex1Array[1] +
                 "<--> Longtitude: "+ feature.properties.pixelCoordinates.Vertex1Array[0] +"</div>");
            }else if (feature.geometry.type == "Polygon"){
    
                for (i =0; i< (feature.properties.pixelCoordinates.Vertex3Array[0]).length-1; i++) {
                    appendLog("<div>" + '\xa0\xa0' +
                    "> Latitude: "+ feature.properties.pixelCoordinates.Vertex3Array[0][i][1] +
                     "<--> Longtitude: "+ feature.properties.pixelCoordinates.Vertex3Array[0][i][0] +"</div>");
                }
    
            }
        }
    });
}

function toServer(message4Server, conn) {
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

        annotation = turf.point([0, 0], {
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
        annotation = turf.polygon([
            [
                [0, 0],
                [0, 0],
                [0, 0],
                [0, 0]
            ]
        ], {
            annotationType: annotationType,
            pixelCoordinates: [pixCoordinates1Array]
        });
        // }    
    }
    return annotation;
}

function appendLog(message) {
    var d = document.getElementById('log');
    var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
    d.innerHTML += message;
    if (doScroll) {
        d.scrollTop = d.scrollHeight - d.clientHeight;
    }
}

function stateChange(leafMaps, newState) {
    setTimeout(function () {
        leafMaps.removeLayer(newState);
    }, 1000);
}