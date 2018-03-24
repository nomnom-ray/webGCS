package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nomnom-ray/webGCS/models"
	"github.com/nomnom-ray/webGCS/server"
	"github.com/nomnom-ray/webGCS/util"
)

//LoadRoutes has r object repalcing http with router to serve complex multi user environments
func LoadRoutes(h *server.Hub) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", server.AuthRequired(indexGetHandler))
	r.HandleFunc("/ws", server.AuthRequired(h.ServeHTTP))

	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")

	fs := http.FileServer(http.Dir("./templates"))                           //inst. a file server object; and where files are served from
	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", fs)) //tell routor to use path with static prefix

	return r
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	util.ExecuteTemplates(w, "index.html", nil)
}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	util.ExecuteTemplates(w, "register.html", nil)
}

func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := server.StoreSessions.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/register", 302)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")

	exists, err := models.CheckUserExist(username)
	if err != nil {
		util.InternalServerErrorHTTP(w)
		return
	}
	if !exists {
		err = models.RegisterUser(username)
		if err != nil {
			util.InternalServerErrorHTTP(w)
			return
		}
	}

	user, err := models.GetUserbyName(username)
	if err != nil {
		switch err {
		case util.ErrUserNotFound:
			util.ExecuteTemplates(w, "login.html", "unknown user")
		default: //server error
			util.InternalServerErrorHTTP(w)
		}
		return //when all error is already handled
	}
	userID, err := user.GetID()
	if err != nil {
		util.InternalServerErrorHTTP(w)

		return
	}
	session, _ := server.StoreSessions.Get(r, "session") //assigns cookied session to client request; create new if none
	session.Values["user_id"] = userID                   //store and save session data
	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
}
