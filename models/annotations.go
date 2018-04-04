package models

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nomnom-ray/webGCS/util"
	"github.com/satori/go.uuid"

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
	AnnotationID   string          `json:"annotationID"`
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
	MsgType string           `json:"msgTypeFromServer"`
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
	annotationUUID := annotation.Feature.Properties["annotationID"].(uuid.UUID).String()

	pipe := client.Pipeline()
	pipe.HSet(redisKey, "id", id)
	pipe.HSet(redisKey, "user_id", userIDInSession) //userIDInSession is user.id; not name; need user:+id string
	pipe.HSet(redisKey, "type", tile38Key)
	pipe.HSet("id:by-uuid", annotationUUID, id)

	pipe.LPush("annotation_list", id)
	pipe.LPush(fmt.Sprintf("user:%d:annotation_list", userIDInSession), id)

	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	rawjson, _ := annotation.Feature.MarshalJSON()
	cmd := redis.NewStringCmd("SET", tile38Key, id, "OBJECT", rawjson)
	tileClient.Process(cmd)
	// v, _ := cmd.Result()
	// log.Println(v)

	return &Annotation{id}, nil
}

func Remove1Annotation(property GjsonProperties, userIDInSession int64) error {

	idFromUUID, err := GetIDbyUUID(property.AnnotationID)
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("annotation:%d", idFromUUID)
	tile38Key := property.AnnotationType

	keyDelCheck := client.Del(redisKey).Val()
	if keyDelCheck < 1 {
		return util.ErrDeleting
	}

	globalListDelCheck := client.LRem("annotation_list", 1, idFromUUID).Val()
	if globalListDelCheck < 1 {
		return util.ErrDeleting
	}

	//TODO: figure out user annotation categorization; having this gives error rigth now
	// userListDelCheck := client.LRem(fmt.Sprintf("user:%d:annotation_list", userIDInSession), 1, idFromUUID).Val()
	// if userListDelCheck < 1 {
	// 	return util.ErrDeleting
	// }

	cmd := redis.NewStringCmd("DEL", tile38Key, idFromUUID)
	tileClient.Process(cmd)

	return nil
}

func GetIDbyUUID(annotationID string) (int64, error) {

	id, err := client.HGet("id:by-uuid", annotationID).Int64()
	if err == redis.Nil {
		return id, util.ErrAnnotationNotFound
	} else if err != nil {
		return id, err
	}
	return id, err
}

// GetAnnotationIDs is a helper function to get IDs to tile38 json
func GetAnnotationIDs(annotationList string) ([]*Annotation, error) {
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

//GetAnnotationContext gets the actual coordinates of the feature based on id
func (l *Annotation) GetAnnotationContext() (Msg2Client, error) {

	var annotation Msg2Client
	redisKey := fmt.Sprintf("annotation:%d", l.FeatureID)
	// uuid, err := client.HGet(redisKey, "gjson_uuid").Bytes() //.Byte() is a go-redis method
	tile38Key, err := client.HGet(redisKey, "type").Result()
	if err != nil {
		return annotation, err
	}
	cmd := redis.NewStringCmd("GET", tile38Key, l.FeatureID, "OBJECT")
	tileClient.Process(redis.NewStringCmd("OUTPUT", "resp"))
	tileClient.Process(cmd)
	rawjson, err := cmd.Result()
	if err != nil {
		return annotation, err
	}
	err = json.Unmarshal([]byte(rawjson), &annotation.Feature)
	if err != nil {
		return annotation, err
	}

	return annotation, nil
}

//GetGlobalAnnotations gets all features in the loaded database (not user specific)
func GetGlobalAnnotations() ([]*Annotation, error) {
	return GetAnnotationIDs("annotation_list")
}
