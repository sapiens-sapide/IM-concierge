package web_api

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"time"
)

// handler for /messages route
func (fs *FrontServer) RouteMessagesHdl(ctx *gin.Context) {
	messages, err := fs.Backend.ListMessagesByDate("#im-concierge-playground", time.Now().AddDate(0, 0, -defaultMessagesDaysAge))
	if err != nil {
		log.WithError(err).Warnln("[FrontServer] : error when calling Backend.ListMessagesByDate")
	}

	ctx.HTML(http.StatusOK, "messages.tmpl", messages)
}
