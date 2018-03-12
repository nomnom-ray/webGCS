package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

//LotParking stores the IDs to map coordinates of Lot parking space features
type LotParking struct {
	FeatureID int64
}

type LotRaster struct {
	LotX0 int64 `json:"lotX0"`
	LotY0 int64 `json:"lotY0"`
	LotX1 int64 `json:"lotX1"`
	LotY1 int64 `json:"lotY1"`
	LotX2 int64 `json:"lotX2"`
	LotY2 int64 `json:"lotY2"`
	LotX3 int64 `json:"lotX3"`
	LotY3 int64 `json:"lotY3"`
}

type MessageProcessed struct {
	MessageprocessedType int                   `json:"messageprocessedtype"`
	Lots2Client          []MessageProcessedLot `json:"lot2client"`
}

type MessageProcessedLot struct {
	MessageOriginal  []int64   `json:"messageprocessedoriginal"`
	Messageprocessed MapVector `json:"messageprocessedvectors"`
}

//NewLotParking constructor FUNCTION of Update struct
func NewLotParking(lotParking MessageProcessed) (*LotParking, error) {
	id, err := Client.Incr("lotparking:next-id").Result() //assign id to assignment to redis
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("lotparking:%d", id)

	lotParkingBin, err := lotParking.MarshalBinary()
	if err != nil {
		return nil, err
	}

	pipe := Client.Pipeline()
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
	LotParkingIDs, err := Client.LRange(key, 0, -1).Result()
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
func (m *MessageProcessed) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary -
func (m *MessageProcessed) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}

//GetLotSpace gets the actual coordinates of the lot based on id
func (l *LotParking) GetLotSpace() (MessageProcessed, error) {

	key := fmt.Sprintf("lotparking:%d", l.FeatureID)

	var lotParking MessageProcessed
	cachedLotParkingBin, err := Client.HGet(key, "lotparking").Result()
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
