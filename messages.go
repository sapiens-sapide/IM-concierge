package IM_concierge

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/query/graphql"
	"github.com/satori/go.uuid"
	"github.com/thoj/go-ircevent"
	"net/http"
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
	var json_quads []byte
	json_quads, _ = json.Marshal(quads)
	http_body := bytes.NewReader(json_quads)
	_, err := http.Post(Config.CayleyAPI+"write", "application/json", http_body)
	if err != nil {
		log.Warnf("Error : %v", err)
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
