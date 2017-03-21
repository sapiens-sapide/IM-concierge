package imc

import (
	"crypto/tls"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	m "github.com/sapiens-sapide/IM-concierge/models"
	"github.com/thoj/go-ircevent"
	"time"
)

func StartIRCservice() error {
	irccon := irc.IRC(Config.IRCNickname, Config.IRCUser)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(Config.IRCRoom)
	})
	irccon.AddCallback("353", func(e *irc.Event) {
		//HandleUsersList(irccon, e)
	})
	irccon.AddCallback("JOIN", func(e *irc.Event) {
		//HandleUsersList(irccon, e)
	})
	irccon.AddCallback("352", HandleWhoReply)
	irccon.AddCallback("PRIVMSG", CareTaker.HandleMessage)
	log.Infoln("going to connect to IRC")
	err := irccon.Connect(Config.IRCserver)
	if err != nil {
		return err
	}
	irccon.GetNick()

	go irccon.Loop()

	return nil
}

func (concierge *Concierge) HandleMessage(e *irc.Event) {
	msg := m.Message{
		Body: e.Message(),
		Date: time.Now(),
		From: e.Nick,
		Id:   m.NewULID(),
		Room: e.Arguments[0],
	}
	err := concierge.Backend.StoreMessage(msg)
	if err != nil {
		log.Warnf("Error [handleMessage.StoreMessage] : %v", err)
	}
	if len(concierge.Clients) > 0 {
		json_msg, err := json.Marshal(msg)
		if err != nil {
			log.Warnf("Error [handleMessage.jsonMarshal] : %v", err)
		} else {
			for _, client := range concierge.Clients {
				client.ToClient <- json_msg
			}
		}
	}
}
