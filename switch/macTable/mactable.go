package macTable

import (
	"log"
	"sync"
	"time"
)

type MacTable struct {
	MacAddress sync.Map //map[string]int
	LifeTime   sync.Map
}

// 初始化就是简单的初始化 不带任何信息的同步 Map
func New() *MacTable {
	return &MacTable{}
}

// 查询目的 MACAddress 是否存在
func (m *MacTable) QueryTableByMacAddress(macAddress string) (int, bool) {
	if v, ok := m.MacAddress.Load(macAddress); ok {
		return v.(int), true
	}
	return 0, false
}

func (m *MacTable) StudyFromSource(macAddress string, port int) {
	clock := make(chan struct{}, 1)
	if v, ok := m.MacAddress.Load(macAddress); ok {
		if port == v.(int) {
			return
		}
		clock <- struct{}{}
		m.LifeTime.Store(macAddress, clock)
	}
	m.MacAddress.Store(macAddress, port)

	m.LifeTime.Store(macAddress, clock)
	// 启动时钟
	go m.startClock(macAddress)
}

// 既然执行这个程序 意味着macAddress一定存在
func (m *MacTable) startClock(macAddress string) {
	t, _ := m.LifeTime.Load(macAddress)
	select {
	case <-time.After(3 * time.Second):
		m.MacAddress.Delete(macAddress)
		log.Printf("macAddress %s 已经过期被删除", macAddress)
	case <-t.(chan struct{}):
		log.Printf("macAddress %s 已经更新对应端口", macAddress)
	}
}
