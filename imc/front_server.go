package imc

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const defaultMessagesDaysAge = 7 // maximum days to rollback to for unfiltered messages list
var frWeekDays = [8]string{"", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi", "Dimanche"}

func StartFrontServer() error {
	router := gin.Default()
	funcMap := template.FuncMap{
		"weekdayDateStr": DateToWeekdayString,
	}
	msgListTmpl, err := template.New("msgListTmpl").Funcs(funcMap).ParseFiles("../../front/templates/messages.tmpl")
	if err != nil {
		return err
	}
	router.SetHTMLTemplate(msgListTmpl)
	//router.LoadHTMLGlob("../../front/templates/*")
	// adds our middlewares

	router.Static("/static/", "../../front/static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Instant Messaging Concierge",
		})
	})

	// adds our routes and handlers
	api := router.Group("/messages")
	api.GET("/", CareTaker.AllMessagesHandler)
	api.GET("/ws", func(c *gin.Context) {
		CareTaker.WsHandler(c.Writer, c.Request)
	})

	// listens
	addr := "localhost:8080"
	go func() {
		err = router.Run(addr)
		if err != nil {
			log.WithError(err).Warn("unable to start gin server")
		}
	}()

	return nil
}

func DateToWeekdayString(t time.Time) string {
	return fmt.Sprintf("%s %s", frWeekDays[t.Weekday()], t.Format("à 15:04:05"))
}
