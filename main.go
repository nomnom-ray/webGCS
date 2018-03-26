package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nomnom-ray/webGCS/models"
	"github.com/nomnom-ray/webGCS/router"
	"github.com/nomnom-ray/webGCS/server"
	"github.com/nomnom-ray/webGCS/util"
)

func main() {

	models.InitRedis()
	tile := models.InitTiles()
	projectedTile := tile.InitProjectedTiles()

	util.LoadTemplates("templates/*.html")

	h := server.NewHub(projectedTile)
	r := router.LoadRoutes(h)

	http.Handle("/", r) //use the mux router as the default handler

	log.Printf("serving on port :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	//web client to get vectors; costs money and slow;
	//client will not run as long as resultRawModel.csv in folder
	_, err := os.Stat("resultVectorModel.csv")
	if err != nil {
		if os.IsNotExist(err) {
			models.GetMapVector(util.Scanner())
			// create a cartesian model with GCS as units
			models.GetModel()
		} else {
			log.Fatalf("fatal error: %s", err)
		}
	}

}
