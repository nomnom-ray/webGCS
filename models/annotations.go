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

//NewAnnotation constructor
func NewAnnotation(annotation Msg2Client, userIDInSession int64) (*Annotation, error) {
	id, err := client.Incr("annotation:next-id").Result()
	if err != nil {
		return nil, err
	}
	redisKey := fmt.Sprintf("annotation:%d", id)
	tile38Key := annotation.Feature.Properties["annotationType"]
	annotationUUID := annotation.Feature.Properties["annotationID"]

	pipe := client.Pipeline()
	pipe.HSet(redisKey, "id", id)
	pipe.HSet(redisKey, "user_id", userIDInSession)
	pipe.HSet(redisKey, "gjson_uuid", annotationUUID)
	pipe.LPush("annotation_list", id)

	pipe.LPush(fmt.Sprintf("user:%d:annotation_list", userIDInSession), id)

	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	rawjson, _ := annotation.Feature.MarshalJSON()
	cmd := redis.NewStringCmd("SET", tile38Key, redisKey, "OBJECT", rawjson)
	tileClient.Process(cmd)
	v, _ := cmd.Result()
	log.Println(v)

	return &Annotation{id}, nil
}

// GetAnnotationUUIDs is a helper function to get IDs to tile38 json
func GetAnnotationUUIDs(annotationList string) ([]*Annotation, error) {
	annotationIDs, err := client.LRange(annotationList, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	annotationsArray := make([]*Annotation, len(annotationIDs))
	for i, idString := range annotationIDs {
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			// TODO:parse error probably
			return nil, err
		}
		annotationsArray[i] = &Annotation{id}
	}

	return annotationsArray, nil
}

// //GetAnnotationUUID gets the actual coordinates of the feature based on id
// func (l *Annotation) GetAnnotationUUID() (Msg2Client, error) {

// 	redisKey := fmt.Sprintf("annotation:%d", l.FeatureID)

// 	var annotation Msg2Client
// 	uuid, err := client.HGet(redisKey, "gjson_uuid").Result()
// 	if err != nil {
// 		return annotation, err
// 	}

// 	pretty.Println("uuid")

// 	pretty.Println(uuid)

// 	return annotation, nil
// }

//GetGlobalAnnotations gets all features in the loaded database (not user specific)
func GetGlobalAnnotations() ([]*Annotation, error) {
	return GetAnnotationUUIDs("annotationList")
}
