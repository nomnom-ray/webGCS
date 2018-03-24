package models

import (
	"context"
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/gocarina/gocsv"
	"github.com/kr/pretty"
	"github.com/nomnom-ray/webGCS/util"
	"googlemaps.github.io/maps"
	pb "gopkg.in/cheggaaa/pb.v1"
)

const (
	//south-east to north-west
	//lat goes south north; long east west
	latEnd, lngEnd      = 43.452100, -80.497000 //43.45245, -80.49600
	latStart, lngStart  = 43.451000, -80.494500 //43.45135, -80.49400
	sampleResolutionLat = 0.00001               //degrees
	sampleResolutionLng = 0.00001               //degrees

)

func GetModel() (maxVert float64) {

	// /*
	compositeVector := []*MapVector{}
	//read 3D model into struct
	clientsFile, err := os.Open("resultVectorModel.csv")
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()
	if err := gocsv.UnmarshalFile(clientsFile, &compositeVector); err != nil {
		panic(err)
	}

	//finds the min/max of ground height, latitude and latitude
	compositeVector[0].VertY = compositeVector[0].Elevation / 0.95 * 0.00001
	minVertY := compositeVector[0].VertY
	maxVertY := compositeVector[0].VertY
	for i := 0; i <= int(len(compositeVector)-1); i++ {
		compositeVector[i].VertY = compositeVector[i].Elevation / 0.95 * 0.00001
		if compositeVector[i].VertY < minVertY {
			minVertY = compositeVector[i].VertY
		}
		if compositeVector[i].VertY > maxVertY {
			maxVertY = compositeVector[i].VertY
		}
	}
	minVertX := math.Abs(compositeVector[0].Latitude)
	minVertZ := math.Abs(compositeVector[0].Longtitude)
	//localize the maximum elevation to the ground reference
	maxVertY = maxVertY - minVertY

	//localize the area using the minimum component of each vector as reference
	for i := 0; i <= int(len(compositeVector)-1); i++ {
		compositeVector[i].VertX = (math.Abs(compositeVector[i].Latitude) - minVertX)
		compositeVector[i].VertY = (compositeVector[i].VertY - minVertY)
		compositeVector[i].VertZ = (math.Abs(compositeVector[i].Longtitude) - minVertZ)
	}
	maxVertX := math.Abs(compositeVector[len(compositeVector)-1].VertX)
	maxVertZ := math.Abs(compositeVector[len(compositeVector)-1].VertZ)
	maxVert = math.Max(math.Max(maxVertX, maxVertZ), maxVertY)

	clientsFile2, err := os.OpenFile("resultNormModel.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile2.Close()
	err = gocsv.MarshalFile(&compositeVector, clientsFile2)
	if err != nil {
		panic(err)
	}

	data := [][]string{
		{strconv.FormatFloat(maxVertX, 'E', -1, 64),
			strconv.FormatFloat(maxVertY, 'E', -1, 64),
			strconv.FormatFloat(maxVertZ, 'E', -1, 64),
			strconv.FormatFloat(maxVert, 'E', -1, 64),
			strconv.FormatFloat(minVertX, 'E', -1, 64),
			strconv.FormatFloat(minVertY, 'E', -1, 64),
			strconv.FormatFloat(minVertZ, 'E', -1, 64)},
	}

	file, err := os.Create("resultNormModelProperties.csv")
	util.CheckError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		util.CheckError("Cannot write to file", err)
	}

	return maxVert
	// */
	// return 0.00199
}

