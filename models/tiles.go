package models

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// //Annotation stores the IDs to map coordinates of features
// type PrimitivesInTile struct {
// 	PrimitiveID int64
// }

func (m *MapTiles) NewRedisTile() {
	// id, err := client.Incr("tile:next-id").Result()
	// if err != nil {
	// 	return nil, err
	// }
	// tileKey := fmt.Sprintf("tile:%d", id)

	// var feature2Redis *geojson.Feature

	for id, x := range m.PrimitiveIndex[0:10] {
		// triangle := [4][2]float64{
		// 	[2]float64{m.Vectors[x.PrimitiveBottom].Latitude, m.Vectors[x.PrimitiveBottom].Longtitude},
		// 	[2]float64{m.Vectors[x.PrimitiveTop].Latitude, m.Vectors[x.PrimitiveTop].Longtitude},
		// 	[2]float64{m.Vectors[x.PrimitiveLeft].Latitude, m.Vectors[x.PrimitiveLeft].Longtitude},
		// 	[2]float64{m.Vectors[x.PrimitiveBottom].Latitude, m.Vectors[x.PrimitiveBottom].Longtitude},
		// }
		// pretty.Println("triangles: ", triangle)

		vertexBottom := orb.Point{m.Vectors[x.PrimitiveBottom].Longtitude, m.Vectors[x.PrimitiveBottom].Latitude}
		vertexTop := orb.Point{m.Vectors[x.PrimitiveTop].Longtitude, m.Vectors[x.PrimitiveTop].Latitude}
		vertexLeft := orb.Point{m.Vectors[x.PrimitiveLeft].Longtitude, m.Vectors[x.PrimitiveLeft].Latitude}
		rings := orb.Ring([]orb.Point{vertexBottom, vertexTop, vertexLeft, vertexBottom})
		polygon := orb.Polygon{rings}
		feature2Redis := geojson.NewFeature(polygon)

		rawjson, err := feature2Redis.MarshalJSON()
		if err != nil {
			log.Println("error writing tile to redis")
			break
		}
		cmd := redis.NewStringCmd("SET", "tileKey", id, "OBJECT", rawjson)
		tileClient.Process(cmd)
	}
}

func TileCheck() bool {
	cmd := redis.NewStringCmd("GET", "tileKey", "1")
	tileClient.Process(cmd)
	tileResult, err := cmd.Result()
	if err != nil {
		log.Println(err)
	}
	if tileResult == "" {
		return false
	}
	return true

}

func GetNavPrimitives(MsgContent *GeojsonFeatures) (*geojson.Feature, error) {

	var navFrame *geojson.Feature
	var polygon Polygon
	err := json.Unmarshal(MsgContent.Geometry.Coordinates, &polygon.Vertex3Array)
	if err != nil {
		log.Println("couldn't unmarshal navigation frame")
		return navFrame, err
	}

	var lines []orb.Point

	for _, x := range polygon.Vertex3Array[0] {
		vertex := orb.Point{x[0], x[1]}
		lines = append(lines, vertex)
	}

	rings := orb.Ring(lines)
	polygon4Orb := orb.Polygon{rings}
	navFrame = geojson.NewFeature(polygon4Orb)

	//marshal navFrame into json to usef navFrame for intersect

	//replace with OBJECT
	cmd := redis.NewStringCmd("INTERSECTS", "tileKey", "OBJECTS", "BOUNDS", 43.451000, -80.497000, 43.452100, -80.494500)
	// cmd := redis.NewStringCmd("INTERSECTS", "tileKey", "OBJECT", navFrame)
	tileClient.Process(redis.NewStringCmd("OUTPUT", "json"))
	tileClient.Process(cmd)
	v, err := cmd.Result()
	if err != nil {
		log.Println(err)
	}
	log.Println(v)

	//output json from tile38 is json; needs conversion to *geojson.Feature to send the data around

	//output json from tile38 is nested array; conflict! data sent to client as *geojson.Feature
	//would be to redo the whole thing to using []*geojson.Feature

	return navFrame, nil

}
