package routingtable

import (
	"sync"
)

type RoutingTable struct {
	Gateway string
	GenMask string
	Metric  int
	Export  int
}

type Table struct {
	Routing  sync.Map // 路由信息和其对应的端口
	LifeTime sync.Map // 控制路由表中信息的失效时间
}

// 直接配置默认的路由表 也是直连的路由表
func New() *Table {
	table := new(Table)
	table.Routing.Store("127.0.0.1", &RoutingTable{
		Gateway: "127.0.0.1",
		GenMask: "255.255.255.0",
		Metric:  0,
		Export:  0,
	})
	table.Routing.Store("172.25.204.193", &RoutingTable{
		Gateway: "172.25.204.193",
		GenMask: "255.255.255.0",
		Metric:  0,
		Export:  1,
	})
	return table
}

func (t *Table) QueryTableByDesIP(DesIP string) (*RoutingTable, bool) {
	if v, ok := t.Routing.Load(DesIP); ok {
		return v.(*RoutingTable), ok
	}
	return nil, false// 先不管找不到的情况
}


