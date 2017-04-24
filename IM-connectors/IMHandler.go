package im_handler

import (
	"errors"
	"fmt"
	"github.com/sapiens-sapide/IM-concierge/IM-connectors/IRC"
	"github.com/sapiens-sapide/IM-concierge/backend"
	. "github.com/sapiens-sapide/IM-concierge/entities"
)

type IMHandler interface {
	Shutdown() error
	Impersonate(string, Identity) error //for now, IRC assumed
	AddConcierge() error                //for now, IRC assumed
	PostMessageFor(string, string) error
	Remove(connector_key string) error // close the underlying IM connection
}

type IMconnector interface {
	Impersonate(Identity) (string, error)
	AddConcierge(Identity) (string, error)
	NotifyConcierge(Notification) error
	Close() error
}

// handle IRC only for now. TODO
type InstantMessagingHandler struct {
	Connectors     map[string]IMconnector // for now, map key is room:nickname, or room:im-concierge for concierge connector. TODO: handle nickname conflict
	Config         ConciergeConfig
	Backend        backend.ConciergeBackend
	ConciergeHatch chan Notification
}

func InitIMHandler(conf ConciergeConfig, backend backend.ConciergeBackend, hatch chan Notification) (imh *InstantMessagingHandler, err error) {
	imh = new(InstantMessagingHandler)
	imh.Config = conf
	imh.Backend = backend
	imh.ConciergeHatch = hatch
	imh.Connectors = make(map[string]IMconnector)
	//only one IRC concierge for now
	imh.AddConcierge()
	return
}

// for now, only IRC concierge
func (imh *InstantMessagingHandler) AddConcierge() error {
	irc_conf := IRCconfig{
		IRCserver: imh.Config.IRCserver,
		IRCRoom:   imh.Config.IRCRoom,
	}
	irc_conn, _ := irc.NewIRCconnector(imh.Backend, imh.ConciergeHatch, irc_conf)
	mapkey, err := irc_conn.AddConcierge(Identity{
		DisplayName: imh.Config.Concierge.IRCNickname,
		Identifier:  imh.Config.Concierge.IRCUser,
	})
	if err != nil {
		return err
	}
	imh.Connectors[mapkey] = irc_conn
	return nil
}

// for now, IRC only
// impersonate the provided identity into the room already handled by the provided connectors key
func (imh *InstantMessagingHandler) Impersonate(connector_key string, user Identity) error {
	if connector, ok := imh.Connectors[connector_key]; ok {
		switch c := connector.(type) {
		case *irc.IRCconnector:
			new_connector, _ := irc.NewIRCconnector(c.Backend, c.ConciergeHatch, IRCconfig{
				IRCserver: c.Config.IRCserver,
				IRCRoom:   c.Config.IRCRoom,
			})

			key, err := new_connector.Impersonate(user)
			if err != nil {
				return err
			}
			imh.Connectors[key] = new_connector
			return nil

		default:
			return errors.New(fmt.Sprintf("connector « %s » has an unknown type", connector_key))
		}

	}
	return errors.New(fmt.Sprintf("connector « %s » not registred", connector_key))
}

func (imh *InstantMessagingHandler) PostMessageFor(connector_key string, msg string) error {
	if connector, ok := imh.Connectors[connector_key]; ok {
		switch c := connector.(type) {
		case *irc.IRCconnector:
			return c.PostMessage(msg)
		default:
			return errors.New(fmt.Sprintf("connector « %s » has an unknown type", connector_key))
		}
	}
	return errors.New(fmt.Sprintf("connector « %s » not registred", connector_key))
}

func (imh *InstantMessagingHandler) Remove(connector_key string) error {
	if connector, ok := imh.Connectors[connector_key]; ok {
		connector.Close()
		delete(imh.Connectors, connector_key)
	}
	return errors.New(fmt.Sprintf("connector « %s » not registred", connector_key))
}

func (imh *InstantMessagingHandler) Shutdown() error {
	for _, connector := range imh.Connectors {
		connector.Close()
	}
	return nil
}
