package chat_concierge

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/query/graphql"
	"time"
)

type (
	MessageType    uint8
	AttachmentType uint8
	IdentifierType uint8
)

const (
	SMTPmail MessageType = iota
	IRCmsg
)
const (
	Image AttachmentType = iota
	URL
)
const (
	EmailAdd IdentifierType = iota
	IRCnickname
)

var connected map[string]Recipient
var Store *cayley.Handle
var Session *graphql.Session

type Message struct {
	Attachments  []Attachment
	Body         string
	Date         time.Time
	DiscussionId []byte
	Id           []byte
	Identities   []Identity
	RawMsgId     []byte
	Recipients   []Recipient
	Subject      string
	UserId       []byte
	Type         MessageType
}

type Attachment struct {
	Blob_id  []byte
	Type     AttachmentType
	Name     string
	Size     int
	Cid      string
	IsInline bool
}

type Identity struct {
	Recipient
	UserId []byte
}

type Recipient struct {
	Identifier  string
	Type        IdentifierType
	DisplayName string
}

type Quad struct {
	Subject   string
	Predicate string
	Object    string
	Label     string
}
