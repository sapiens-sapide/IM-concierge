package front

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cayleygraph/cayley/schema"
	"github.com/gin-gonic/gin"
	// Import RDF vocabulary definitions to be able to expand IRIs like rdf:label.
	"bytes"
	"fmt"
	_ "github.com/cayleygraph/cayley/voc/core"
	"github.com/gorilla/websocket"
	"github.com/sapiens-sapide/IM-concierge/imc"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err)
		return
	}
	newClient := imc.FrontClient{
		Ws:          conn,
		Identity:    imc.Identity{},
		FromClient:  make(chan []byte),
		ToClient:    make(chan []byte),
		LeaveClient: make(chan bool),
	}
	go newClient.WsClientHandler()
}

func AllMessagesHandler(ctx *gin.Context) {
	var messages []imc.Message
	err := schema.LoadTo(nil, imc.Store, &messages)
	if err != nil {
		log.Warnf("Error [LoadTo] : %v", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
	/*
		p := path.NewPath(Store).Has(quad.String("hour"), quad.Int(16))
		err = p.Iterate(nil).EachValue(Store, func(v quad.Value) {
			log.Println("hour", v.Native())
		})
		if err != nil {
			log.WithError(err)
		}
	*/
	imc.SortByDate(messages)
	var buf bytes.Buffer
	for _, message := range messages {
		buf.WriteString(fmt.Sprintf("Le %d à %d:%d:%d, %s : « %s ».\n",
			message.Day, message.Hour, message.Minute, message.Second, message.From, message.Body))
	}
	ctx.String(200, buf.String())

}
