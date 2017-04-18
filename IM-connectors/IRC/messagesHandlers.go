package irc

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/sapiens-sapide/IM-concierge/entities"
	"github.com/thoj/go-ircevent"
	"time"
)

func (conn *IRCconnector) HandleMessage(e *irc.Event) {
	msg := Message{
		Body:     e.Message(),
		Date:     time.Now(),
		From:     e.Nick,
		Id:       NewULID(),
		Room:     e.Arguments[0],
		Protocol: IRC,
	}
	err := conn.Backend.StoreMessage(msg)
	if err != nil {
		log.Warnf("Error [handleMessage.StoreMessage] : %v", err)
	}
	//create an event and send it to concierge
	newMsgEvt := NewIRCMessageEvent{NewMessage, msg}
	conn.NotifyConcierge(newMsgEvt)
}

func (conn *IRCconnector) PostMessage(msg string) error {
	conn.IrcConn.Privmsg(conn.Config.IRCRoom, msg)
	return nil
}
