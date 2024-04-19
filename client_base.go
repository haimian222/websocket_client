package websocker_client

import (
	"github.com/gorilla/websocket"
	"time"
)

// NewClientBase 新建WebSocket客户端基类
func NewClientBase(url string, clientID int, messageChan chan Message, eventChan chan Event) *ClientBase {
	return &ClientBase{
		url:           url,
		clientID:      clientID,
		maxRetry:      3,
		retryInterval: 3,
		messageChan:   messageChan,
		eventChan:     eventChan,
	}
}

// ClientBase WebSocket客户端基类
type ClientBase struct {
	clientID      int             // 客户端ID
	url           string          // WebSocket服务器地址
	conn          *websocket.Conn // WebSocket连接
	maxRetry      int             // 最大重试次数
	retryInterval int             // 重试间隔 单位: 秒
	messageChan   chan Message    // 消息通道
	eventChan     chan Event      // 事件通道
	isConnect     bool            // 是否连接
}

// SetMaxRetry 设置最大重试次数
func (c *ClientBase) SetMaxRetry(maxRetry int) {
	c.maxRetry = maxRetry
}

// SetRetryInterval 设置重试间隔 单位: 秒
func (c *ClientBase) SetRetryInterval(retryInterval int) {
	c.retryInterval = retryInterval
}

// GetClientID 获取客户端ID
func (c *ClientBase) GetClientID() int {
	return c.clientID
}

// GetUrl 获取WebSocket服务器地址
func (c *ClientBase) GetUrl() string {
	return c.url
}

// Connect 连接WebSocket服务器
func (c *ClientBase) Connect() {
	c.isConnect = true
	retryCount := 0
	for c.isConnect {
		var err error
		c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			c.eventChan <- Event{
				ClientID: c.clientID,
				Type:     "connect_fail",
			}
			retryCount++
			if retryCount > c.maxRetry {
				break
			}
			time.Sleep(time.Duration(c.retryInterval) * time.Second)
			continue
		}

		// 连接成功
		retryCount = 0
		c.eventChan <- Event{
			ClientID: c.clientID,
			Type:     "connect_success",
		}

		// 接收消息
		for c.isConnect {
			messageType, messageData, err := c.conn.ReadMessage()
			if err != nil {
				c.eventChan <- Event{
					ClientID: c.clientID,
					Type:     "disconnect",
				}
				break
			}
			c.messageChan <- Message{
				ClientID: c.clientID,
				Type:     messageType,
				Data:     messageData,
			}
		}
	}
	if !c.isConnect {
		// 重试次数用尽
		c.isConnect = false
		c.eventChan <- Event{
			ClientID: c.clientID,
			Type:     "max_retry",
		}
	}
}

// Disconnect 断开
func (c *ClientBase) Disconnect() (err error) {
	c.isConnect = false
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

// Close 关闭连接
func (c *ClientBase) Close() (err error) {
	c.isConnect = false
	if err = c.conn.Close(); err != nil {
		return err
	}
	return nil
}
