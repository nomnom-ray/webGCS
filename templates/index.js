var gMaps;
var conn;

var directionsDisplay;
var directionsService; 

var infowindow;

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

    infowindow = new google.maps.InfoWindow();

    gMaps.data.setStyle({
        fillColor: '#3794ff',
        fillOpacity: 0.3,
        strokeWeight: 1,
        strokeColor: '#3794ff',
        icon: '/templates/js-lib/leaflet-color-markers/img/marker-icon-blue.png'
    });

    directionsService = new google.maps.DirectionsService();
    var lineSymbol = {
        path: 'M 0,-1 0,1',
        strokeOpacity: 1,
        scale: 4
    };
    var dashedPolyline = {
        strokeOpacity: 0,
        icons: [{
          icon: lineSymbol,
          offset: '0',
          repeat: '20px'
        }]
    };
    directionsDisplay = new google.maps.DirectionsRenderer({polylineOptions:dashedPolyline});
    directionsDisplay.setMap(gMaps);

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
            gSelectedAnnotation = e.feature;
            var annotationType = e.feature.getProperty("annotationType");
            if (annotationType == "Point") {
                e.feature.getGeometry().forEachLatLng(function (path) {
                //     appendLog("<div>" + '\xa0\xa0' +
                //         "> Latitude: " + path.lat() + ";Longtitude: " + path.lng() + "</div>");
                });
                infowindow.setPosition(e.feature.getGeometry().get());
                infowindow.setOptions({
                    pixelOffset: new google.maps.Size(0, -30)
                });
            } else if (annotationType == "LotParking") {
                e.feature.getGeometry().forEachLatLng(function (path) {
                //     appendLog("<div>" + '\xa0\xa0' +
                //         ">Polygon corner " +y+ "<--> Latitude: " + path.lat() + "; Longtitude: " + path.lng() + "</div>");
                });
                infowindow.setPosition(e.latLng);
                infowindow.setOptions({
                    pixelOffset: new google.maps.Size(0, -5)
                });
            }
            if (annotationType =="Point"){
                annotationPosition=[e.feature.getGeometry().get().lat(),e.feature.getGeometry().get().lng()];
            }else{
                annotationPosition=[e.latLng.lat(),e.latLng.lng()];
            }
            infowindow.setContent(
                '<div style="width:150px; text-align: left;">' + "This feature is a " +
                annotationType + '. Evaluate its accuracy by turning on satellite view on the upper-right corner. "Navigate" creates route to a demo location "Dallas". </div><br/>' +
                '<button id="deleteButton" onclick="deleteAnnotation(gSelectedAnnotation,conn);">Delete</button>' +
                '<button id="navButton" onclick="getNavRoutes(annotationPosition,conn,infowindow)">Navigate</button>'
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

    if (window.WebSocket) {
        window.onbeforeunload = function () {
            conn.close();
        };
        appendLog("<div><b>" + '\xa0\xa0' + "Connection Established (Instructions Below).</b></div>");
        appendLog("<div>" + '\xa0\xa0' +
            "> The palette toolbar contains a marker and a polygon annotation tool. " +
            "Deselect a tool by selecting it again. " +
            "Delete an annotation by selecting the Bin tool and clicking on the annoation. " +
            "The annotations are selectable by clicking on them, " +
            "but please wait until the annotation syncs to Google Maps.</div>");

        appendLog("<div><b>" + '\xa0\xa0' + "> The sync is slow due to the use of a free slow server.</b></div>");

        // conn = new WebSocket("ws://localhost:8080/ws");
        conn = new WebSocket("ws://" + window.location.hostname + "/ws");
        conn.addEventListener('message', function (e) {
            var msgServer = JSON.parse(e.data);
            if (!msgServer.features) {
                // It doesn't exist, do nothing
            } else {
                if (msgServer.features.properties.annotationStatus == "no error") {
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
}

function deleteAnnotation(gSelectedAnnotation,conn) {
    var featureUUID = gSelectedAnnotation.getProperty("annotationID");
    geojson.eachLayer(function (layer) {
        // layer.feature is the original geojson feature
        if (layer.feature.properties.annotationID === featureUUID) {
            geojson.removeLayer(layer);
        }
    });
    gMaps.data.remove(gSelectedAnnotation);
    gSelectedAnnotation = 0;
}

var routePolygon = null;   
function getNavRoutes(annotationPosition,conn,infowindow) {    

    var start = new google.maps.LatLng(annotationPosition[0],annotationPosition[1]);
    var end = new google.maps.LatLng(43.451883, -80.495026);
    var request = {
        origin: start,
        destination: end,
        travelMode: google.maps.DirectionsTravelMode.WALKING
    };

    directionsService.route(request, function (response, status) {
        if (status == google.maps.DirectionsStatus.OK) {
            // directionsDisplay.setDirections(response);
            var navFrame = 0.007 / 111.12;
            var overviewPath = response.routes[0].overview_path;
            var geoInput = googleMaps2JTS(overviewPath);
            var geometryFactory = new jsts.geom.GeometryFactory();
            var shell = geometryFactory.createLineString(geoInput);
            var polygon = shell.buffer(navFrame);

            var navFrameCoor = [];
            for (var latlng in polygon._shell._points._coordinates){
                navFrameCoor.push([polygon._shell._points._coordinates[latlng].y,polygon._shell._points._coordinates[latlng].x]);
            }
            //creating navigation frame to receive 3D tile data; push to later date
            // var messageType = "navigationFrame";
            // var messageContent = {geometry:{type:"Polygon",coordinates:[navFrameCoor]}};
            // var message4Server = JSON.stringify({messagetype:messageType,messagecontent:messageContent});

            // var sent = toServer(message4Server, conn);
            // if (!sent) {
            //     appendLog("<div><b>" + '\xa0\xa0' + "Message not sent.</b></div>");
            // }
            
            var oLanLng = [];
            var oCoordinates;
            oCoordinates = polygon._shell._points._coordinates[0];
            for (i = 0; i < oCoordinates.length; i++) {
                var oItem;
                oItem = oCoordinates[i];
                oLanLng.push(new google.maps.LatLng(oItem[1], oItem[0]));
            }
            if (routePolygon && routePolygon.setMap) routePolygon.setMap(null);
            routePolygon = new google.maps.Polygon({
                paths: jsts2googleMaps(polygon),
                map: gMaps
            });
            infowindow.close(gMaps);

            var infoWindow = new google.maps.InfoWindow();
            google.maps.event.addListener(routePolygon, 'click', function (event) {
                var contentString = 'Minimal sever/client<br>' + 
                'data exchange for<br>' + 
                'autonomous navigation.<br>'+
                '<b>Route Coordinates:</b><br>Lat: ' + 
                precisionRound(event.latLng.lat(),6) +
                '<br>Lng: ' + precisionRound(event.latLng.lng(),6) + '<br/>' +
                '<button id="deleteButton" onclick="deleteRoute(routePolygon);">Delete</button>';

                infoWindow.setContent(contentString);
                infoWindow.setPosition(event.latLng);
                infoWindow.open(gMaps, routePolygon);
            });
        }
    });
}

function deleteRoute(routePolygon){
routePolygon.setMap(null);
}

var jsts2googleMaps = function (geometry) {
    var coordArray = geometry.getCoordinates();
    GMcoords = [];
    for (var i = 0; i < coordArray.length; i++) {
      GMcoords.push(new google.maps.LatLng(coordArray[i].x, coordArray[i].y));
    }
    return GMcoords;
  };

function googleMaps2JTS(boundaries) {
    var coordinates = [];
    var length = 0;
    if (boundaries && boundaries.getLength) length = boundaries.getLength();
    else if (boundaries && boundaries.length) length = boundaries.length;
    for (var i = 0; i < length; i++) {
        if (boundaries.getLength) coordinates.push(new jsts.geom.Coordinate(
        boundaries.getAt(i).lat(), boundaries.getAt(i).lng()));
        else if (boundaries.length) coordinates.push(new jsts.geom.Coordinate(
        boundaries[i].lat(), boundaries[i].lng()));
    }
    return coordinates;
}

function leafInit(conn) {

    var width = 600;
    var height = 600;
    var leafMaps = L.map('leafMaps', {
        minZoom: 0,
        maxZoom: 0,
        zoom: 0,
        zoomControl: false,
        center: [300, 300],
        crs: L.CRS.Simple
    });
    var northEast = leafMaps.unproject([0, width]);
    var southWest = leafMaps.unproject([height, 0]);
    var imageBounds = new L.LatLngBounds(southWest, northEast);
    var imageUrl = '/templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';


    leafMaps.setMaxBounds(imageBounds);
    L.imageOverlay(imageUrl, imageBounds).addTo(leafMaps);

    leafMaps.dragging.disable();

    var blueIcon = new L.Icon({
        iconUrl: '/templates/js-lib/leaflet-color-markers/img/marker-icon-2x-blue.png',
        shadowUrl: '/templates/js-lib/leaflet-color-markers/img/marker-shadow.png',
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
            icon: blueIcon,
        },
        templineStyle: {
            color: '#3794ff',
            opacity: 0.5,
        },
        hintlineStyle: {
            color: '#3794ff',
            dashArray: [5, 5],
            opacity: 0.5,
        },
        pathOptions: {
            color: '#3794ff',
            opacity: 0.5,
            fillColor: '#3794ff',
            fillOpacity: 0.2,
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
                icon: blueIcon
            });
        },
        style: polygonStyle,
        onEachFeature: onEachFeature
    }).addTo(leafMaps);


    leafMaps.on('pm:drawstart', function (e) {
        if (e.shape === "Marker") {
            appendLog("<div><b>" + '\xa0\xa0' + "> Place marker to get a location coordinate.</b></div>");
        } else if (e.shape === "Poly") {
            appendLog("<div><b>" + '\xa0\xa0' + "> click the 4 corners of a parking space to encapsulating it.</b></div>");
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
        var messageContent = createGEOJSON(geometryType, pixCoordinates, annotationType);
        var messageType = "annotationStore";

        message4Server = JSON.stringify({messagetype:messageType,messagecontent:messageContent});
        conn.addEventListener('message', function (evt) {
            var msgServer = JSON.parse(evt.data);
            if (!msgServer.features) {
                // It doesn't exist, do nothing
            } else {
                leafMaps.removeLayer(e.layer);
                this.removeEventListener('message', arguments.callee, false);
            }
        });

        var sent = toServer(message4Server, conn);
        if (!sent) {
            appendLog("<div><b>" + '\xa0\xa0' + "Message not sent.</b></div>");
        }
    });

    conn.addEventListener('message', function (event) {
        var msgServer = JSON.parse(event.data);
        if (!msgServer.features) {
            // It doesn't exist, do nothing
        } else {
            if (msgServer.features.properties.annotationStatus == "no error") {
                var swapCoordinates = msgServer.features.geometry.coordinates;

                if (msgServer.features.geometry.type == "Point") {
                    msgServer.features.geometry.coordinates = msgServer.features.properties.pixelCoordinates.Vertex1Array;
                    msgServer.features.properties.pixelCoordinates.Vertex1Array = swapCoordinates;
                } else if (msgServer.features.geometry.type == "Polygon") {
                    msgServer.features.geometry.coordinates = msgServer.features.properties.pixelCoordinates.Vertex3Array;
                    msgServer.features.properties.pixelCoordinates.Vertex3Array = swapCoordinates;
                }
                geojson.addData(msgServer.features);
            } else if (msgServer.features.properties.annotationStatus == "no selection") {
                appendLog("<div><b>" + '\xa0\xa0' + "> Please place marker on ground surfaces.</b></div>");
            }
            // this.removeEventListener('message', arguments.callee, false);
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
        annotationType + '. Coordinates are in textbox. Click on the annotation on Google Maps to see more.</div>');

    var featureUUID = feature.properties.annotationID;
    layer.on('remove', function () {
        gMaps.data.forEach(function (feature) {
            if (feature.getProperty('annotationID') == featureUUID) {

                if (annotationType != "Point"){
                    var messageType = "annotationRemove";
                    var messageContent = {properties:{annotationID:featureUUID,annotationType:annotationType}};
                    var message4Server = JSON.stringify({messagetype:messageType,messagecontent:messageContent});
                    var sent = toServer(message4Server, conn);
                    if (!sent) {
                        appendLog("<div><b>" + '\xa0\xa0' + "Message not sent.</b></div>");
                    }
                }
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
    layer.on('click', function () {
        clicked = !clicked;
        if (clicked) {
            if (feature.geometry.type == "Point") {
                appendLog("<div>" + '\xa0\xa0' +
                    "> Latitude: " + feature.properties.pixelCoordinates.Vertex1Array[1] +
                    "; Longtitude: " + feature.properties.pixelCoordinates.Vertex1Array[0] + "</div>");
            } else if (feature.geometry.type == "Polygon") {

                for (i = 0; i < (feature.properties.pixelCoordinates.Vertex3Array[0]).length - 1; i++) {
                    var y=i+1;
                    appendLog("<div>" + '\xa0\xa0' + 
                        "> Polygon corner " +y+ "<--> Latitude: " + feature.properties.pixelCoordinates.Vertex3Array[0][i][1] +
                        "; Longtitude: " + feature.properties.pixelCoordinates.Vertex3Array[0][i][0] + "</div>");
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

function precisionRound(number, precision) {
    var factor = Math.pow(10, precision);
    return Math.round(number * factor) / factor;
  }