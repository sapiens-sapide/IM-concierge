package imc

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Concierge struct {
	clients  []FrontClient // users currently connected to the front server
	ircConn  []string      // active connection(s) to irc server(s)
	channels []string      // channels currently monitored
}

type FrontClient struct {
	Ws          *websocket.Conn
	Identity    Identity
	FromClient  chan []byte
	ToClient    chan []byte
	LeaveClient chan bool
}

type Channel struct {
	Name           string
	connectedUsers []*FrontClient
	knownUsers     []Identity
}

func (client *FrontClient) WsClientHandler() {
	client.Ws.SetReadLimit(maxMessageSize)
	client.Ws.SetReadDeadline(time.Now().Add(pongWait))
	client.Ws.SetPongHandler(func(string) error { client.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	defer func() {
		close(client.FromClient)
		close(client.ToClient)
		client.Ws.Close()
		//todo : send end message to concierge to pop the client from list
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
	}(client.Ws)

	//handle communication with client
	ticker := time.Tick(time.Second * 15)
	for {
		select {
		case <-client.LeaveClient:
			return
		case message, ok := <-client.ToClient:
			err := client.Ws.WriteMessage(websocket.TextMessage, message)
			if ok && err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		case message, ok := <-client.FromClient:
			log.Infof("message received : %s", message)
			err := client.Ws.WriteMessage(websocket.TextMessage, []byte("got the message"))
			if ok && err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		case <-ticker:
			err := client.Ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*2))
			if err != nil {
				log.WithError(err).Warnln(err)
				return
			}
		}
	}
}
