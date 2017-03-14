package imc

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"strings"
)

func HandleUsersList(conn *irc.Connection, e *irc.Event) {
	switch e.Code {
	case "353":
		//users list sent at connection
		users := strings.Split(e.Message(), " ")
		for _, user := range users {
			var clean_nick string
			switch user[0] {
			case '~', '@':
				clean_nick = user[1:]
			default:
				clean_nick = user
			}
			Users[clean_nick] = Recipient{
				DisplayName: clean_nick,
			}
		}
		for _, rcpt := range Users {
			conn.Who(rcpt.DisplayName)
		}
	case "JOIN":
		//a user is joining the channel
		var clean_nick string
		switch e.Nick[0] {
		case '~', '@':
			clean_nick = e.Nick[1:]
		default:
			clean_nick = e.Nick
		}
		if _, ok := Users[clean_nick]; !ok {
			Users[clean_nick] = Recipient{
				DisplayName: clean_nick,
			}
			conn.Who(clean_nick)
		}
	}

	fmt.Printf("Users list : %+v\n", Users)
}

func HandleWhoReply(e *irc.Event) {
	rcpt := Users[e.Arguments[5]]
	rcpt.Identifier = strings.Split(e.Message(), " ")[1]
	Users[e.Arguments[5]] = rcpt
}

func HandleLeavingUser(e *irc.Event) {

}
