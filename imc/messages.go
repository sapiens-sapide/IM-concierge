package imc

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/query/graphql"
	"github.com/cayleygraph/cayley/schema"
	"github.com/satori/go.uuid"
	"github.com/thoj/go-ircevent"
	"time"
)

func HandleMessage(e *irc.Event) {
	msg_id := uuid.NewV4()
	quads := []Quad{}
	quads = append(quads,
		Quad{
			e.Nick,
			"sent",
			msg_id.String(),
			"action",
		},
		Quad{
			msg_id.String(),
			"body",
			e.Message(),
			"property",
		},
		Quad{
			msg_id.String(),
			"date_sent",
			time.Now().Format(time.RFC3339Nano),
			"property",
		},
		Quad{
			msg_id.String(),
			"recipient",
			Config.IRCChannel,
			"property",
		})
	t := time.Now()
	/*from := []Recipient{
	Recipient{
		DisplayName: e.Nick,
		Identifier: e.User,
	},
	}*/
	msg := Message{
		Id:         quad.IRI(msg_id.String()),
		From:       []string{e.Nick},
		Body:       e.Message(),
		Year:       t.Year(),
		Month:      int(t.Month()),
		Day:        t.Day(),
		Hour:       t.Hour(),
		Minute:     t.Minute(),
		Second:     t.Second(),
		Nano:       t.Nanosecond(),
		DateString: t.Format(time.RFC3339Nano),
	}
	qw := graph.NewWriter(Store)
	_, err := schema.WriteAsQuads(qw, msg)
	if err != nil {
		log.Warnf("Error [WriteAsQuads] : %v", err)
	}
}

func InitDb() (store *cayley.Handle, err error) {
	graph.InitQuadStore("bolt", "/Users/stan/Dev/IM-concierge-playground/cayley.db", nil)
	store, err = cayley.NewGraph("bolt", "/Users/stan/Dev/IM-concierge-playground/cayley.db", nil)

	return store, err
}

func InitSession() (session *graphql.Session, err error) {
	/*	graph.InitQuadStore("bolt", "/Users/stan/Dev/IM-concierge-playground/cayley.db", nil)
		var store graph.QuadStore
		store, err = cayley.NewGraph("bolt", "/Users/stan/Dev/IM-concierge-playground/cayley.db", nil)
		session = graphql.NewSession(store)*/
	return session, err
}
