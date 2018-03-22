package util

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	UserExists       = errors.New("username exists")
	BadPrimitivePick = errors.New("Primitive not selected")
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
