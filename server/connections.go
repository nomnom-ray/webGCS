package server

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
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

func (c *Connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()
	//read message from clients
	for {
		var message Message
		err := wsConn.ReadJSON(&message.lot)
		if err != nil {
			break
		}
		messageProcessed, err := processing(message)
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

func processing(message Message) (models.MessageProcessed, error) {

	var messageRX [][]int64
	var messageProcessed models.MessageProcessed
	var lot2client []models.MessageProcessedLot

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
		lot2client = append(lot2client, camera(message))
	}
	messageProcessed.Lot2Client = lot2client

	if len(lot2client) == 4 {
		err := PostNewFeatures(messageProcessed)
		if err != nil {
			return messageProcessed, err
		}
	}

	return messageProcessed, nil
}

func camera(message []int64) models.MessageProcessedLot {

	clientsFile, err := os.Open("resultNormModelProperties.csv")
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	propertiesReader := csv.NewReader(bufio.NewReader(clientsFile))

	var maxVert float64

	for i := 0; i < 7; i++ {
		property, error := propertiesReader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		maxVert, err = strconv.ParseFloat(property[3], 64)
		if err != nil {
			panic(err)
		}
	}

	//find camera location in GCS
	cameraLatitude := 43.4516288
	cameraLongtitude := -80.4961367
	cameraElevation := 0.0000113308
	cameraLocation := &models.MapVector{
		VertX:      0,
		VertY:      0,
		VertZ:      0,
		Latitude:   cameraLatitude,
		Longtitude: cameraLongtitude,
		Elevation:  cameraElevation,
	}
	cameraLocation = models.NormCameraLocation(cameraLocation)
	cameraPerspective := models.CameraModel(maxVert, cameraLocation)
	//3D-2D conversion
	triangles, primitiveOnScreen := models.Projection(maxVert, cameraPerspective)

	//TODO:check: picked vertex put into mapvector; replace with interface?

	var lotCoordinates models.MapVector
	var lotRaster []int64

	if primitiveSelected, vertexSelected, ok := models.RasterPicking(int(message[0]), int(message[1]), triangles, primitiveOnScreen, cameraPerspective); ok {
		pretty.Println(primitiveSelected)
		pretty.Println(vertexSelected)

		lotCoordinates.VertX = util.RoundToF7(vertexSelected.Position.X)
		lotCoordinates.VertY = util.RoundToF7(vertexSelected.Position.Y)
		lotCoordinates.VertZ = util.RoundToF7(vertexSelected.Position.Z)
		lotCoordinates.Latitude = util.RoundToF7(vertexSelected.Texture.X)
		lotCoordinates.Elevation = util.RoundToF7(vertexSelected.Texture.Y)
		lotCoordinates.Longtitude = util.RoundToF7(vertexSelected.Texture.Z)
		lotRaster = []int64{message[0], message[1]}

	} else {
		pretty.Println("picking: primitive not selected.")

	}

	lot2clients := models.MessageProcessedLot{MessageOriginal: lotRaster, Messageprocessed: lotCoordinates}

	return lot2clients
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
