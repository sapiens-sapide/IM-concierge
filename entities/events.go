package entities

// Event types for inter-componants notifications
type EventType int

const (
	ShutDown         EventType = iota // All done, please stop working asap
	NewMessage                        // A message is published into a room
	ClientConnect                     // A new client connects to the front end
	ClientLeave                       // A client disconnects from the front end
	ParticipantJoin                   // A new user joins a room
	ParticipantLeave                  // An user leaves a room
	ClientPostMessage
)

type Notification interface {
	EventType() EventType // returns the event type of the notification
	Payload() (interface{}, error)
}
