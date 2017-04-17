package backend

import (
	. "github.com/sapiens-sapide/IM-concierge/entities"
	"time"
)

type ConciergeBackend interface {
	StoreMessage(msg Message) error

	/*
		ListMessagesByDate returns an array of date ordered messages
		ranging from now to dateLimit
	*/
	ListMessagesByDate(channel string, dateLimit time.Time) ([]Message, error)

	/*
		Store room's name to ensure it exists into db
	*/
	RegisterRoom(room string) (roomId string, err error) //should returns known users also

	Shutdown() error
}

type MemoryDB interface {
}
