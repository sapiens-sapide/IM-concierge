package imc

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	m "github.com/sapiens-sapide/IM-concierge/models"
	"net/http"
	"time"
)

// handler for /messages route
func (c *Concierge) AllMessagesHandler(ctx *gin.Context) {
	messages, err := c.Backend.ListMessagesByDate("#im-concierge-playground", time.Now().AddDate(0, 0, -defaultMessagesDaysAge))
	if err != nil {
		log.WithError(err).Warnln("[FrontServer] : error when calling Backend.ListMessagesByDate")
	}

	ctx.HTML(http.StatusOK, "messages.tmpl", messages)
}

// handler for route /ws, in charge of upgrading client to websocket protocol
func (c *Concierge) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err)
		return
	}
	c.ClientsMux.Lock()
	newClient := FrontClient{
		Websocket:   conn,
		Identity:    m.Identity{},
		FromClient:  make(chan []byte),
		ToClient:    make(chan []byte),
		LeaveClient: make(chan bool),
		ClientPos:   len(c.Clients),
	}
	c.Clients = append(c.Clients, &newClient)
	go newClient.WsClientHandler(c.PopClient)
	c.ClientsMux.Unlock()
}

// handles communications with connected clients upgraded to websocket protocol
func (client *FrontClient) WsClientHandler(popChan chan int) {
	client.Websocket.SetReadLimit(maxMessageSize)
	client.Websocket.SetReadDeadline(time.Now().Add(pongWait))
	client.Websocket.SetPongHandler(func(string) error { client.Websocket.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	defer func() {
		close(client.FromClient)
		close(client.ToClient)
		client.Websocket.Close()
		popChan <- client.ClientPos // remove this client from concierge's clients array
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
