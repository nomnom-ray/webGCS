package models

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/nfnt/resize"
	"github.com/nomnom-ray/fauxgl"
	"github.com/nomnom-ray/webGCS/util"
)

//MapVector are vector coordinates in global coordinates and raster coordinates
type MapVector struct {
	VertX, VertY, VertZ             float64
	Latitude, Longtitude, Elevation float64
}

//MapPrimitiveIndex are object buffer index for vertices
type MapPrimitiveIndex struct {
	PrimitiveBottom, PrimitiveTop, PrimitiveLeft int
}

var (
	windWidth        = 600 //1280.0
	windHeight       = 600 //720.0
	imageAspectRatio = float64(windWidth / windHeight)
	scale            = 4                                  // optional supersampling
	fovy             = 80.0                               // vertical field of view in degrees
	near             = 0.001                              // near clipping plane
	far              = 10.0                               // far clipping plane
	eye              = fauxgl.V(0, 0, 0)                  // camera position
	center           = fauxgl.V(0, 0, 1)                  // view center position
	up               = fauxgl.V(0, 1, 0)                  // up vector
	light            = fauxgl.V(0.75, 0.5, 1).Normalize() // light direction
	color            = fauxgl.HexColor("#ffb5b5")         // object color
	background       = fauxgl.HexColor("#FFF8E3")         // background color
)

// CameraModel combines camera and projection parameters to create a single matrix
func CameraModel(maxVert float64, cameraLocation *MapVector) fauxgl.Matrix {

	cameraRotationLR := float64(-205) - 90 + 0.2          //-ve rotates camera clockwise in degrees
	cameraRotationUD := float64(-10.0)                    //-ve rotates camera downwards in degrees
	cameraX := float64(cameraLocation.VertX)              //-ve pans camera to the right
	cameraZ := float64(cameraLocation.VertZ)              //-ve pans camera to the back
	cameraHeight := float64(-0.00002252)                  //height of the camera from ground
	groundRef := float64(-cameraLocation.VertY) + 0.00004 //ground reference to the lowest ground point in the tile

	cameraPosition := fauxgl.Vector{
		X: cameraX / maxVert,
		Y: (cameraHeight + groundRef) / maxVert,
		Z: cameraZ / maxVert,
	}
	cameraViewDirection := fauxgl.Vector{
		X: 0,
		Y: 0,
		Z: 1,
	}
	cameraUp := fauxgl.Vector{
		X: 0,
		Y: -1,
		Z: 0,
	}
	cameraViewDirection = fauxgl.QuatRotate(
		util.DegToRad(cameraRotationLR), cameraUp).Rotate(cameraViewDirection)
	cameraViewDirection = fauxgl.QuatRotate(
		util.DegToRad(cameraRotationUD), cameraViewDirection.Cross(cameraUp)).Rotate(cameraViewDirection)
	cameraPerspective := fauxgl.LookAt(
		cameraPosition, (cameraPosition).Add(cameraViewDirection), cameraUp).Perspective(
		fovy, imageAspectRatio, near, far)

	return cameraPerspective
}

//NormCameraLocation normalized camera location from GCS to raster coordinates
func NormCameraLocation(cameraLocation *MapVector) *MapVector {

	var compositeVector []*MapVector
	compositeVector = append(compositeVector, cameraLocation)

	clientsFile, err := os.Open("resultNormModelProperties.csv")
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	propertiesReader := csv.NewReader(bufio.NewReader(clientsFile))

	var _, minVertX, minVertZ float64

	for i := 0; i < 7; i++ {
		property, error := propertiesReader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		_, err = strconv.ParseFloat(property[3], 64)
		if err != nil {
			panic(err)
		}
		minVertX, _ = strconv.ParseFloat(property[4], 64)
		minVertZ, _ = strconv.ParseFloat(property[6], 64)
	}

	//localize the area using the minimum component of each vector as reference
	for i := 0; i <= int(len(compositeVector)-1); i++ {
		compositeVector[i].VertX = (math.Abs(compositeVector[i].Latitude) - minVertX)
		compositeVector[i].VertY = compositeVector[i].Elevation
		compositeVector[i].VertZ = (math.Abs(compositeVector[i].Longtitude) - minVertZ)
	}

	return cameraLocation
}