func GetMapVector(apiKey *string) ([]*MapVector, []*MapPrimitiveIndex) {

	var compositeVector []*MapVector
	var compositeVectorElem *MapVector
	var primitiveIndex []*MapPrimitiveIndex
	var primitiveIndexElem *MapPrimitiveIndex

	baseLat := 2
	baseLng := int(util.Round(math.Abs((lngEnd - lngStart) / sampleResolutionLng)))
	latHeight := int(util.Round(math.Abs((latEnd - latStart) / sampleResolutionLat)))
	latBaseIndex, lngBaseIndex := 1, 1
	latBaseHeight := latStart + sampleResolutionLat
	latBaseGround := latStart
	vectorIndex := 0.0
	//sometimes the count target is over by baseLng elements, because sampling goes over the boundary
	downloadProgress := pb.StartNew(int((baseLng) * (latHeight - 1 + baseLat)))
	// 	key := strings.TrimSuffix(*apiKey, "\n") on windows
	key := strings.TrimSuffix(*apiKey, "\n")
	pretty.Println("here is the apiKey:", apiKey)
	pretty.Println("here is the key:", key)

	clientAccount, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	//TODO: hangle case for less than 2 baseLng
	for lngBaseIndex+latBaseIndex <= baseLat+baseLng {
		for lngBaseIndex > 0 && latBaseIndex > 0 && lngBaseIndex <= baseLng && latBaseIndex <= baseLat {

			lngLocation := lngStart + (float64(lngBaseIndex-1)*(lngEnd-lngStart))/(float64(baseLng)-1.0)
			latLocation := latBaseGround + (float64(latBaseIndex-1)*(latBaseHeight-latBaseGround))/(float64(baseLat)-1.0)

			r := &maps.ElevationRequest{
				Locations: []maps.LatLng{
					{Lat: latLocation, Lng: lngLocation},
				},
			}
			baseVector, err := clientAccount.Elevation(context.Background(), r)
			if err != nil {
				log.Fatalf("fatal error: %s", err)
			}

			compositeVectorElem = &MapVector{
				VertX: 0,
				//90deg on X is flip Y and Z,then -ve nowY; -90deg is flip then -ve nowZ
				VertZ:      0,
				VertY:      0,
				Latitude:   (*baseVector[0].Location).Lat,
				Longtitude: (*baseVector[0].Location).Lng,
				Elevation:  baseVector[0].Elevation,
			}

			compositeVector = append(compositeVector, compositeVectorElem)

			primitiveCounter := math.Mod(vectorIndex, 2)

			if primitiveCounter == 1 && len(compositeVector) > 2 {
				primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
				primitiveIndexElem.PrimitiveBottom = int(vectorIndex - 3)
				primitiveIndexElem.PrimitiveTop = int(vectorIndex - 2)
				primitiveIndexElem.PrimitiveLeft = int(vectorIndex - 1)
				primitiveIndex = append(primitiveIndex, primitiveIndexElem)
				primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
				primitiveIndexElem.PrimitiveBottom = int(vectorIndex - 1)
				primitiveIndexElem.PrimitiveTop = int(vectorIndex - 2)
				primitiveIndexElem.PrimitiveLeft = int(vectorIndex - 0)
				primitiveIndex = append(primitiveIndex, primitiveIndexElem)
			}
			downloadProgress.Increment()
			vectorIndex++
			latBaseIndex--
			lngBaseIndex++
		}
		latBaseIndex += lngBaseIndex
		lngBaseIndex = 1
		if latBaseIndex >= baseLat {
			lngBaseIndex += latBaseIndex - baseLat
			latBaseIndex = baseLat
		}
	}
	vectorIndex--

	if 1 <= latHeight-1 {
		for latTier := 0; latTier <= latHeight-baseLat; latTier += 2 {
			for assignedIndex, assignedVector := range compositeVector[(latTier)*baseLng : (latTier+2)*baseLng] {
				if util.Odd(assignedIndex) {

					r := &maps.ElevationRequest{
						Locations: []maps.LatLng{
							{Lat: assignedVector.Latitude + sampleResolutionLat, Lng: assignedVector.Longtitude},
						},
					}
					baseVector, err := clientAccount.Elevation(context.Background(), r)
					if err != nil {
						log.Fatalf("fatal error: %s", err)
					}
					compositeVectorElem = &MapVector{
						VertX:      0,
						VertZ:      0,
						VertY:      0,
						Latitude:   (*baseVector[0].Location).Lat,
						Longtitude: (*baseVector[0].Location).Lng,
						Elevation:  baseVector[0].Elevation,
					}
					compositeVector = append(compositeVector, compositeVectorElem)

					r = &maps.ElevationRequest{
						Locations: []maps.LatLng{
							{Lat: assignedVector.Latitude + sampleResolutionLat*2, Lng: assignedVector.Longtitude},
						},
					}
					baseVector, err = clientAccount.Elevation(context.Background(), r)
					if err != nil {
						log.Fatalf("fatal error: %s", err)
					}
					compositeVectorElem = &MapVector{
						VertX:      0,
						VertZ:      0,
						VertY:      0,
						Latitude:   (*baseVector[0].Location).Lat,
						Longtitude: (*baseVector[0].Location).Lng,
						Elevation:  baseVector[0].Elevation,
					}
					compositeVector = append(compositeVector, compositeVectorElem)

					primitiveCounterTiers := math.Mod(float64(latTier+1), 2)
					indexBoundaryLng := ((latTier-latTier/2-1)+2)*baseLng*2 + 2
					loopTierCounter := ((latTier - latTier/2 - 1) + 1) * baseLng * 2

					if primitiveCounterTiers == 1 && len(compositeVector) > indexBoundaryLng {
						primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
						primitiveIndexElem.PrimitiveBottom = int(assignedIndex + loopTierCounter - 2)
						primitiveIndexElem.PrimitiveTop = int(vectorIndex - 1)
						primitiveIndexElem.PrimitiveLeft = int(assignedIndex + loopTierCounter)
						primitiveIndex = append(primitiveIndex, primitiveIndexElem)
						primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
						primitiveIndexElem.PrimitiveBottom = int(assignedIndex + loopTierCounter)
						primitiveIndexElem.PrimitiveTop = int(vectorIndex - 1)
						primitiveIndexElem.PrimitiveLeft = int(vectorIndex + 1)
						primitiveIndex = append(primitiveIndex, primitiveIndexElem)
					}
					vectorIndex += 2

					if assignedIndex == baseLng*2-1 {
						for i := 0; baseLng-1 > i; i++ {
							primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
							primitiveIndexElem.PrimitiveBottom = int(vectorIndex) - (baseLng*2 - 1) + i*2
							primitiveIndexElem.PrimitiveTop = int(vectorIndex) - (baseLng*2 - 2) + i*2
							primitiveIndexElem.PrimitiveLeft = int(vectorIndex) - (baseLng*2 - 3) + i*2
							primitiveIndex = append(primitiveIndex, primitiveIndexElem)
							primitiveIndexElem = &MapPrimitiveIndex{0, 0, 0}
							primitiveIndexElem.PrimitiveBottom = int(vectorIndex) - (baseLng*2 - 3) + i*2
							primitiveIndexElem.PrimitiveTop = int(vectorIndex) - (baseLng*2 - 2) + i*2
							primitiveIndexElem.PrimitiveLeft = int(vectorIndex) - (baseLng*2 - 4) + i*2
							primitiveIndex = append(primitiveIndex, primitiveIndexElem)
						}
					}
					downloadProgress.Increment()
					downloadProgress.Increment()
				}

			}

		}
	}

	clientsFile, err := os.OpenFile("resultVectorModel.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	err = gocsv.MarshalFile(&compositeVector, clientsFile)
	if err != nil {
		panic(err)
	}

	clientsFile2, err := os.OpenFile("resultPrimativeModel.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile2.Close()

	err = gocsv.MarshalFile(&primitiveIndex, clientsFile2)
	if err != nil {
		panic(err)
	}

	downloadProgress.FinishPrint("Vectors downloaded.")
	return compositeVector, primitiveIndex
}

func MapBoundary() (float64, float64, float64) {
	ellipsoidConfig := ellipsoid.Init(
		"WGS84",
		ellipsoid.Degrees,
		ellipsoid.Meter,
		ellipsoid.LongitudeIsSymmetric,
		ellipsoid.BearingIsSymmetric)

	xDistance, _ := ellipsoidConfig.To(
		latStart,
		lngStart,
		latStart,
		lngEnd)

	yDistance, _ := ellipsoidConfig.To(
		latStart,
		lngStart,
		latEnd,
		lngStart)

	return yDistance, xDistance, math.Max(xDistance, yDistance)
}
