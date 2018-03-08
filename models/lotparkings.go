package models

import (
	"fmt"
	"strconv"
)

//LotParking stores the IDs to map coordinates of Lot parking space features
type LotParking struct {
	FeatureID int64
}

//NewLotParking constructor FUNCTION of Update struct
func NewLotParking(lot string) (*LotParking, error) {
	id, err := Client.Incr("lotparking:next-id").Result() //assign id to assignment to redis
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("lotparking:%d", id)

	pipe := Client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "space", lot)
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

//GetLotSpace gets the actual coordinates of the lot based on id
func (l *LotParking) GetLotSpace() string {

	key := fmt.Sprintf("lotparking:%d", l.FeatureID)
	feature, err := Client.HGet(key, "space").Result()
	if err != nil {
		panic(err)
	}

	return feature
}

//GetGlobalLotParkings gets all lots spaces in the loaded database (not user specific)
func GetGlobalLotParkings() ([]*LotParking, error) {
	return queryLotParkings("lotparkinglist")
}
