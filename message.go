package websocker_client

type Message struct {
	ClientID int    // 客户端ID
	Type     int    // 消息类型
	Data     []byte // 消息数据
}
