package server

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kr/pretty"
)

//Hub defines what the connections are doing
type Hub struct {
	// the mutex to protect connections
	connectionsMx sync.RWMutex
	// Registered connections.
	connections map[*Connection]struct{}
	// Inbound messages from the connections.
	broadcast chan MessageProcessed
	process   chan Message
}

//NewHub creates hub instances for connections
func NewHub() *Hub {
	h := &Hub{
		connectionsMx: sync.RWMutex{},
		connections:   make(map[*Connection]struct{}),
		broadcast:     make(chan MessageProcessed),
		process:       make(chan Message),
	}

	go func() {
		for {
			msg := <-h.broadcast
			h.connectionsMx.RLock()
			for connections := range h.connections {
				select {
				case connections.send <- msg: //send msg to connection type on connection channel
				// stop trying to send to this connection after trying for 1 second.
				// if we have to stop, it means that a reader died so remove the connection also.
				case <-time.After(1 * time.Second):
					pretty.Println("shutting down connection: ", connections)
					h.removeConnection(connections)
				}
			}
			h.connectionsMx.RUnlock()
		}
	}()
	return h
}

func (h *Hub) addConnection(conn *Connection) {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()
	h.connections[conn] = struct{}{}
}

func (h *Hub) removeConnection(conn *Connection) {
	h.connectionsMx.Lock()
	defer h.connectionsMx.Unlock()
	if _, ok := h.connections[conn]; ok {
		delete(h.connections, conn)
		close(conn.send)
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading %s", err)
		return
	}

	c := &Connection{send: make(chan MessageProcessed), h: h}
	c.h.addConnection(c)
	defer c.h.removeConnection(c)

	c.syncToDatabase(wsConn)

	var wg sync.WaitGroup
	wg.Add(2)
	go c.writer(&wg, wsConn)
	go c.reader(&wg, wsConn)
	wg.Wait()
	wsConn.Close()
}
