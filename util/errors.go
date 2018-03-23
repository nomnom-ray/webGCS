package util

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("username exists")
	ErrBadPrimitivePick   = errors.New("Primitive not selected")
	ErrAnnotationNotFound = errors.New("annotation not in redis")
	ErrDeleting           = errors.New("annotation not removed or not exist")
	ErrDeletePoint        = errors.New("deleting a point")
)

func InternalServerError(err error, wsConn *websocket.Conn) {
	if ce, ok := err.(*websocket.CloseError); ok {
		switch ce.Code {
		case websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
			websocket.CloseNoStatusReceived:
			log.Printf("Web socket closed by client: %s", err)
			wsConn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

func InternalServerErrorHTTP(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("internal server error"))
}

func CheckError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
