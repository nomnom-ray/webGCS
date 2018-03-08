package server

import (
	"bufio"
	"encoding/csv"
	"fmt"
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
	PixelX1 int64 `json:"pixelX1"`
	PixelY1 int64 `json:"pixelY1"`
}

type MessageProcessed struct {
	Messageprocessed string `json:"messageprocessed"`
}

type Connection struct {
	// unBuffered channel of outbound messages.
	send chan MessageProcessed
	// The hub.
	h *Hub
}

func (c *Connection) reader(wg *sync.WaitGroup, wsConn *websocket.Conn) {
	defer wg.Done()

	//read message from clients
	for {
		var message Message
		err := wsConn.ReadJSON(&message)
		if err != nil {
			break
		}
		var messageProcessed MessageProcessed
		messageProcessed.Messageprocessed = processing(message)
		err = PostNewFeatures(messageProcessed)
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

func processing(message Message) string {

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
	var messageString string
	cameraPerspective := models.CameraModel(maxVert, cameraLocation)
	//3D-2D conversion
	triangles, primitiveOnScreen := models.Projection(maxVert, cameraPerspective)

	if primitiveSelected, vertexSelected, ok := models.RasterPicking(int(message.PixelX1), int(message.PixelY1), triangles, primitiveOnScreen, cameraPerspective); ok {
		pretty.Println(primitiveSelected)
		pretty.Println(vertexSelected)
		messageString = fmt.Sprintf("%s%d%s%d%s%.7f%s%.7f%s%.7f",
			"Pixel raster: X: ", int(message.PixelX1), "  Y:", int(message.PixelY1),
			" <===> Pixel GCS: Latitude:", vertexSelected.Texture.X, "  Elevation:", vertexSelected.Texture.Y, "  Lontitude:", vertexSelected.Texture.Z)

	} else {
		pretty.Println("picking: primitive not selected.")
		messageString = "picking: primitive not selected."
	}

	return messageString
}

func (c *Connection) syncToDatabase(wsConn *websocket.Conn) {

	var messageProcessed MessageProcessed
	lotParkings, err := models.GetGlobalLotParkings()
	if err != nil {
		util.InternalServerError(err, wsConn)
		return
	}
	for _, lotParking := range lotParkings {
		messageProcessed.Messageprocessed = lotParking.GetLotSpace()
		c.h.connectionsMx.RLock()
		err = wsConn.WriteJSON(messageProcessed)
		if err != nil {
			break
		}
		c.h.connectionsMx.RUnlock()
	}

}

//PostNewFeatures pushes primitive types to the model for storing in redis
func PostNewFeatures(messageProcessed MessageProcessed) error {

	lot := messageProcessed.Messageprocessed
	_, err := models.NewLotParking(lot)
	return err
}
