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

	//COMMENT: change between localhost and server IP for dev vs. production
	var deployFlag = true

	models.InitRedis()
	tile := models.InitTiles()

	//for transmitting tile sections to client; this should be done by POSTGIS, not redis.
	// if ok := models.TileCheck(); !ok {
	// 	tile.NewRedisTile()
	// }

	projectedTile := tile.InitProjectedTiles()

	util.LoadTemplates("templates/*.html")

	h := server.NewHub(projectedTile)
	r := router.LoadRoutes(h)

	http.Handle("/", r) //use the mux router as the default handler

	if !deployFlag {
		log.Printf("serving HTTP on port :8080")
		log.Printf("serving HTTPS on port :8081")
		go http.ListenAndServe(":8080", http.HandlerFunc(router.RedirectToHTTPSDev))
		go log.Fatal(http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", r))
	} else {
		log.Printf("serving HTTP on port :80")
		log.Printf("serving HTTPS on port :81")
		go http.ListenAndServe(":80", http.HandlerFunc(router.RedirectToHTTPS))
		go log.Fatal(http.ListenAndServeTLS(":81", "cert.pem", "key.pem", r))
	}

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
