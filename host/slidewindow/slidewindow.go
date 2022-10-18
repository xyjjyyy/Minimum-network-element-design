package slidewindow

import (
	"container/list"
	"errors"
	"sync"
)

const (
	windowLen = 4
)

var ErrWaiting error = errors.New("wating for ack")

type Window struct {
	FrameWaitAck    *list.List
	FrameNeedSend   *list.List
	FrameStored     *list.List
	FrameNeedReSend *list.List

	mutex      sync.Mutex
	WindowLen  int

	//优先队列 最小堆
	recvAckPool *Uint64Heap
}
