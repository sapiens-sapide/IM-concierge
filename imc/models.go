package imc

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/query/graphql"
	"time"
	// Import RDF vocabulary definitions to be able to expand IRIs like rdf:label.
	_ "github.com/cayleygraph/cayley/voc/core"
	"sort"
)

type (
	AttachmentType uint8
	MessageType    uint8
	ProtocolType   uint8
	RecipientType  uint8
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
	From RecipientType = iota
	To
	Cc
	Bcc
)

const (
	Email ProtocolType = iota
	IRC
	XMPP
)

var (
	Config  ConciergeConfig
	Users   map[string]Recipient
	Store   *cayley.Handle
	Session *graphql.Session
)

type Message struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType struct{} `quad:"@type > Message"`
	Body    string   `json:"body"`
	From    []string `json:"from"`
	Id      quad.IRI `json:"@id"`
	//date fields
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	Hour       int    `json:"hour"`
	Minute     int    `json:"minute"`
	Second     int    `json:"second"`
	Nano       int    `json:"nano"`
	DateString string `json:"full_date"`
}
type Message_all struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType     struct{}     `quad:"@type > Message"`
	Attachments []Attachment `json:"attachment"`
	Bcc         []Recipient  `json:"bcc"`
	Body        string       `json:"body"`
	Cc          []Recipient  `json:"cc"`
	From        []string     `json:"from"`
	Id          quad.IRI     `json:"@id"`
	Identities  []Identity   `json:"identity"`
	RawMsgId    []byte       `json:"raw_msg_id"`
	Subject     string       `json:"subject"`
	Tags        []Tag        `json:"tag"`
	Type        MessageType  `json:"type"`
	UserId      []byte       `json:"user_id"`
	//date fields
	Year       int       `json:"year"`
	Month      int       `json:"month"`
	Day        int       `json:"day"`
	Hour       int       `json:"hour"`
	Minute     int       `json:"minute"`
	Second     int       `json:"second"`
	Nano       int       `json:"nano"`
	Date       time.Time `json:"—"`
	DateString string    `json:"full_date"`
}

type Attachment struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType  struct{}       `quad:"@type > Attachment"`
	Blob_id  []byte         `json:"blob_id"`
	Type     AttachmentType `json:"type"`
	Name     string         `json:"name"`
	Size     int            `json:"size"`
	Cid      quad.IRI       `json:"@id"`
	IsInline bool           `json:"isInline"`
}

type Identity struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType struct{} `quad:"@type > ex:Identity"`
	Id      quad.IRI `json:"@id"`
	Recipient
	UserId []byte `json:"user_id"`
}

type Recipient struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType     struct{} `quad:"@type > Recipient"`
	DisplayName string   `json:"display_name"`
	Id          quad.IRI `json:"@id"`
	Identifier  string   `json:"identifier"`
}

type Recipient_all struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType     struct{}      `quad:"@type > Recipient"`
	DisplayName string        `json:"display_name"`
	Id          quad.IRI      `json:"@id"`
	Identifier  string        `json:"identifier"`
	Protocol    ProtocolType  `json:"protocol"`
	Type        RecipientType `json:"type"`
}

type Tag string

type Quad struct {
	Subject   string
	Predicate string
	Object    string
	Label     string
}

type ConciergeConfig struct {
	IRCserver   string `mapstructure:"irc_server"`
	IRCChannel  string `mapstructure:"irc_channel"`
	IRCUser     string `mapstructure:"irc_user"`
	IRCNickname string `mapstructure:"irc_nickname"`
	CayleyAPI   string `mapstructure:"cayley_api_path"`
}

type Discussion struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType     struct{}      `quad:"@type > Discussion"`
	Id          quad.IRI      `json:"@id"`
	Nodes       []MessageNode `json:"node"`
	SubjectHash string        `json:"@subject_hash"`
}

type MessageNode struct {
	// dummy field to enforce all object to have a <id> <rdf:type> <ex:Message> relation
	// means nothing for Go itself
	rdfType      struct{} `quad:"@type > MessageNode"`
	Id           quad.IRI `json:"@id"`
	MessageId    []byte   `json:"message_id"`
	NextMessages [][]byte `json:"next_msg"`
	PrevMessages [][]byte `json:"prev_msg"`
	SeqNumber    int      `json:"seq_number"`
}

type Date struct {
}

/*** sort functions for Messages ***/
// By is the type of a "less" function that defines the ordering of its Message arguments.
type By func(m1, m2 *Message) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(messages []Message) {
	ps := &messageSorter{
		messages: messages,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type messageSorter struct {
	messages []Message
	by       func(m1, m2 *Message) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *messageSorter) Len() int {
	return len(s.messages)
}

// Swap is part of sort.Interface.
func (s *messageSorter) Swap(i, j int) {
	s.messages[i], s.messages[j] = s.messages[j], s.messages[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *messageSorter) Less(i, j int) bool {
	return s.by(&s.messages[i], &s.messages[j])
}

func SortByDate(messages []Message) {
	// Closures that order the Message structure in descending order
	date := func(m1, m2 *Message) bool {
		return m1.DateString > m2.DateString
	}
	By(date).Sort(messages) //type conversion
}
