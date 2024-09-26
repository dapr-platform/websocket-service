package entity

import "encoding/json"

var (
	MESSAGE_TYPE_CONNECT               = "connect"
	MESSAGE_TYPE_DISCONNECT            = "disconnect"
	MESSAGE_TYPE_PING                  = "ping"
	MESSAGE_TYPE_PONG                  = "pong"
	MESSAGE_TYPE_TO_PUBSUB_TOPIC_TOPIC = "to_topic"      //前端将消息发送到服务器其他队列 ，可以不通过websocket消息发送，可以调用api发送。
	MESSAGE_TYPE_PROGRESS_RATE         = "progress_rate" //事件进度消息，例如上传模型时处理进度
)

type PeerMessage struct {
	Type         string `json:"type"`                    //"connect" | "disconnect" | "ping" | "pong" | "to_topic"
	Mark         string `json:"mark"`                    //消息主题（一级分类）需要跟后台约定标识，eg：告警(alarm)
	BusinessType string `json:"business_type,omitempty"` //业务主题(二级分类)
	ConnectId    string `json:"connect_id"`              //连接id
	Message      string `json:"message"`                 //消息字符串 json格式
}

func (p *PeerMessage) ToMap() (result map[string]any) {
	buf, _ := json.Marshal(p)
	_ = json.Unmarshal(buf, &result)
	return
}
