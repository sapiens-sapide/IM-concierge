package main

import (
	"github.com/thoj/go-ircevent"
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/sapiens-sapide/chat-concierge"
)

const channel = "#chat-concierge-playground";
const serverssl = "barjavel.freenode.net:7000"

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	var err error
	//chat_concierge.Store, err = chat_concierge.InitDb()
	chat_concierge.Session, err = chat_concierge.InitSession()
	if err != nil {
		log.WithError(err).Fatalln("cayley.NewGraph error")
	}

	ircnick1 := "stansab_bot"
	irccon := irc.IRC(ircnick1, "chat-concierge")
	irccon.VerboseCallbackHandler = false
	irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("353", func(e *irc.Event) { chat_concierge.HandleUsersList(irccon, e)})
	irccon.AddCallback("352", chat_concierge.HandleWhoReply)
	irccon.AddCallback("PRIVMSG", chat_concierge.HandleMessage)
	err = irccon.Connect(serverssl)
	if err != nil {
		fmt.Printf("Err %s", err )
		return
	}
	irccon.GetNick()

	irccon.Loop()
}