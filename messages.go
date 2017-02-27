package chat_concierge

import (
	"github.com/thoj/go-ircevent"
	"github.com/satori/go.uuid"
	"fmt"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/query/graphql"
	"net/http"
	"encoding/json"
	"bytes"
	log "github.com/Sirupsen/logrus"
)

func HandleMessage(e *irc.Event) {

	msg := Message{
		Id: uuid.NewV4().Bytes(),
		Body: e.Message(),
	}
	fmt.Printf("Arguments : %s\n", e.Arguments)
	fmt.Printf("code : %d\n", e.Code)
	fmt.Printf("host : %s\n", e.Host)
	fmt.Printf("message : %s\n", e.Message())
	fmt.Printf("nick : %s\n", e.Nick)
	fmt.Printf("raw : %s\n", e.Raw)
	fmt.Printf("source : %s\n", e.Source)
	fmt.Printf("user : %s\n", e.User)

	fmt.Printf("%+v", msg)
	//Store.AddQuad(quad.Make(e.Nick, "sent", e.Message(), nil))
	quad := Quad{
		e.Nick,
		"sent",
		e.Message(),
		"",
	}
	var json_quad bytes.Buffer
	json_quad.WriteRune('[')
	enc := json.NewEncoder(&json_quad)
	enc.Encode(quad)
	json_quad.WriteRune(']')

	resp, err := http.Post("http://127.0.0.1:64210/api/v1/write", "application/json", &json_quad)
	if err != nil {
		log.Warnf("Error : %v", err)
	}
	log.Infof("response : %s", resp)
}

func InitDb() (store *cayley.Handle, err error){
	graph.InitQuadStore("bolt", "/Users/stan/Dev/chat-concierge-playground/cayley.db", nil)
	store, err = cayley.NewGraph("bolt", "/Users/stan/Dev/chat-concierge-playground/cayley.db", nil)

	return store, err
}

func InitSession() (session *graphql.Session, err error) {
/*	graph.InitQuadStore("bolt", "/Users/stan/Dev/chat-concierge-playground/cayley.db", nil)
	var store graph.QuadStore
	store, err = cayley.NewGraph("bolt", "/Users/stan/Dev/chat-concierge-playground/cayley.db", nil)
	session = graphql.NewSession(store)*/
	return session, err
}