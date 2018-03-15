leafInit();


function leafInit() {
    var width = 600;
    var height = 600;
    var leafMaps = L.map('leafMaps', {
        minZoom: 1,
        maxZoom: 1,
        zoom: 1,
        zoomControl: false,
        center: [0, 0],
        crs: L.CRS.Simple
    });
    var northEast = leafMaps.unproject([width, 0], leafMaps.getMaxZoom());
    var southWest = leafMaps.unproject([0, height], leafMaps.getMaxZoom());
    var imageBounds = new L.LatLngBounds(southWest, northEast);
    var imageUrl = 'http://localhost:8080/templates/rBW-loc43_4516288_-80_4961367-fov80-heading205-pitch-10.jpg';

    leafMaps.setMaxBounds(new L.LatLngBounds(southWest, northEast));
    L.imageOverlay(imageUrl, imageBounds).addTo(leafMaps);
    leafMaps.setMaxBounds(imageBounds);

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

leafMaps.on('pm:drawstart', function(e) {
    if (e.shape==="Marker"){
        appendLog("<div>" + '\xa0\xa0' + "> Place marker to get a location coordinate.</div>");
    }else if(e.shape==="Poly"){
        appendLog("<div>" + '\xa0\xa0' + "> Draw polygon to get coordinates encapsulating a feature.</div>"); 
    }
});

leafMaps.on('pm:create', function(e) {
    
    var geometryType;
    var coordinates;
    var annotationType;

    if (e.shape === "Marker"){
        geometryType = e.shape;
        coordinates = e.layer._latlng;
        annotationType = "Point";        
    }else if (e.shape === "Poly"){
        geometryType = e.shape;
        coordinates = e.layer._latlngs;
        annotationType = "LotParking";    
    }

    // console.log(e.layer);

    createGEOJSON(geometryType,coordinates,annotationType);



    // if (!conn) {
    //     return false;
    // }
    // conn.send(message4Server);
});
}


function createGEOJSON(geometryType,coordinates,annotationType){

if (geometryType === "Marker"){

var point = turf.point([coordinates.lat, coordinates.lng],{ annotationType: annotationType });

    console.log(point);

}else if (e.shape === "Poly"){



}






        // console.log(e.layer._latlngs[0][0].lat);
    // console.log(e.layer._latlngs[0][0].lng);

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

function appendLog(message) {
    var d = document.getElementById('log');
    var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
    d.innerHTML += message;
    if (doScroll) {
        d.scrollTop = d.scrollHeight - d.clientHeight;
    }
}

if (window.WebSocket) {
    var conn;
    appendLog("<div>" + '\xa0\xa0' + "Connection Established.</div>");
    appendLog("<div>" + '\xa0\xa0' + 
    "> Use the palette tools to beginning annotation. " +
    "Deselect a tool by clicking it again. " + 
    "Delete an annotation by selecting the Bin tool and clicking on the annoation.</div>");
    conn = new WebSocket("ws://localhost:8080/ws");

    conn.addEventListener('message', function (e) {
        var msgServer = JSON.parse(e.data);
    });

    conn.onclose = function (evt) {
        appendLog("<div>" + '\xa0\xa0' + "Connection closed.</div>");
    };
} else {
    appendLog("<div><b>" + '\xa0\xa0' + "Your browser does not support WebSockets.</b></div>");
}