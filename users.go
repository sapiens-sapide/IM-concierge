package chat_concierge

import (
	"github.com/thoj/go-ircevent"
	"strings"
	"fmt"
)

func HandleUsersList(conn *irc.Connection, e *irc.Event) {
	users := strings.Split(e.Message(), " ")
	connected = make(map[string]Recipient)
	for _, user := range users {
		connected[user] = Recipient{
			Type: IRCnickname,
			DisplayName: user,
		}
	}
	for _, rcpt := range connected {
		conn.Who(rcpt.DisplayName)
	}
	fmt.Printf("connected : %+v", connected)
}

func HandleWhoReply (e *irc.Event) {
	rcpt := connected[e.Arguments[5]]
	rcpt.Identifier = strings.Split(e.Message(), " ")[1]
	connected[e.Arguments[5]] = rcpt
	fmt.Printf("Connected : %+v\n", connected)
}