//Projection transforms 3D spaces to raster space using CameraModel and NormCameraLocation
func Projection(maxVert float64, cameraPerspective fauxgl.Matrix) ([]*fauxgl.Triangle, []int) {

	compositeVector := []*MapVector{}
	primitiveIndex := []*MapPrimitiveIndex{}

	//read 3D vector model into struct
	clientsFile, err := os.Open("resultNormModel.csv")
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()
	if err := gocsv.UnmarshalFile(clientsFile, &compositeVector); err != nil {
		panic(err)
	}
	//read 3D vector index into struct
	clientsFile2, err := os.Open("resultPrimativeModel.csv")
	if err != nil {
		panic(err)
	}
	defer clientsFile2.Close()
	if err := gocsv.UnmarshalFile(clientsFile2, &primitiveIndex); err != nil {
		panic(err)
	}

	//normalize 3D model to 1:1:1 camera space
	for i := 0; i <= int(len(compositeVector)-1); i++ {
		compositeVector[i].VertX = compositeVector[i].VertX / maxVert
		compositeVector[i].VertY = compositeVector[i].VertY / maxVert
		compositeVector[i].VertZ = compositeVector[i].VertZ / maxVert
	}

	// pretty.Println("map location:", compositeVector[6464])

	//constructing a mesh of triangles from index to normalized vertices
	var triangles []*fauxgl.Triangle
	counter := 0.0
	primitiveIDCounter := 0

	primitiveCounter := 0.0
	for _, index := range primitiveIndex[:] {
		var triangle fauxgl.Triangle
		for inner := 0; inner < 3; inner++ {
			primitiveCounter = math.Mod(counter, 3)
			if primitiveCounter == 0 {
				triangle.V1.Position = fauxgl.Vector{
					X: compositeVector[index.PrimitiveBottom].VertX,
					Y: compositeVector[index.PrimitiveBottom].VertY,
					Z: compositeVector[index.PrimitiveBottom].VertZ,
				}
				triangle.V1.Texture = fauxgl.Vector{
					X: compositeVector[index.PrimitiveBottom].Latitude,
					Y: compositeVector[index.PrimitiveBottom].Elevation,
					Z: compositeVector[index.PrimitiveBottom].Longtitude,
				}
			} else if primitiveCounter == 1 {
				triangle.V2.Position = fauxgl.Vector{
					X: compositeVector[index.PrimitiveTop].VertX,
					Y: compositeVector[index.PrimitiveTop].VertY,
					Z: compositeVector[index.PrimitiveTop].VertZ,
				}
				triangle.V2.Texture = fauxgl.Vector{
					X: compositeVector[index.PrimitiveTop].Latitude,
					Y: compositeVector[index.PrimitiveTop].Elevation,
					Z: compositeVector[index.PrimitiveTop].Longtitude,
				}
			} else if primitiveCounter == 2 {
				triangle.V3.Position = fauxgl.Vector{
					X: compositeVector[index.PrimitiveLeft].VertX,
					Y: compositeVector[index.PrimitiveLeft].VertY,
					Z: compositeVector[index.PrimitiveLeft].VertZ,
				}
				triangle.V3.Texture = fauxgl.Vector{
					X: compositeVector[index.PrimitiveLeft].Latitude,
					Y: compositeVector[index.PrimitiveLeft].Elevation,
					Z: compositeVector[index.PrimitiveLeft].Longtitude,
				}
			}
			counter++
		}
		triangle.PrimitiveID = int(primitiveIDCounter)
		triangle.FixNormals()
		triangles = append(triangles, &triangle)
		primitiveIDCounter++
	}
	mesh := fauxgl.NewEmptyMesh()
	triangleMesh := fauxgl.NewTriangleMesh(triangles)
	mesh.Add(triangleMesh)

	//creating the window for CPU render
	contextRender := fauxgl.NewContext(windWidth*scale, windHeight*scale)
	contextRender.Wireframe = true
	contextRender.SetPickingFlag(false)
	contextRender.ClearColorBufferWith(fauxgl.Transparent)
	// contextRender.ClearDepthBuffer()

	//shading
	shader := fauxgl.NewSolidColorShader(cameraPerspective, color)
	contextRender.Shader = shader
	start := time.Now()
	contextRender.DrawMesh(mesh)
	fmt.Println("**********RENDERING**********", time.Since(start), "**********RENDERING**********")

	image := contextRender.Image()
	image = resize.Resize(uint(windWidth), uint(windHeight), image, resize.Bilinear)

	fauxgl.SavePNG("out.png", image)

	return triangles, contextRender.PrimitiveSelectable()
}

//RasterPicking reiterate projection with a single set of raster coordinates to find 1 vertex
func RasterPicking(pickedX, pickedY int,
	triangles []*fauxgl.Triangle, primitiveOnScreen []int, cameraPerspective fauxgl.Matrix) (*fauxgl.Triangle, *fauxgl.Vertex, bool) {

	var trianglesOnScreen []*fauxgl.Triangle

	primitiveOnScreen = sliceUniqMap(primitiveOnScreen)

	if len(primitiveOnScreen) == 1 {
		trianglesOnScreen = append(trianglesOnScreen, triangles[primitiveOnScreen[0]])
	} else if len(primitiveOnScreen) > 1 {
		for _, i := range primitiveOnScreen {
			trianglesOnScreen = append(trianglesOnScreen, triangles[i])
		}
	}

	meshOnScreen := fauxgl.NewEmptyMesh()
	triangleMesh := fauxgl.NewTriangleMesh(trianglesOnScreen)
	meshOnScreen.Add(triangleMesh)

	//creating the window for CPU render
	contextPicking := fauxgl.NewContext(windWidth*scale, windHeight*scale)
	contextPicking.SetPickedXY(pickedX*scale, pickedY*scale)
	contextPicking.SetPickingFlag(true)
	contextPicking.SetPrimitiveOnScreen(nil)
	// contextPicking.ClearDepthBuffer()

	//shading
	shader := fauxgl.NewSolidColorShader(cameraPerspective, color)
	contextPicking.Shader = shader
	start := time.Now()
	contextPicking.DrawMesh(meshOnScreen)
	fmt.Println("***********PICKING***********", time.Since(start), "***********PICKING***********")

	if ok, _ := contextPicking.ReturnedPick(); ok == nil {
		return nil, nil, false
	}

	triangle, vertex := contextPicking.ReturnedPick()

	return triangle, vertex, true
}

func sliceUniqMap(s []int) []int {
	seen := make(map[int]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
