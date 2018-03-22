package server

import (
	"net/http"

	"github.com/gorilla/sessions"
)

//storeSessions use cookies (from other ways) to store session of a user
var StoreSessions = sessions.NewCookieStore([]byte("T0p-s3cr3t")) //make unreproducible cookies: s3cr3t as encrption key

//AuthRequired generates a session(connection to server) for the client; with optional login function
func AuthRequired(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		//middelware: the tasks of the middleware before reaching the handler "h" object to be executed
		session, _ := StoreSessions.Get(r, "session") //get session
		_, ok := session.Values["user_id"]
		if !ok {
			http.Redirect(w, r, "/register", 302) //go to login page if no session
			return                                //return stops the process; doesn't proceed in main
		}
		//execute the handler
		h.ServeHTTP(w, r)
	}
}
