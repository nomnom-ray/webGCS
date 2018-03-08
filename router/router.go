package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nomnom-ray/webGCS/server"
	"github.com/nomnom-ray/webGCS/util"
)

//LoadRoutes has r object repalcing http with router to serve complex multi user environments
func LoadRoutes(h *server.Hub) *mux.Router {
	r := mux.NewRouter()

	//different handles for different tasks
	r.HandleFunc("/", indexGetHandler)
	r.HandleFunc("/ws", h.ServeHTTP)

	fs := http.FileServer(http.Dir("./templates"))                           //inst. a file server object; and where files are served from
	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", fs)) //tell routor to use path with static prefix

	return r
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	util.ExecuteTemplates(w, "index.html", nil)
}
