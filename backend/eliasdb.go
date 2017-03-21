package backend

import (
	"devt.de/eliasdb/eql"
	"devt.de/eliasdb/graph"
	"devt.de/eliasdb/graph/data"
	"devt.de/eliasdb/graph/graphstorage"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/oklog/ulid"
	m "github.com/sapiens-sapide/IM-concierge/models"
	"sync"
	"time"
)

type EliasBackend struct {
	gm             *graph.Manager
	MsgPointerMux  sync.Mutex
	lastMsgPointer ulid.ULID
}

const billion = 1000000000

func InitEliasBackend() (eb *EliasBackend, err error) {

	eb = &EliasBackend{}
	// Open or create a graph storage
	gs, err := graphstorage.NewDiskGraphStorage("../../data", false)
	if err != nil {
		return nil, err
	}

	eb.gm = graph.NewGraphManager(gs)

	//get a ref to the last stored message
	lastMsg, _ := eb.gm.FetchNode("main", "0", "lastMessagePointer")
	if err != nil {
		return nil, err
	}

	if lastMsg == nil {
		//try to (re)build the lastMessagePointer from messages list
		res, err := eql.RunQuery("LastMessageByDate", "main", "lookup message with ordering(descending key)", eb.gm)
		if err == nil && res.RowCount() > 0 {
			eb.MsgPointerMux.Lock()
			eb.lastMsgPointer, _ = ulid.Parse(res.Row(0)[0].(string))
			eb.MsgPointerMux.Unlock()
		}
	} else {
		eb.MsgPointerMux.Lock()
		eb.lastMsgPointer, _ = ulid.Parse(lastMsg.Attr("ulid").(string))
		eb.MsgPointerMux.Unlock()
	}

	//ensure 'room' node kind exists in db
	emptyRoom, _ := eb.gm.FetchNode("main", "0", "room")
	if err != nil {
		return nil, err
	}
	if emptyRoom == nil {
		room_node := data.NewGraphNodeFromMap(map[string]interface{}{
			"kind": "room",
			"key":  "0",
			"name": "empty_room",
		})
		eb.gm.StoreNode("main", room_node)
	}

	return eb, nil
}

func (eb *EliasBackend) Shutdown() error {
	eb.MsgPointerMux.Lock()
	msgPointer := data.NewGraphNodeFromMap(map[string]interface{}{
		"kind": "lastMessagePointer",
		"key":  "0",
		"ulid": eb.lastMsgPointer.String(),
	})
	eb.gm.StoreNode("main", msgPointer)
	eb.MsgPointerMux.Unlock()
	return nil
}

func (eb *EliasBackend) StoreMessage(msg m.Message) error {

	// Create transaction
	trans := graph.NewGraphTrans(eb.gm)

	// Store msgNode
	eb.MsgPointerMux.Lock()
	defer eb.MsgPointerMux.Unlock()
	msgNode := data.NewGraphNodeFromMap(map[string]interface{}{
		"kind":    "message",
		"key":     msg.Id.String(),
		"body":    msg.Body,
		"date":    msg.Date.UnixNano(),
		"from":    msg.From,
		"prevMsg": eb.lastMsgPointer.String(),
	})

	err := trans.StoreNode("main", msgNode)
	if err != nil {
		return err
	}
	var emptyULID ulid.ULID
	if eb.lastMsgPointer != emptyULID {

		//link msg to previous one
		edge := data.NewGraphEdge()
		edge.SetAttr(data.NodeKey, msg.Id.String()+":"+eb.lastMsgPointer.String())
		edge.SetAttr(data.NodeKind, "MessageChain")
		edge.SetAttr(data.EdgeEnd1Key, msg.Id.String())
		edge.SetAttr(data.EdgeEnd1Kind, "message")
		edge.SetAttr(data.EdgeEnd1Role, "previousMsg")
		edge.SetAttr(data.EdgeEnd1Cascading, false)
		edge.SetAttr(data.EdgeEnd2Key, eb.lastMsgPointer.String())
		edge.SetAttr(data.EdgeEnd2Kind, "message")
		edge.SetAttr(data.EdgeEnd2Role, "nextMsg")
		edge.SetAttr(data.EdgeEnd2Cascading, false)

		err = trans.StoreEdge("main", edge)
		if err != nil {
			return err
		}
		//link msg to room
		res, err := eql.RunQuery("getRoomByName", "main", fmt.Sprintf("get room where name = '%s'", msg.Room), eb.gm)
		if err == nil && res.RowCount() == 1 {
			edge_room := data.NewGraphEdge()
			edge_room.SetAttr(data.NodeKey, msg.Id.String()+":"+res.Row(0)[0].(string))
			edge_room.SetAttr(data.NodeKind, "inRoom")
			edge_room.SetAttr(data.EdgeEnd1Key, msg.Id.String())
			edge_room.SetAttr(data.EdgeEnd1Kind, "message")
			edge_room.SetAttr(data.EdgeEnd1Role, "inRoom")
			edge_room.SetAttr(data.EdgeEnd1Cascading, false)
			edge_room.SetAttr(data.EdgeEnd2Key, res.Row(0)[0].(string))
			edge_room.SetAttr(data.EdgeEnd2Kind, "room")
			edge_room.SetAttr(data.EdgeEnd2Role, "room")
			edge_room.SetAttr(data.EdgeEnd2Cascading, false)

			err = trans.StoreEdge("main", edge_room)
			if err != nil {
				return err
			}
		} else {
			log.WithError(err).Warnf("[eliasDB.StoreMessage] can't link to room %s (count %d)", msg.Room, res.RowCount())
			return err
		}

		//link msg to participant(s)

	}
	err = trans.Commit()
	if err != nil {
		return err
	}
	eb.lastMsgPointer = msg.Id
	return nil
}

func (eb *EliasBackend) ListMessagesByDate(room string, dateLimit time.Time) ([]m.Message, error) {

	msgList := []m.Message{}

	queryStr := fmt.Sprintf("get message where date > %d with ordering(descending date)", dateLimit.UnixNano())
	res, err := eql.RunQuery("MessagesByDate", "main", queryStr, eb.gm)
	if err != nil {
		return msgList, err
	}
	for _, row := range res.Rows() {
		date := time.Unix(int64(row[2].(int64)/billion), int64(row[2].(int64)%billion))
		id, _ := ulid.Parse(row[0].(string))
		prevId, _ := ulid.Parse(row[4].(string))

		msgList = append(msgList, m.Message{
			Body:    row[1].(string),
			Date:    date,
			Id:      id,
			From:    row[3].(string),
			PrevMsg: prevId,
		})
	}
	return msgList, nil
}

func (eb *EliasBackend) RegisterRoom(room string) (roomId string, err error) {
	//TODO: returns known users in that room
	res, err := eql.RunQuery("getRoomByName", "main", fmt.Sprintf("get room where name = '%s'", room), eb.gm)
	if err != nil {
		return
	}
	if res.RowCount() > 1 {
		err = errors.New(fmt.Sprintf("more than one room found in db with name <%s>", room))
		return
	}
	if res.RowCount() == 0 {
		roomModel := data.NewGraphNodeFromMap(map[string]interface{}{
			"kind": "room",
			"key":  m.NewULID().String(),
			"name": room,
		})
		err = eb.gm.StoreNode("main", roomModel)
		if err != nil {
			return
		}
		roomId = roomModel.Attr("key").(string)
		return
	}

	roomId = res.Row(0)[0].(string)
	return
}
