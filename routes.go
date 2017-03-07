package IM_concierge

import (
	"github.com/gin-gonic/gin"
	"github.com/cayleygraph/cayley/schema"
	log "github.com/Sirupsen/logrus"
	// Import RDF vocabulary definitions to be able to expand IRIs like rdf:label.
	_ "github.com/cayleygraph/cayley/voc/core"
	"net/http"
	"bytes"
	"fmt"
)

func AllMessagesHandler(ctx *gin.Context) {
	var messages []Message
	err := schema.LoadTo(nil, Store, &messages)
	log.Println("aiens")

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
	SortByDate(messages)
	var buf bytes.Buffer
	for _, message := range messages {
		buf.WriteString(fmt.Sprintf("Le %d à %d:%d:%d, %s : « %s ».\n",
			message.Day, message.Hour, message.Minute, message.Second, message.From, message.Body))
	}
	ctx.String(200, buf.String())

}

