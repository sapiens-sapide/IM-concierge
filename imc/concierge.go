package imc

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid"
	"github.com/sapiens-sapide/IM-concierge/backend"
	m "github.com/sapiens-sapide/IM-concierge/models"
	"sync"
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
	ClientsMux sync.Mutex
	Clients    []*FrontClient  // users currently connected to the front server
	IrcConn    []string        // active connection(s) to irc server(s)
	Rooms      map[string]Room // IRC rooms currently monitored
	Backend    backend.ConciergeBackend
	PopClient  chan int //chan to receive disconnection info from clients
}

type FrontClient struct {
	Websocket   *websocket.Conn
	Identity    m.Identity
	FromClient  chan []byte
	ToClient    chan []byte
	LeaveClient chan bool
	ClientPos   int // client position within Concierge's Clients array
}

type Room struct {
	Name           string
	connectedUsers []*FrontClient
	knownUsers     []m.Identity
	Id             ulid.ULID
}

var (
	Config    m.ConciergeConfig
	Users     map[string]m.Recipient
	CareTaker *Concierge
)

func NewCareTaker() (*Concierge, error) {
	be, err := backend.InitEliasBackend()
	if err != nil {
		log.WithError(err).Fatalln("failed to init Elias backend")
	}

	CareTaker = &Concierge{
		Backend:   be,
		Clients:   []*FrontClient{},
		PopClient: make(chan int),
		Rooms:     map[string]Room{},
	}

	err = CareTaker.RegisterRoom(Config.IRCRoom)
	if err != nil {
		return nil, err
	}

	go func(c *Concierge) {
		for clientPos := range c.PopClient {
			c.ClientsMux.Lock()
			c.Clients = append(c.Clients[:clientPos], c.Clients[clientPos+1:]...)
			for _, client := range c.Clients[clientPos:] {
				client.ClientPos--
			}
			c.ClientsMux.Unlock()
		}
	}(CareTaker)

	return CareTaker, nil
}

func (c *Concierge) RegisterRoom(room string) error {
	if _, ok := c.Rooms[room]; !ok {
		roomId, err := c.Backend.RegisterRoom(room)
		if err != nil {
			return err
		}
		room_ulid, _ := ulid.Parse(roomId)
		c.Rooms[room] = Room{
			Name:           room,
			Id:             room_ulid,
			connectedUsers: []*FrontClient{},
			knownUsers:     []m.Identity{},
		}
	}
	return nil
}
