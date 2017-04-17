// Websocket IO with clients
package web_api

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	ent "github.com/sapiens-sapide/IM-concierge/entities"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the ws peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the ws peer.
	pongWait = 60 * time.Second

	// Send pings to ws peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from ws peer.
	maxMessageSize = 512
)

// handler for route /ws, in charge of upgrading client to websocket protocol
func (fs *FrontServer) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err)
		return
	}
	fs.ClientsMux.Lock()
	newClient := FrontClient{
		Websocket:   conn,
		Identity:    ent.Identity{},
		FromClient:  make(chan []byte),
		ToClient:    make(chan []byte),
		LeaveClient: make(chan bool),
		ClientPos:   len(fs.Clients),
	}
	fs.Clients = append(fs.Clients, &newClient)
	go fs.WsClientHandler(&newClient)
	fs.ClientsMux.Unlock()
}

// handles ws communications with connected clients already upgraded to websocket protocol
func (fs *FrontServer) WsClientHandler(client *FrontClient) {
	client.Websocket.SetReadLimit(maxMessageSize)
	client.Websocket.SetReadDeadline(time.Now().Add(pongWait))
	client.Websocket.SetPongHandler(func(string) error { client.Websocket.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	defer func() {
		close(client.FromClient)
		close(client.ToClient)
		client.Websocket.Close()
		fs.PopClient <- client.ClientPos // remove this client from frontServer's clients array
	}()

	//listen and handle payload coming from client
	go func(c *websocket.Conn) {
		for {
			messageType, message, err := c.NextReader()
			if err != nil {
				log.WithError(err).Warnln(err)
				client.LeaveClient <- true
				break
			}
			if messageType == 1 || messageType == 2 {
				var buf bytes.Buffer
				_, err := buf.ReadFrom(message)
				if err != nil {
					log.WithError(err).Warnln(err)
					//todo
				} else {
					client.FromClient <- buf.Bytes()
				}
			}
		}
	}(client.Websocket)

	//handle communication with client
	ticker := time.Tick(time.Second * 15)
	for {
		select {
		case <-client.LeaveClient:
			return
		case message, ok := <-client.ToClient:
			err := client.Websocket.WriteMessage(websocket.TextMessage, message)
			if ok && err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		case message, ok := <-client.FromClient:
			//for now, send back payload to client… completely useless…
			err := client.Websocket.WriteJSON(message)
			if ok && err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		case <-ticker:
			err := client.Websocket.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
			if err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		}
	}
}
