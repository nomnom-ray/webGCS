package server

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nomnom-ray/webGCS/models"
	"github.com/nomnom-ray/webGCS/util"
)

type Message struct {
	lot *models.LotRaster
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
		err := wsConn.ReadJSON(&message.lot)
		if err != nil {
			break
		}
		messageProcessed, err := processing(message, projectedTile)
		if err != nil {
			return
		}

		c.h.broadcast <- messageProcessed
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

	var messageRX [][]int64
	var messageProcessed models.MessageProcessed
	var lots2Client []models.MessageProcessedLot

	if message.lot.LotX1 == 0 {
		messageRX = [][]int64{
			{message.lot.LotX0, message.lot.LotY0},
		}
		messageProcessed.MessageprocessedType = 0
	} else {
		messageRX = [][]int64{
			{message.lot.LotX0, message.lot.LotY0},
			{message.lot.LotX1, message.lot.LotY1},
			{message.lot.LotX2, message.lot.LotY2},
			{message.lot.LotX3, message.lot.LotY3},
		}
		messageProcessed.MessageprocessedType = 1
	}

	for _, message := range messageRX {

		//TODO: ok condition here for camera(); ok from whether primitive is selected
		lot2Client, err := camera(message, projectedTile)
		if err != nil {
			// fmt.Print(err)
			messageProcessed.MessageprocessedType = 9
		}

		lots2Client = append(lots2Client, lot2Client)
	}
	messageProcessed.Lots2Client = lots2Client

	if len(lots2Client) == 4 {
		err := PostNewFeatures(messageProcessed)
		if err != nil {
			return messageProcessed, err
		}
	}

	return messageProcessed, nil
}

func camera(message []int64, projectedTile *models.ProjectedTiles) (models.MessageProcessedLot, error) {

	//TODO:check: picked vertex put into mapvector; replace with interface?

	var lotCoordinates models.MapVector
	var lotRaster []int64
	var lot2Client models.MessageProcessedLot

	if _, vertexSelected, ok := projectedTile.RasterPicking(int(message[0]), int(message[1])); ok {
		// pretty.Println(primitiveSelected)
		// pretty.Println(vertexSelected)

		lotCoordinates.VertX = util.RoundToF7(vertexSelected.Position.X)
		lotCoordinates.VertY = util.RoundToF7(vertexSelected.Position.Y)
		lotCoordinates.VertZ = util.RoundToF7(vertexSelected.Position.Z)
		lotCoordinates.Latitude = util.RoundToF7(vertexSelected.Texture.X)
		lotCoordinates.Elevation = util.RoundToF7(vertexSelected.Texture.Y)
		lotCoordinates.Longtitude = util.RoundToF7(vertexSelected.Texture.Z)
		lotRaster = []int64{message[0], message[1]}
		lot2Client = models.MessageProcessedLot{MessageOriginal: lotRaster, Messageprocessed: lotCoordinates}

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
