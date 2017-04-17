package web_api

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
	"gopkg.in/gin-gonic/gin.v1"
	"html/template"
	"net/http"
	"sync"
	"time"
)

const defaultMessagesDaysAge = 7 // maximum days to rollback to for unfiltered messages list

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	frWeekDays = [7]string{"Dimanche", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi"}
)

type WebApi interface {
	RouteMessagesHdl(ctx *gin.Context)
	WsHandler(w http.ResponseWriter, r *http.Request)
	WsClientHandler(client *FrontClient)
	Start() error
	BroadcastMessage([]byte)
	Shutdown() error
}

type FrontServer struct {
	ClientsMux sync.Mutex
	Clients    []*FrontClient // users currently connected to the front server
	PopClient  chan int       //chan to receive disconnection order from concierge
	Backend    backend.ConciergeBackend
	Memory     backend.MemoryDB
}

func InitFrontServer(conf ConciergeConfig, b backend.ConciergeBackend, m backend.MemoryDB) (fs *FrontServer, err error) {
	fs = &FrontServer{
		ClientsMux: sync.Mutex{},
		Clients:    []*FrontClient{},
		PopClient:  make(chan int),
		Backend:    b,
		Memory:     m,
	}
	return
}

func (fs *FrontServer) Start() error {

	router := gin.Default()
	funcMap := template.FuncMap{
		"weekdayDateStr": DateToWeekdayString,
	}
	msgListTmpl, err := template.New("msgListTmpl").Funcs(funcMap).ParseFiles("../../UI-web/templates/messages.tmpl")
	if err != nil {
		return err
	}
	router.SetHTMLTemplate(msgListTmpl)
	//router.LoadHTMLGlob("../../front/templates/*")
	// adds our middlewares

	router.Static("/static/", "../../UI-web/static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Instant Messaging Concierge",
		})
	})

	// adds our routes and handlers
	api := router.Group("/messages")
	api.GET("/", fs.RouteMessagesHdl)
	api.GET("/ws", func(c *gin.Context) {
		fs.WsHandler(c.Writer, c.Request)
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

func (fs *FrontServer) BroadcastMessage(msg []byte) {
	log.Info(string(msg))
	for _, client := range fs.Clients {
		client.ToClient <- msg
	}
}

func DateToWeekdayString(t time.Time) string {
	return fmt.Sprintf("%s %s", frWeekDays[t.Weekday()], t.Format("à 15:04:05"))
}

func (fs *FrontServer) Shutdown() error {
	//TODO: shutdown gracefully. See github.com/gin-gonic/gin/issues/296
	return nil
}
