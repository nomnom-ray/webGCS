package models

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/paulmach/orb/geojson"
)

//LotParking stores the IDs to map coordinates of Lot parking space features
type LotParking struct {
	FeatureID int64
}

type GeojsonFeatures struct {
	FeatureType string          `json:"type"`
	Property    GjsonProperties `json:"properties"`
	Geometry    GjsonGeometry   `json:"geometry"`
}

type GjsonProperties struct {
	AnnotationType string          `json:"annotationType"`
	PixCoordinates json.RawMessage `json:"pixelCoordinates"`
	Point          Point
	Line           Line
	Polygon        Polygon
}

type GjsonGeometry struct {
	GeometryType string          `json:"type"`
	Coordinates  json.RawMessage `json:"coordinates"`
}

type Point struct {
	Vertex1Array []float64
}

type Line struct {
	Vertex2Array [][]float64
}

type Polygon struct {
	Vertex3Array [][][]float64
}

// Gjsn2Clnt
type Msg2Client struct {
	Feature *geojson.Feature `json:"features"`
}

//NewLotParking constructor FUNCTION of Update struct
func NewLotParking(lotParking Msg2Client) (*LotParking, error) {
	id, err := client.Incr("lotparking:next-id").Result() //assign id to assignment to redis
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("lotparking:%d", id)

	//TODO:tile38 test
	//All Tile38 commands should use redis.NewStringCmd(...args) to orgnize parameters,
	//then, use Process() to request result ,and get result by using Result().
	// cmd := redis.NewStringCmd("SET", "fleet", "truck", "POINT", 33.32, 115.423)
	// tileClient.Process(cmd)
	// v, _ := cmd.Result()
	// log.Println(v)
	// cmd1 := redis.NewStringCmd("GET", "fleet", "truck")
	// tileClient.Process(cmd1)
	// v1, _ := cmd1.Result()
	// log.Println(v1)

	lotParkingBin, err := lotParking.MarshalBinary()
	if err != nil {
		return nil, err
	}

	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "lotparking", lotParkingBin)
	pipe.RPush("lotparkinglist", id)

	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &LotParking{id}, nil
}

// queryUpdates is a helper function to get updates
func queryLotParkings(key string) ([]*LotParking, error) {
	LotParkingIDs, err := client.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	lotParkings := make([]*LotParking, len(LotParkingIDs)) //allocate memmory for update struct by each updateID
	for i, idString := range LotParkingIDs {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return nil, err
		}
		lotParkings[i] = &LotParking{id} //populate update memory space with update by key
	}

	return lotParkings, nil
}

// MarshalBinary -
func (m *Msg2Client) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary -
func (m *Msg2Client) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}

//GetLotSpace gets the actual coordinates of the lot based on id
func (l *LotParking) GetLotSpace() (Msg2Client, error) {

	key := fmt.Sprintf("lotparking:%d", l.FeatureID)

	var lotParking Msg2Client
	cachedLotParkingBin, err := client.HGet(key, "lotparking").Result()
	if err != nil {
		return lotParking, err
	}
	err = lotParking.UnmarshalBinary([]byte(cachedLotParkingBin))
	if err != nil {
		return lotParking, err
	}
	return lotParking, nil
}

//GetGlobalLotParkings gets all lots spaces in the loaded database (not user specific)
func GetGlobalLotParkings() ([]*LotParking, error) {
	return queryLotParkings("lotparkinglist")
}
