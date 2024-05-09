package websocker_client

import (
	"errors"
	"log"
	"sync"
)

func NewClientManage() *ClientManage {
	return &ClientManage{
		clientMap:   make(map[int]*ClientBase),
		messageChan: make(chan Message, 10240),
		eventChan:   make(chan Event, 10240),
	}
}

// ClientManage 客户端管理
type ClientManage struct {
	clientMap   map[int]*ClientBase // 客户端映射
	clientLock  sync.Mutex          // 客户端锁
	messageChan chan Message        // 消息通道
	eventChan   chan Event          // 事件通道
}

// GetMessageChan 获取消息通道
func (cm *ClientManage) GetMessageChan() chan Message {
	return cm.messageChan
}

// GetEventChan 获取事件通道
func (cm *ClientManage) GetEventChan() chan Event {
	return cm.eventChan
}

// 获取一个不与已有ID重复的客户端ID 从0开始
func (cm *ClientManage) getNewClientID() (clientID int) {
	clientID = 0
	for {
		if _, ok := cm.clientMap[clientID]; !ok {
			return clientID
		}
		clientID++
	}
}

// IsExistClient 判断客户端是否存在
func (cm *ClientManage) IsExistClient(clientID int) bool {
	_, ok := cm.clientMap[clientID]
	return ok
}

// IsExistClientByUrl Url判断客户端是否存在
func (cm *ClientManage) IsExistClientByUrl(url string) bool {
	for _, client := range cm.clientMap {
		if client.GetUrl() == url {
			return true
		}
	}
	return false
}

// AddClient 添加客户端
func (cm *ClientManage) AddClient(url string) (clientID int, err error) {
	cm.clientLock.Lock()
	defer cm.clientLock.Unlock()
	//判断是否已经存在
	if cm.IsExistClientByUrl(url) {
		return -1, errors.New("error: client is exist")
	}
	clientID = cm.getNewClientID()
	client := NewClientBase(url, clientID)
	client.SetMessageChan(cm.messageChan)
	client.SetEventChan(cm.eventChan)
	cm.clientMap[clientID] = client
	return clientID, nil
}

// ConnectClient 连接客户端
func (cm *ClientManage) ConnectClient(clientID int) (err error) {
	//判断是否存在
	if !cm.IsExistClient(clientID) {
		return errors.New("error: client is not exist")
	}
	//判断是否连接
	if cm.clientMap[clientID].isConnect {
		return errors.New("error: client is connected")
	}
	go cm.clientMap[clientID].Connect()
	return nil
}

// DelClient 删除客户端
func (cm *ClientManage) DelClient(clientID int) (err error) {
	//判断是否存在
	if !cm.IsExistClient(clientID) {
		return errors.New("error: client is not exist")
	}

	//判断是否连接
	if cm.clientMap[clientID].isConnect {
		//断开连接
		if err := cm.clientMap[clientID].Disconnect(); err != nil {
			return err
		}
	}

	delete(cm.clientMap, clientID)
	return nil
}

// DisconnectClient 断开客户端
func (cm *ClientManage) DisconnectClient(clientID int) (err error) {
	//判断是否存在
	if !cm.IsExistClient(clientID) {
		return errors.New("error: client is not exist")
	}
	//判断是否连接
	if !cm.clientMap[clientID].isConnect {
		return errors.New("error: client is not connect")
	}

	if err := cm.clientMap[clientID].Disconnect(); err != nil {
		return err
	}
	return nil
}

// SetMaxRetry 设置最大重试次数 clientID设置-1表示全部 maxRetry设置-1表示不限制
func (cm *ClientManage) SetMaxRetry(clientID int, maxRetry int) (err error) {
	if clientID == -1 {
		for _, client := range cm.clientMap {
			client.SetMaxRetry(maxRetry)
		}
	} else {
		if !cm.IsExistClient(clientID) {
			return errors.New("error: client is not exist")
		}
		cm.clientMap[clientID].SetMaxRetry(maxRetry)
	}
	return nil
}

// SetRetryInterval 设置重试间隔 单位: 秒 clientID设置-1表示全部
func (cm *ClientManage) SetRetryInterval(clientID int, retryInterval int) (err error) {
	if clientID == -1 {
		for _, client := range cm.clientMap {
			client.SetRetryInterval(retryInterval)
		}
	} else {
		if !cm.IsExistClient(clientID) {
			return errors.New("error: client is not exist")
		}
		cm.clientMap[clientID].SetRetryInterval(retryInterval)
	}
	return nil
}

// GetClientIDByUrl 通过URL获取客户端ID
func (cm *ClientManage) GetClientIDByUrl(url string) (clientID int, err error) {
	for _, client := range cm.clientMap {
		if client.GetUrl() == url {
			return client.GetClientID(), nil
		}
	}
	return 0, errors.New("client is not exist")
}

// GetClientUrlByID 通过客户端ID获取客户端Url
func (cm *ClientManage) GetClientUrlByID(clientID int) (url string, err error) {
	if !cm.IsExistClient(clientID) {
		return "", errors.New("client is not exist")
	}
	return cm.clientMap[clientID].GetUrl(), nil
}

// GetAllClientList 获取所有客户端列表
func (cm *ClientManage) GetAllClientList() (clientList []int) {
	for clientID := range cm.clientMap {
		clientList = append(clientList, clientID)
	}
	return clientList

}

// GetClientCount 获取客户端数量
func (cm *ClientManage) GetClientCount() int {
	return len(cm.clientMap)
}

// GetOnlineClientCount 获取在线客户端数量
func (cm *ClientManage) GetOnlineClientCount() int {
	count := 0
	for _, client := range cm.clientMap {
		if client.isConnect {
			count++
		}
	}
	return count
}

// GetDisconnectClientCount 获取断开的客户端数量
func (cm *ClientManage) GetDisconnectClientCount() int {
	count := 0
	for _, client := range cm.clientMap {
		if !client.isConnect {
			count++
		}
	}
	return count
}

// GetClientConnectStatus 获取客户端连接状态
func (cm *ClientManage) GetClientConnectStatus(clientID int) (isConnect bool, err error) {
	if !cm.IsExistClient(clientID) {
		return false, errors.New("error: client is not exist")
	}
	return cm.clientMap[clientID].isConnect, nil
}

// GetDisconnectClientList 获取断开客户端列表
func (cm *ClientManage) GetDisconnectClientList() (clientIDs []int) {
	for clientID, client := range cm.clientMap {
		if !client.isConnect {
			clientIDs = append(clientIDs, clientID)
		}
	}
	return clientIDs
}

// GetOnlineClientList 获取在线客户端列表
func (cm *ClientManage) GetOnlineClientList() (clientIDs []int) {
	for clientID, client := range cm.clientMap {
		if client.isConnect {
			clientIDs = append(clientIDs, clientID)
		}
	}
	return clientIDs
}

// Close 关闭
func (cm *ClientManage) Close() {
	for _, client := range cm.clientMap {
		if client.isConnect {
			if err := client.Disconnect(); err != nil {
				log.Printf("error: %v\n", err)
				continue
			}
		}
		delete(cm.clientMap, client.GetClientID())
	}
	close(cm.messageChan)
	close(cm.eventChan)
}
