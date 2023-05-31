package mactable

import (
	"log"
	"sync"
	"time"
)

const (
	// mac表中的信息的生命周期
	lifeTime = 10 * time.Second
)

type MacTable struct {
	MacAddress sync.Map // mac地址和其对应的端口
	LifeTime   sync.Map // 控制mac表中信息的失效时间
}

// 初始化就是简单的初始化 不带任何信息的同步 Map
func New() *MacTable {
	return &MacTable{}
}

// 查询目的 MACAddress 是否存在
// 不存在的话就返回-1 会进行广播
func (m *MacTable) GetPort(macAddress string) (int, bool) {
	if v, ok := m.MacAddress.Load(macAddress); ok {
		return v.(int), true
	}
	return -1, false
}

func (m *MacTable) StudyFromSource(macAddress string, port int) {
	m.MacAddress.Store(macAddress, port)
	clock, loaded := m.LifeTime.LoadOrStore(macAddress, time.NewTimer(lifeTime))
	if !loaded {
		// 启动时钟
		go m.startClock(macAddress)
	} else {
		timer := clock.(*time.Timer)
		timer.Reset(lifeTime)
	}

}

// 启动时钟
func (m *MacTable) startClock(macAddress string) {
	t, ok := m.LifeTime.Load(macAddress)
	if ok {
		<-t.(*time.Timer).C
		m.MacAddress.Delete(macAddress)
		m.LifeTime.Delete(macAddress)
		log.Println("mac address:", macAddress, "time out")
	}
}
