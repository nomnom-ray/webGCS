package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nomnom-ray/webGCS/models"
	"github.com/nomnom-ray/webGCS/util"
)

type Message struct {
	feature *models.GeojsonFeatures
}

type Connection struct {
	// unBuffered channel of outbound messages.
	send chan models.MessageProcessed
	// The hub.
	h *Hub
}

func (c *Connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn, projectedTile *models.ProjectedTiles) {
	defer wg.Done()
	//read message from clients
	for {
		var message Message
		err := wsConn.ReadJSON(&message.feature)
		if err != nil {
			fmt.Println(err)
			break
		}

		_, err = processing(message, projectedTile)
		if err != nil {
			return
		}

		// c.h.broadcast <- messageProcessed
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

func processing(message Message, projectedTile *models.ProjectedTiles) (models.MessageProcessed, error) {

	g := &message.feature.Geometry
	var err error
	var messageRX [][]float64

	switch g.GeometryType {
	case "Point":
		err = json.Unmarshal(g.Coordinates, &g.Point.Vertex1Array)
		messageRX = append(messageRX, []float64{g.Point.Vertex1Array[0], g.Point.Vertex1Array[1]})
	case "LineString":
		err = json.Unmarshal(g.Coordinates, &g.Line.Vertex2Array)
	case "Polygon":
		err = json.Unmarshal(g.Coordinates, &g.Polygon.Vertex3Array)
		for _, vertex := range g.Polygon.Vertex3Array[0] {
			messageRX = append(messageRX, []float64{vertex[0], vertex[1]})
		}
	default:
		panic("Unknown type")
	}
	if err != nil {
		fmt.Printf("Failed to convert %s: %s", g.GeometryType, err)
	}

	var messageProcessed models.MessageProcessed

	// for _, message := range messageRX {

	// //TODO: ok condition here for camera(); ok from whether primitive is selected
	// message4Client, err := camera(message, projectedTile)
	// if err != nil {
	// 	// fmt.Print(err)
	// 	messageProcessed.MessageprocessedType = 9
	// }

	// 	lots2Client = append(lots2Client, message4Client)
	// }
	// messageProcessed.Lots2Client = lots2Client

	// if len(lots2Client) == 4 {
	// 	err := PostNewFeatures(messageProcessed)
	// 	if err != nil {
	// 		return messageProcessed, err
	// 	}
	// }

	return messageProcessed, nil
}

func camera(message []float64, projectedTile *models.ProjectedTiles) (models.MessageProcessedLot, error) {

	//TODO:check: picked vertex put into mapvector; replace with interface?

	var featureCoordinates models.MapVector
	var lotRaster []int64
	var lot2Client models.MessageProcessedLot

	if _, vertexSelected, ok := projectedTile.RasterPicking(int(message[0]), int(message[1])); ok {
		// pretty.Println(primitiveSelected)
		// pretty.Println(vertexSelected)

		featureCoordinates.VertX = util.RoundToF7(vertexSelected.Position.X)
		featureCoordinates.VertY = util.RoundToF7(vertexSelected.Position.Y)
		featureCoordinates.VertZ = util.RoundToF7(vertexSelected.Position.Z)
		featureCoordinates.Latitude = util.RoundToF7(vertexSelected.Texture.X)
		featureCoordinates.Elevation = util.RoundToF7(vertexSelected.Texture.Y)
		featureCoordinates.Longtitude = util.RoundToF7(vertexSelected.Texture.Z)
		lotRaster = []int64{int64(message[0]), int64(message[1])}

		lot2Client = models.MessageProcessedLot{MessageOriginal: lotRaster, Messageprocessed: featureCoordinates}

	} else {
		err := errors.New("picking: primitive not selected")
		if err != nil {
			return lot2Client, err
		}
	}

	return lot2Client, nil
}

func (c *Connection) syncToDatabase(wsConn *websocket.Conn) error {

	var messageProcessed models.MessageProcessed
	lotParkings, err := models.GetGlobalLotParkings()
	if err != nil {
		util.InternalServerError(err, wsConn)
		return err
	}
	for _, lotParking := range lotParkings {
		messageProcessed, err = lotParking.GetLotSpace()
		if err != nil {
			return err
		}

		c.h.connectionsMx.RLock()
		err = wsConn.WriteJSON(messageProcessed)
		if err != nil {
			util.InternalServerError(err, wsConn)
			break
		}
		c.h.connectionsMx.RUnlock()
	}
	return nil
}

//PostNewFeatures pushes primitive types to the model for storing in redis
func PostNewFeatures(messageProcessed models.MessageProcessed) error {

	lot := messageProcessed
	_, err := models.NewLotParking(lot)
	return err
}
