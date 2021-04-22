package pkg

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request, hub *ClientHub) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// Allow connections on local network
		if strings.HasPrefix(r.RemoteAddr, "192.168.1.") {
			return true
		}
		// Allow connections on localhost
		if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || strings.HasPrefix(r.RemoteAddr, "localhost") {
			return true
		}
		return false
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := Client{
		hub:  hub,
		conn: conn,
	}

	client.hub.Register <- &client

	defer func() {
		client.hub.Deregister <- &client
		client.conn.Close()
	}()

	for {
		messageType, bytes, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		msg := Message{
			bytes:       bytes,
			messageType: messageType,
		}

		hub.Send <- msg
	}
}
