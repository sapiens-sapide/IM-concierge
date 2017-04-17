package im_connectors

import (
	"github.com/sapiens-sapide/IM-concierge/IM-connectors/IRC"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
)

type Connector interface {
	NotifyConcierge(Notification) error
	Start() error
	Shutdown() error
}

func InitConnectors(conf ConciergeConfig, backend backend.ConciergeBackend, hatch chan Notification) (connectors []Connector, err error) {
	// for now, only one IRC connector
	ircconn, err := irc.InitIRCconnector(conf, backend, hatch)
	if err != nil {
		return
	}
	connectors = append(connectors, ircconn)
	return
}
