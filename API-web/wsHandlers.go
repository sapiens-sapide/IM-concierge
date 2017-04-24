// Websocket IO with clients
package web_api

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	. "github.com/sapiens-sapide/IM-concierge/entities"
	"github.com/satori/go.uuid"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//TODO : handle authentication with websocket proto

// handler for route /ws, in charge of upgrading client to websocket protocol
// if ws upgrade is ok, client is added to frontserver's clients map and concierge is warned
func (fs *FrontServer) RegisterClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err)
		return
	}
	fs.ClientsMux.Lock()
	newClient := FrontClient{
		Websocket: conn,
		Identity: Identity{
			UserId:      uuid.NewV4(),
			DisplayName: fs.Config.User.IRCNickname,
			Identifier:  fs.Config.User.IRCUser,
		}, //TODO
		FromClient:  make(chan wsEvent),
		ToClient:    make(chan []byte),
		LeaveClient: make(chan bool),
	}
	fs.Clients[newClient.Identity.UserId] = &newClient

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
		clientQuitEvent := clientEvent{ClientLeave, client.Identity}
		fs.NotifyConcierge(clientQuitEvent)
		fs.removeClient(client.Identity.UserId)
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
					client.LeaveClient <- true
				} else {
					var evt wsEvent
					err := json.Unmarshal(buf.Bytes(), &evt)
					if err != nil {
						log.Warn("unable to unmarshall ws event")
						continue
					} else {
						client.FromClient <- evt
					}
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
		case message := <-client.FromClient:
			switch message.Event {
			case "connect":
				newClientEvt := clientEvent{ImpersonnateUser, client.Identity}
				fs.NotifyConcierge(newClientEvt)
			case "disconnect":
				// TODO: for now, only remove user from the config file
				userQuitEvent := clientEvent{StopImpersonnateUser, client.Identity}
				fs.NotifyConcierge(userQuitEvent)
			case "message":
				newMessageClient := newMessageClientEvent{
					Type:    ClientPostMessage,
					Message: message.Payload,
				}
				fs.NotifyConcierge(newMessageClient)
			default:
				log.Warnf("unknown event <%s> from client %s", message.Event, client.Identity.UserId.String())
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
