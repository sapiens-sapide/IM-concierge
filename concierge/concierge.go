// Package concierge manages IM connectors and front server
// and dispatches events between componants
package concierge

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sapiens-sapide/IM-concierge/API-web"
	im_conn "github.com/sapiens-sapide/IM-concierge/IM-connectors"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
)

type Concierge struct {
	Backend      backend.ConciergeBackend //backend to persist & retreive data
	Config       ConciergeConfig
	FrontServer  web_api.WebApi      //serve web UI & handle users interactions
	Hatch        chan Notification   //chan to receive events from componants
	IMConnectors []im_conn.Connector //connectors handle connections to IM servers
	Memory       backend.MemoryDB    //backend to handle runtime data
}

// Load configurations, load backends, initialize componants. No service is running at this stage
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

	// Init Instant Messaging connectors
	concierge.IMConnectors, err = im_conn.InitConnectors(concierge.Config, concierge.Backend, concierge.Hatch)
	if err != nil {
		log.Warn("Connectors initialization failed")
		return nil, err
	}

	// Init front server
	concierge.FrontServer, err = web_api.InitFrontServer(concierge.Config, concierge.Backend, concierge.Memory)
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

	for _, connector := range c.IMConnectors {
		err = connector.Start()
		if err != nil {
			return err
		}
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
		}
	}
}

func (c *Concierge) Shutdown() error {
	c.FrontServer.Shutdown()
	for _, connector := range c.IMConnectors {
		connector.Shutdown()
	}
	c.Backend.Shutdown()
	return nil
}
