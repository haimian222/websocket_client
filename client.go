package websocker_client

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

// NewClient 新建WebSocket客户端
func NewClient(url string) *Client {
	return &Client{
		url:           url,
		maxRetry:      3,
		retryInterval: 3,
		messageChan:   make(chan Message, 1024),
		eventChan:     make(chan Event, 1024),
		isConnect:     false,
		isDisconnect:  false,
	}
}

// Client WebSocket客户端
type Client struct {
	url           string          // WebSocket服务器地址
	conn          *websocket.Conn // WebSocket连接
	maxRetry      int             // 最大重试次数
	retryInterval int             // 重试间隔 单位: 秒
	messageChan   chan Message    // 消息通道
	eventChan     chan Event      // 事件通道
	isConnect     bool            // 是否连接
	isDisconnect  bool            // 是否断开
}

// GetMessageChan 获取消息通道
func (c *Client) GetMessageChan() chan Message {
	return c.messageChan
}

// GetEventChan 获取事件通道
func (c *Client) GetEventChan() chan Event {
	return c.eventChan
}

// SetMaxRetry 设置最大重试次数 -1为无限重试
func (c *Client) SetMaxRetry(maxRetry int) {
	c.maxRetry = maxRetry
}

// SetRetryInterval 设置重试间隔 单位: 秒
func (c *Client) SetRetryInterval(retryInterval int) {
	c.retryInterval = retryInterval
}

// GetUrl 获取WebSocket服务器地址
func (c *Client) GetUrl() string {
	return c.url
}

// GetConnectStatus 获取连接状态
func (c *Client) GetConnectStatus() bool {
	return c.isConnect
}

// Connect 连接WebSocket服务器
func (c *Client) Connect() {
	if c.isConnect {
		log.Printf("error: %v\n", "client is connected")
		return
	}
	c.isDisconnect = false
	retryCount := 0
	for !c.isDisconnect {
		var err error
		c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			if c.maxRetry != -1 {
				// 重试次数限制
				retryCount++
				if retryCount > c.maxRetry {
					break
				}
			}
			time.Sleep(time.Duration(c.retryInterval) * time.Second)
			continue
		}

		// 连接成功
		retryCount = 0
		c.isConnect = true
		if c.eventChan != nil {
			c.eventChan <- Event{
				Type: "connect_success",
			}
		} else {
			log.Printf("error: %v\n", "eventChan is nil")
		}

		// 接收消息
		for !c.isDisconnect {
			messageType, messageData, err := c.conn.ReadMessage()
			if err != nil {
				break
			}
			if c.messageChan != nil {
				c.messageChan <- Message{
					Type: messageType,
					Data: messageData,
				}
			} else {
				log.Printf("error: %v\n", "messageChan is nil")
			}
		}
		c.isConnect = false
		if c.isDisconnect {
			// 主动断开
			break
		}
		time.Sleep(time.Duration(c.retryInterval) * time.Second)
	}

	if !c.isDisconnect {
		// 断开连接
		if err := c.conn.Close(); err != nil {
			log.Printf("error: %v\n", err)
		}
		// 连接失败
		if c.eventChan != nil {
			c.eventChan <- Event{
				Type: "connect_fail",
			}
		} else {
			log.Printf("error: %v\n", "eventChan is nil")
		}
	} else {
		// 主动断开
		if c.eventChan != nil {
			c.eventChan <- Event{
				Type: "disconnect",
			}
		} else {
			log.Printf("error: %v\n", "eventChan is nil")
		}
	}
}

// Disconnect 断开
func (c *Client) Disconnect() (err error) {
	c.isDisconnect = true
	return c.conn.Close()
}
