package web_api

import (
	"github.com/gorilla/websocket"
	"github.com/sapiens-sapide/IM-concierge/entities"
)

type FrontClient struct {
	Websocket   *websocket.Conn
	Identity    entities.Identity
	FromClient  chan []byte
	ToClient    chan []byte
	LeaveClient chan bool
	ClientPos   int // client position within Concierge's Clients array
}
