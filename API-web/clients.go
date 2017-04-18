package web_api

import (
	"github.com/gorilla/websocket"
	. "github.com/sapiens-sapide/IM-concierge/entities"
)

//TODO: add protocols & channels infos into frontClient struct
// for now IRC assumed
type FrontClient struct {
	Websocket   *websocket.Conn
	Identity    Identity
	FromClient  chan []byte
	ToClient    chan []byte
	LeaveClient chan bool
}

type newClientEvent struct {
	Type   EventType
	Client Identity
}

func (nce newClientEvent) EventType() EventType {
	return nce.Type
}

func (nce newClientEvent) Payload() (interface{}, error) {
	return nce.Client, nil
}

type newMessageClientEvent struct {
	Type    EventType
	Message string
}

func (nmce newMessageClientEvent) EventType() EventType {
	return nmce.Type
}

func (nmce newMessageClientEvent) Payload() (interface{}, error) {
	return nmce.Message, nil
}
