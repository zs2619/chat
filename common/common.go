package common

const GlobalMessageChannel = "globalMessageChannel"
const RTMMessageChannel = "rtmMessageChannel"

type NotifyEvent struct {
	UserIDList []string `json:"userIDList"`
	Broadcast  bool     `json:"broadcast"` //true 广播 false发送userId
	EventType  int32    `json:"eventID"`
	Msg        string   `json:"msg"`
}

type RTMEvent struct {
	SendUserID string `json:"sendUserID"`
	MsgID      uint16 `json:"msgID"`
	Msg        []byte `json:"msg"`
}
