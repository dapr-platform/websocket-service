package handler

import (
	"context"
	"encoding/json"
	mycommon "github.com/dapr-platform/common"
	"github.com/dapr/go-sdk/service/common"
	"sync"
)

type MessageCenter struct {
	WsConnMap  *sync.Map
	PubConnMap *sync.Map
	connLock   *sync.Mutex
}

func NewMessageCenter() *MessageCenter {
	return &MessageCenter{
		WsConnMap: &sync.Map{},
		connLock:  &sync.Mutex{},
	}
}

func (m *MessageCenter) EventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	mycommon.Logger.Debugf("event - PubsubName: %s, Topic: %s, ID: %s, PreNodeResult: %s\n", e.PubsubName, e.Topic, e.ID, e.Data)
	key := "sub_" + e.PubsubName + "_" + e.Topic
	buf, err := json.Marshal(e.Data)
	if err != nil {
		mycommon.Logger.Debug(err)
		return false, err
	}

	var arr []*WsHandler
	val, exists := m.WsConnMap.Load(key)

	if exists {
		arr = val.([]*WsHandler)
		mycommon.Logger.Debug("wsHandler len=", len(arr))
		for _, h := range arr {
			sendToCh(buf, h.EventChan)
		}
	}
	return false, nil
}
func sendToCh(buf []byte, ch chan []byte) error {
	defer func() {
		if info := recover(); info != nil {
			mycommon.Logger.Error("sendToCh recover ", info)
		}
	}()
	ch <- buf
	return nil
}

func (m *MessageCenter) AddWsHandler(key string, handler *WsHandler) {
	m.connLock.Lock()
	defer m.connLock.Unlock()
	var arr []*WsHandler
	_, exist := m.WsConnMap.Load(key)
	if !exist {
		arr = make([]*WsHandler, 0)

	} else {
		v, _ := m.WsConnMap.Load(key)
		arr = v.([]*WsHandler)
	}
	arr = append(arr, handler)
	m.WsConnMap.Store(key, arr)
}

func (m *MessageCenter) RemoveSubHandler(key string, handler *WsHandler) {
	m.connLock.Lock()
	defer m.connLock.Unlock()
	var arr []*WsHandler
	_, exist := m.WsConnMap.Load(key)
	if !exist {
		mycommon.Logger.Error("can't find key", key)
		return

	} else {
		v, _ := m.WsConnMap.Load(key)
		arr = v.([]*WsHandler)
	}
	j := 0
	for _, v := range arr {
		if v.Id != handler.Id {
			arr[j] = v
			j++

		}
	}
	m.WsConnMap.Store(key, arr[:j])
}
