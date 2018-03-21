package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nomnom-ray/webGCS/models"
	"github.com/nomnom-ray/webGCS/util"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type MsgFromClient struct {
	feature *models.GeojsonFeatures
}

type Connection struct {
	// unBuffered channel of outbound messages.
	send chan models.Msg2Client
	// The hub.
	h *Hub
}

func (c *Connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn, projectedTile *models.ProjectedTiles) {
	defer wg.Done()
	//read message from clients
	for {
		var message MsgFromClient
		err := wsConn.ReadJSON(&message.feature)
		if err != nil {
			fmt.Println(err)
			break
		}

		msg2Clients, err := processing(message, projectedTile)
		if err != nil {
			return
		}

		c.h.broadcast <- msg2Clients
	}
}

func (c *Connection) writer(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	for message := range c.send {
		err := wsConn.WriteJSON(message)
		if err != nil {
			break
		}
	}
}

func processing(message MsgFromClient, projectedTile *models.ProjectedTiles) (models.Msg2Client, error) {

	p := &message.feature.Property
	g := &message.feature.Geometry
	var msg2Client models.Msg2Client
	var err error
	var messageRX [][]float64

	//TODO:take case of the panic cases to something that won't crash the program
	switch g.GeometryType {
	case "Point":
		err = json.Unmarshal(p.PixCoordinates, &p.Point.Vertex1Array)
		messageRX = append(messageRX, []float64{p.Point.Vertex1Array[0], p.Point.Vertex1Array[1]})
	case "LineString":
		err = json.Unmarshal(p.PixCoordinates, &p.Line.Vertex2Array)
	case "Polygon":
		err = json.Unmarshal(p.PixCoordinates, &p.Polygon.Vertex3Array)
		for _, vertex := range p.Polygon.Vertex3Array[0] {
			messageRX = append(messageRX, []float64{vertex[0], vertex[1]})
		}
	default:
		panic("Unknown type")
	}
	if err != nil {
		fmt.Printf("Failed to convert %s: %s", g.GeometryType, err)
	}

	var geoCoordinates [][]float64
	geoStatus := "no error"
	for _, pixCoor := range messageRX {
		geoCoordinate, err := camera(pixCoor, projectedTile)
		if err != nil {
			geoStatus = "no selection"
		}
		geoCoordinates = append(geoCoordinates, geoCoordinate)
	}

	annotationID, err := util.AnnotationID()
	if err != nil {
		return msg2Client, err
	}

	var points orb.Point
	var lines orb.LineString
	var rings orb.Ring
	var polygon orb.Polygon
	for _, geoCoordinate := range geoCoordinates {
		points = orb.Point{geoCoordinate[0], geoCoordinate[1]}
		lines = append(lines, points)
		rings = orb.Ring(lines)
	}
	polygon = append(polygon, rings)

	var feature2Clnts *geojson.Feature

	switch g.GeometryType {
	case "Point":
		feature2Clnts = geojson.NewFeature(points)
		feature2Clnts.Properties["pixelCoordinates"] = p.Point
		feature2Clnts.Properties["annotationType"] = p.AnnotationType
		feature2Clnts.Properties["annotationID"] = annotationID
		feature2Clnts.Properties["annotationStatus"] = geoStatus
	case "LineString":

	case "Polygon":
		feature2Clnts = geojson.NewFeature(polygon)
		feature2Clnts.Properties["pixelCoordinates"] = p.Polygon
		feature2Clnts.Properties["annotationType"] = p.AnnotationType
		feature2Clnts.Properties["annotationID"] = annotationID
		feature2Clnts.Properties["annotationStatus"] = geoStatus
	default:
		panic("Unknown type")
	}
	if err != nil {
		fmt.Printf("Failed to convert %s: %s", g.GeometryType, err)
	}

	msg2Client.Feature = feature2Clnts

	err = PostNewFeatures(msg2Client)
	if err != nil {
		return msg2Client, err
	}

	return msg2Client, nil
}

func camera(pixCoor []float64, projectedTile *models.ProjectedTiles) ([]float64, error) {

	geoCoordinate := make([]float64, 2)

	//returns primitiveSelected and vertexSelected
	if _, vertexSelected, ok := projectedTile.RasterPicking(int(pixCoor[0]), int(pixCoor[1])); ok {
		geoCoordinate[0] = util.RoundToF7(vertexSelected.Texture.Z)
		geoCoordinate[1] = util.RoundToF7(vertexSelected.Texture.X)
		// geoCoordinate.Elevation = util.RoundToF7(vertexSelected.Texture.Y)
	} else {
		err := errors.New("picking: primitive not selected")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}
	return geoCoordinate, nil
}

func (c *Connection) syncToDatabase(wsConn *websocket.Conn) error {

	var msg2Client models.Msg2Client
	annotations, err := models.GetGlobalAnnotations()
	if err != nil {
		util.InternalServerError(err, wsConn)
		return err
	}
	for _, annotation := range annotations {
		msg2Client, err = annotation.GetAnnotationCntxt()
		if err != nil {
			return err
		}

		c.h.connectionsMx.RLock()
		err = wsConn.WriteJSON(msg2Client)
		if err != nil {
			util.InternalServerError(err, wsConn)
			break
		}
		c.h.connectionsMx.RUnlock()
	}
	return nil
}

//PostNewFeatures pushes primitive types to the model for storing in redis
func PostNewFeatures(msg2DB models.Msg2Client) error {
	_, err := models.NewAnnotation(msg2DB)
	return err
}
