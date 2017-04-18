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
	IrcConn        *ircevt.Connection // active connection to irc server
	Config         IRCconfig
	Identity       Identity // user's identity currently impersonated, or im-concierge identity in concierge mode
	Is_Concierge   bool     // if true user is "im-concierge", not into impersonate mode
	Backend        backend.ConciergeBackend
	ConciergeHatch chan Notification
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

func NewIRCconnector(b backend.ConciergeBackend, hatch chan Notification, conf IRCconfig) (conn *IRCconnector, err error) {
	conn = new(IRCconnector)
	conn.Config = conf
	conn.ConciergeHatch = hatch
	conn.Backend = b
	return
}

// Add a connection to an IRC room as "IM-concierge" to monitor it
func (conn *IRCconnector) AddConcierge(user Identity) (mapkey string, err error) {
	conn.Identity = user
	conn.Is_Concierge = true

	irccon := ircevt.IRC(conn.Identity.DisplayName, conn.Identity.Identifier)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *ircevt.Event) {
		log.Infof("joinning room %s", conn.Config.IRCRoom)
		irccon.Join(conn.Config.IRCRoom)
	})
	irccon.AddCallback("353", func(e *ircevt.Event) {
		//HandleUsersList(irccon, e)
	})
	irccon.AddCallback("JOIN", func(e *ircevt.Event) {
		//HandleUsersList(irccon, e)
	})
	//irccon.AddCallback("352", HandleWhoReply)
	irccon.AddCallback("PRIVMSG", conn.HandleMessage)

	err = irccon.Connect(conn.Config.IRCserver)
	if err != nil {
		return
	}

	conn.IrcConn = irccon
	go conn.IrcConn.Loop()

	mapkey = conn.Config.IRCRoom + ":" + conn.Identity.DisplayName
	return
}

// Add a new connection to the server/room with an user's identity
// Return a new connector ready to send message on behalf of user
func (conn *IRCconnector) Impersonate(user Identity) (mapkey string, err error) {
	conn.Identity = user
	conn.Is_Concierge = false

	irccon := ircevt.IRC(conn.Identity.DisplayName, conn.Identity.Identifier)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *ircevt.Event) {
		log.Infoln("joinning room %s as %s", conn.Config.IRCRoom, conn.Identity.DisplayName)
		irccon.Join(conn.Config.IRCRoom)
	})

	err = irccon.Connect(conn.Config.IRCserver)
	if err != nil {
		return
	}

	conn.IrcConn = irccon
	go conn.IrcConn.Loop()

	mapkey = conn.Config.IRCRoom + ":" + conn.Identity.DisplayName
	return
}

func (conn *IRCconnector) NotifyConcierge(notif Notification) error {
	timer := time.NewTimer(2 * time.Second)
	select {
	case conn.ConciergeHatch <- notif:
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

func (conn *IRCconnector) Remove() error {
	conn.IrcConn.Quit()
	conn.IrcConn.Disconnect()
	return nil
}
