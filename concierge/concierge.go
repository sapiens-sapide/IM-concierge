// Package concierge manages IM connectors and front server
// and dispatches events between componants
package concierge

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sapiens-sapide/IM-concierge/API-web"
	im_conn "github.com/sapiens-sapide/IM-concierge/IM-connectors"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
	"github.com/satori/go.uuid"
)

type Concierge struct {
	Backend     backend.ConciergeBackend //backend to persist & retreive data
	Config      ConciergeConfig
	FrontServer web_api.WebApi    //serve web UI & handle users interactions
	Hatch       chan Notification //chan to receive events from componants
	IMHandler   im_conn.IMHandler //handle connections to IM servers
	Memory      backend.MemoryDB  //backend to handle runtime data
}

// Load configurations, load backends, initialize componants.
func NewConcierge(conf ConciergeConfig) (concierge *Concierge, err error) {
	be, err := backend.InitEliasBackend(conf.Backend)
	if err != nil {
		log.WithError(err).Fatalln("failed to init Elias backend")
	}

	concierge = &Concierge{
		Backend: be,
		Config:  conf,
		Hatch:   make(chan Notification),
		Memory:  nil, //TODO
	}

	// Init Instant Messaging handler
	concierge.IMHandler, err = im_conn.InitIMHandler(concierge.Config, concierge.Backend, concierge.Hatch)
	if err != nil {
		log.Warn("Connectors initialization failed")
		return nil, err
	}

	// Init front server
	concierge.FrontServer, err = web_api.InitFrontServer(concierge.Config, concierge.Backend, concierge.Memory, concierge.Hatch)
	if err != nil {
		log.Warn("Front server initialization failed")
		return nil, err
	}

	return concierge, nil
}

// Effectively start running componants
func (c *Concierge) Start() error {
	err := c.FrontServer.Start()
	if err != nil {
		return err
	}

	go c.eventsBus()

	return nil
}

// Dispacth events between concierge's componants.
// Each componant (front server, connector…) handles its own events and may publish some of them
// eventsBus is in charge of forwarding published events to the relevant componant(s)
func (c *Concierge) eventsBus() {
	for evt := range c.Hatch {
		switch evt.EventType() {
		case NewMessage:
			payload, err := evt.Payload()
			if err == nil {
				go c.FrontServer.BroadcastMessage(payload.([]byte))
			}
		case ImpersonateUser:
			identity, _ := evt.Payload()
			connector_key := c.Config.IRCRoom + ":" + c.Config.Concierge.IRCNickname
			go c.IMHandler.Impersonate(connector_key, identity.(Identity))
		case ClientPostMessage:
			connector_key := c.Config.IRCRoom + ":" + c.Config.User.IRCNickname
			msg, _ := evt.Payload()
			go c.IMHandler.PostMessageFor(connector_key, msg.(string))
		case ClientLeave, StopImpersonateUser:
			identity, _ := evt.Payload()
			connector_key := c.Config.IRCRoom + ":" + identity.(Identity).DisplayName
			c.IMHandler.Remove(connector_key)
		case ClientImpersonated:
			payload, _ := evt.Payload()
			client_id, _ := uuid.FromString(payload.(string))
			go c.FrontServer.NotifyClient(client_id, "connected", "")
		case ImpersonateFailed:
			payload, _ := evt.Payload()
			client_id, _ := uuid.FromString(payload.(string))
			go c.FrontServer.NotifyClient(client_id, "impersonate failed", "")
		}
	}
}

func (c *Concierge) Shutdown() error {
	c.FrontServer.Shutdown()
	c.IMHandler.Shutdown()
	c.Backend.Shutdown()
	return nil
}
