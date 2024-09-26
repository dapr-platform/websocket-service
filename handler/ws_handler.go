package handler

import (
	"context"
	"encoding/json"
	"github.com/dapr-platform/common"
	mycommon "github.com/dapr-platform/common"
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"websocket-service/entity"
	"websocket-service/eventpub"
)

type WsHandler struct {
	Id            string
	Key           string
	Mark          string
	ConnectId     string
	BusinessType  string
	MessageCenter *MessageCenter
	ThConn        *ThreadSafeConn
	SubscribedMap map[string]bool
	Mutex         *sync.Mutex
	EventChan     chan []byte
}

func NewWsHandler(messageCenter *MessageCenter, key string, conn *websocket.Conn) *WsHandler {
	s := &WsHandler{
		Id:            common.NanoId(),
		Key:           key,
		ConnectId:     "",
		MessageCenter: messageCenter,
		ThConn:        &ThreadSafeConn{conn, sync.Mutex{}},
		EventChan:     make(chan []byte, 10),
		SubscribedMap: make(map[string]bool, 0),
		Mutex:         &sync.Mutex{},
	}
	messageCenter.AddWsHandler(key, s)
	return s
}
func (s *WsHandler) Start() {
	//go s.startPingPong()
	go s.startReadPeerMsg()
	for buf := range s.EventChan {
		s.processCommonMessage(buf)
	}
}
func (s *WsHandler) processCommonMessage(buf []byte) (err error) {
	var msg common.CommonMessage
	err = json.Unmarshal(buf, &msg)
	if err != nil {
		mycommon.Logger.Error(err)
		return
	}
	subscribeKey := ""
	_, exists := msg[common.COMMON_MESSAGE_KEY_MARK]
	if exists {
		subscribeKey = msg[common.COMMON_MESSAGE_KEY_MARK].(string)
	}

	if msg[common.COMMON_MESSAGE_KEY_BUSINESS_TYPE] != nil && msg[common.COMMON_MESSAGE_KEY_BUSINESS_TYPE] != "" {
		subscribeKey = subscribeKey + "_" + msg[common.COMMON_MESSAGE_KEY_BUSINESS_TYPE].(string)
	}
	if msg[common.COMMON_MESSAGE_KEY_CONNECT_ID] != nil && msg[common.COMMON_MESSAGE_KEY_CONNECT_ID] != "" {
		subscribeKey = subscribeKey + "_" + msg[common.COMMON_MESSAGE_KEY_CONNECT_ID].(string)
	}
	_, exists = s.SubscribedMap[subscribeKey]

	if exists {
		err = s.ThConn.WriteMessage(websocket.TextMessage, buf)
		mycommon.Logger.Debugf("send peer msg: %s", string(buf))
		if err != nil {
			mycommon.Logger.Error(err)
			return
		}

	}
	return

}
func (s *WsHandler) startPingPong() {

	pingMsg := entity.PeerMessage{
		Type: entity.MESSAGE_TYPE_PING,
		Mark: "websocket-service",
	}
	buf, _ := json.Marshal(pingMsg)
	for {
		err := s.ThConn.WriteMessage(websocket.TextMessage, buf)
		if err != nil {
			return
		}
		time.Sleep(time.Second * 5)
	}

}
func (s *WsHandler) writePong() {

	pongMsg := common.CommonMessage{
		common.COMMON_MESSAGE_KEY_TYPE: common.COMMON_MESSAGE_TYPE_PONG,
	}

	buf, _ := json.Marshal(pongMsg)
	err := s.ThConn.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		common.Logger.Error(err)
		return
	}

}

func (s *WsHandler) close() {
	s.MessageCenter.RemoveSubHandler(s.Key, s)
	s.ThConn.Close()
	close(s.EventChan)
}
func (s *WsHandler) startReadPeerMsg() {
	for {
		msgType, data, err := s.ThConn.ReadMessage()
		if err != nil { //check websocket is closed
			mycommon.Logger.Debug("websocket closed", err)
			s.close()
			return
		}
		if len(data) == 0 {

			continue
		}
		if msgType != websocket.TextMessage {
			mycommon.Logger.Error("only use text message")
			continue
		}
		var msg entity.PeerMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			mycommon.Logger.Error(err)
			continue
		}
		if msg.Type == entity.MESSAGE_TYPE_CONNECT {
			subKey := msg.Mark
			if msg.BusinessType != "" {
				subKey = subKey + "_" + msg.BusinessType
			}
			if msg.ConnectId != "" {
				subKey = subKey + "_" + msg.ConnectId
			}
			s.Mutex.Lock()
			s.SubscribedMap[subKey] = true
			s.Mutex.Unlock()
			im := common.InternalMessage{}
			err = im.FromStruct(msg)
			if err != nil {
				mycommon.Logger.Debug("to internal msg err:  " + err.Error())
			} else {
				im.SetType(common.INTERNAL_MESSAGE_TYPE_WEB_CONNECT)
				err = eventpub.PublishInternalMessage(context.Background(), &im)
				if err != nil {
					mycommon.Logger.Debug("publish internal msg err:  " + err.Error())
				}
			}

			mycommon.Logger.Debug("receive connect msg  ", string(data))
		} else if msg.Type == entity.MESSAGE_TYPE_DISCONNECT {
			subKey := msg.Mark
			if msg.BusinessType != "" {
				subKey = subKey + "_" + msg.BusinessType
			}
			if msg.ConnectId != "" {
				subKey = subKey + "_" + msg.ConnectId
			}
			s.Mutex.Lock()
			delete(s.SubscribedMap, subKey)
			s.Mutex.Unlock()
			im := common.InternalMessage{}
			err = im.FromStruct(msg)
			if err != nil {
				mycommon.Logger.Debug("to internal msg err:  " + err.Error())
			} else {
				im.SetType(common.INTERNAL_MESSAGE_TYPE_WEB_DISCONNECT)
				err = eventpub.PublishInternalMessage(context.Background(), &im)
				if err != nil {
					mycommon.Logger.Debug("publish internal msg err:  " + err.Error())
				}
			}

		} else if msg.Type == entity.MESSAGE_TYPE_PONG {
			//TODO
		} else if msg.Type == entity.MESSAGE_TYPE_PING {
			s.writePong()
		}
	}
}
