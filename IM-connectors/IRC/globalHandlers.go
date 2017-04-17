package irc

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/oklog/ulid"
	"github.com/sapiens-sapide/IM-concierge/API-web"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
	ircevt "github.com/thoj/go-ircevent"
	"time"
)

type IRCconnector struct {
	IrcConn *ircevt.Connection       // active connection to irc server
	Rooms   map[string]Room          // IRC rooms currently monitored
	Backend backend.ConciergeBackend //backend to persist & retreive data
	Config  ConciergeConfig
	Hatch   chan Notification // to send events to concierge
}

type Room struct {
	Name           string
	connectedUsers []*web_api.FrontClient
	knownUsers     []Identity
	Id             ulid.ULID
}

type NewIRCMessageEvent struct {
	Type    EventType
	Message Message
}

func InitIRCconnector(conf ConciergeConfig, backend backend.ConciergeBackend, hatch chan Notification) (connector *IRCconnector, err error) {
	connector = new(IRCconnector)
	connector.Config = conf
	connector.Backend = backend
	connector.Hatch = hatch
	irccon := ircevt.IRC(conf.IRCNickname, conf.IRCUser)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *ircevt.Event) {
		irccon.Join(conf.IRCRoom)
	})
	irccon.AddCallback("353", func(e *ircevt.Event) {
		//HandleUsersList(irccon, e)
	})
	irccon.AddCallback("JOIN", func(e *ircevt.Event) {
		//HandleUsersList(irccon, e)
	})
	//irccon.AddCallback("352", HandleWhoReply)
	irccon.AddCallback("PRIVMSG", connector.HandleMessage)

	connector.IrcConn = irccon

	return
}

func (conn *IRCconnector) Start() error {
	log.Infoln("going to connect to IRC")
	err := conn.IrcConn.Connect(conn.Config.IRCserver)
	if err != nil {
		return err
	}
	nick := conn.IrcConn.GetNick()
	log.Infof("In room %s, as %s", conn.Config.IRCRoom, nick)
	go conn.IrcConn.Loop()
	return nil
}

/*
func (conn *IRCconnector) RegisterRooms(c *Concierge, rooms []string) error {
	for _, room := range rooms {
		if _, ok := c.Rooms[room]; !ok {
			roomId, err := c.Backend.RegisterRoom(room)
			if err != nil {
				return err
			}
			room_ulid, _ := ulid.Parse(roomId)
			c.Rooms[room] = ent.Room{
				Name:           room,
				Id:             room_ulid,
				ConnectedUsers: []*ent.FrontClient{},
				KnownUsers:     []ent.Identity{},
			}
		}
	}
	return nil
}
*/
func (conn *IRCconnector) NotifyConcierge(not Notification) error {
	timer := time.NewTimer(2 * time.Second)
	select {
	case conn.Hatch <- not:
		timer.Stop()
		return nil
	case <-timer.C:
		return errors.New("irc connector timeout when notifying concierge")
	}
}

func (nme NewIRCMessageEvent) EventType() EventType {
	return nme.Type
}

func (nme NewIRCMessageEvent) Payload() (interface{}, error) {
	return json.Marshal(nme.Message)
}

func (conn *IRCconnector) Shutdown() error {
	conn.IrcConn.Quit()
	conn.IrcConn.Disconnect()
	return nil
}
