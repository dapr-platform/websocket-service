package handler

import "sync"
import "github.com/gorilla/websocket"

type ThreadSafeConn struct {
	*websocket.Conn
	sync.Mutex
}

func (t *ThreadSafeConn) ReadMessage() (messageType int, p []byte, err error) {
	return t.Conn.ReadMessage()
}
func (t *ThreadSafeConn) WriteMessage(messageType int, data []byte) error {
	t.Lock()
	defer t.Unlock()
	return t.Conn.WriteMessage(messageType, data)
}
