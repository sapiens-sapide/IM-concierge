package backend

import (
	m "github.com/sapiens-sapide/IM-concierge/models"
	"time"
)

type ConciergeBackend interface {
	StoreMessage(msg m.Message) error

	/*
		ListMessagesByDate returns an array of date ordered messages
		ranging from now to dateLimit
	*/
	ListMessagesByDate(channel string, dateLimit time.Time) ([]m.Message, error)

	/*
		Store room's name to ensure it exists into db
	*/
	RegisterRoom(room string) (roomId string, err error) //should returns known users also

	Shutdown() error
}
