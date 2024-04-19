package websocker_client

import (
	"errors"
	"sync"
)

func NewClientManage() *ClientManage {
	return &ClientManage{
		clientMap: make(map[int]*ClientBase),
		//cursorID:    0,
		messageChan: make(chan Message, 10240),
		eventChan:   make(chan Event, 10240),
	}
}

// ClientManage 客户端管理
type ClientManage struct {
	clientMap map[int]*ClientBase // 客户端映射
	//cursorID    int                 // 客户端ID游标
	addClientLock sync.Mutex   // 添加客户端锁
	messageChan   chan Message // 消息通道
	eventChan     chan Event   // 事件通道
}

// GetMessageChan 获取消息通道
func (cm *ClientManage) GetMessageChan() chan Message {
	return cm.messageChan
}

// GetEventChan 获取事件通道
func (cm *ClientManage) GetEventChan() chan Event {
	return cm.eventChan
}

// 获取一个不与已有ID重复的客户端ID
func (cm *ClientManage) getNewClientID() (clientID int) {
	for {
		clientID++
		if _, ok := cm.clientMap[clientID]; !ok {
			return clientID
		}
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
	cm.addClientLock.Lock()
	defer cm.addClientLock.Unlock()
	//判断是否已经存在
	if cm.IsExistClientByUrl(url) {
		return 0, errors.New("client is exist")
	}

	clientID = cm.getNewClientID()
	client := NewClientBase(url, clientID, cm.messageChan, cm.eventChan)
	cm.clientMap[clientID] = client
	go client.Connect()
	return clientID, nil
}

// DelClient 删除客户端
func (cm *ClientManage) DelClient(clientID int) (err error) {
	//判断是否存在
	if !cm.IsExistClient(clientID) {
		return errors.New("client is not exist")
	}
	if err := cm.clientMap[clientID].Close(); err != nil {
		return err
	}
	delete(cm.clientMap, clientID)
	return nil
}

// DisconnectClient 断开客户端
func (cm *ClientManage) DisconnectClient(clientID int) (err error) {
	//判断是否存在
	if !cm.IsExistClient(clientID) {
		return errors.New("client is not exist")
	}
	if err := cm.clientMap[clientID].Disconnect(); err != nil {
		return err
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
		return false, errors.New("client is not exist")
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

// Close 关闭所有客户端
func (cm *ClientManage) Close() {
	for _, client := range cm.clientMap {
		if err := client.Close(); err != nil {
			continue
		}
	}
	close(cm.messageChan)
	close(cm.eventChan)
}
