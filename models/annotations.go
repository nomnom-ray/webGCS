package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/paulmach/orb/geojson"
)

//Annotation stores the IDs to map coordinates of features
type Annotation struct {
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
type Point struct {
	Vertex1Array []float64
}
type Line struct {
	Vertex2Array [][]float64
}
type Polygon struct {
	Vertex3Array [][][]float64
}

type GjsonGeometry struct {
	GeometryType string          `json:"type"`
	Coordinates  json.RawMessage `json:"coordinates"`
}

type Msg2Client struct {
	Feature *geojson.Feature `json:"features"`
}

//NewAnnotation constructor FUNCTION of Update struct
func NewAnnotation(annotation Msg2Client) (*Annotation, error) {
	// id, err := client.Incr("annotation:next-id").Result() //assign id to assignment to redis
	// if err != nil {
	// 	return nil, err
	// }
	// key := fmt.Sprintf("annotation:%d", id)

	// annotationBin, err := annotation.MarshalBinary()
	// if err != nil {
	// 	return nil, err
	// }

	// pipe := client.Pipeline()
	// pipe.HSet(key, "id", id)
	// pipe.HSet(key, "annotation", annotationBin)
	// pipe.RPush("annotationlist", id)

	// _, err = pipe.Exec()
	// if err != nil {
	// 	return nil, err
	// }

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

	// pretty.Println(annotation.Feature)

	rawjson, _ := annotation.Feature.MarshalJSON()

	// m := GeojsonFeatures{
	// 	FeatureType: "Feature",
	// 	Property: GjsonProperties{
	// 		AnnotationType: "Point",
	// 	},
	// 	Geometry: GjsonGeometry{
	// 		GeometryType: "Point",
	// 	},
	// }

	cmd := redis.NewStringCmd("SET", "fleet", "truck", "OBJECT", rawjson)

	tileClient.Process(cmd)
	v, _ := cmd.Result()
	log.Println(v)

	cmd1 := redis.NewStringCmd("GET", "fleet", "truck")
	tileClient.Process(cmd1)
	v1, _ := cmd1.Result()
	log.Println(v1)

	id := int64(0)
	return &Annotation{id}, nil
}

// queryUpdates is a helper function to get updates
func queryAnnotations(key string) ([]*Annotation, error) {
	annotationIDs, err := client.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	annotations := make([]*Annotation, len(annotationIDs)) //allocate memmory for update struct by each updateID
	for i, idString := range annotationIDs {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return nil, err
		}
		annotations[i] = &Annotation{id} //populate update memory space with update by key
	}

	return annotations, nil
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

//GetAnnotationCntxt gets the actual coordinates of the feature based on id
func (l *Annotation) GetAnnotationCntxt() (Msg2Client, error) {

	key := fmt.Sprintf("annotation:%d", l.FeatureID)

	var annotation Msg2Client
	cachedAnnotationBin, err := client.HGet(key, "annotation").Result()
	if err != nil {
		return annotation, err
	}
	err = annotation.UnmarshalBinary([]byte(cachedAnnotationBin))
	if err != nil {
		return annotation, err
	}
	return annotation, nil
}

//GetGlobalAnnotations gets all features in the loaded database (not user specific)
func GetGlobalAnnotations() ([]*Annotation, error) {
	return queryAnnotations("annotationlist")
}
