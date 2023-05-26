package slidewindow

import (
	"container/heap"
	"container/list"
	"errors"
	"fmt"
	"log"
	"network/host/pdu"
	"time"
)

// 初始化一个滑动窗口 Window
func New() *Window {
	return &Window{
		FrameWaitAck:    list.New(),
		FrameNeedSend:   list.New(),
		FrameStored:     list.New(),
		FrameNeedReSend: list.New(),
		WindowLen:       windowLen,
		recvAckPool:     &Uint64Heap{},
	}
}

func (w *Window) adjust() {
	for w.FrameNeedSend.Len()+w.FrameWaitAck.Len() < w.WindowLen {

		if w.FrameStored.Len() == 0 {
			break
		}
		frameNeedSend := w.FrameStored.Remove(w.FrameStored.Front()).(*pdu.Frame)
		w.FrameNeedSend.PushBack(frameNeedSend)
	}
}

func (w *Window) AddFrame(frameList []*pdu.Frame) {
	if frameList == nil {
		return
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, frame := range frameList {
		w.FrameStored.PushBack(frame)
	}

	// 适应一下窗口
	w.adjust()
}

func (w *Window) SlideWindowForSend() (*pdu.Frame, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.FrameNeedReSend.Len() != 0 {
		frame := w.FrameNeedReSend.Remove(w.FrameNeedReSend.Front()).(*pdu.Frame)
		fmt.Printf("SerialNum: %d 已经被重发\n", frame.SerialNum)
		return frame, nil
	}
	w.adjust()

	if w.FrameNeedSend.Len() != 0 {
		frame := w.FrameNeedSend.Remove(w.FrameNeedSend.Front()).(*pdu.Frame)

		w.FrameWaitAck.PushBack(frame)
		return frame, nil
	} else if w.FrameStored.Len() != 0 {
		return nil, ErrWaiting
	}

	return nil, nil
}

func (w *Window) SlideWindowForAck(frameAck *pdu.Frame) error {
	if frameAck == nil {
		return errors.New("frameAck is nil")
	}

	if w.FrameWaitAck.Len() == 0 {
		return errors.New("frameWaitAck is empty")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	front := w.FrameWaitAck.Front().Value.(*pdu.Frame)
	back := w.FrameWaitAck.Back().Value.(*pdu.Frame)

	heap.Push(w.recvAckPool, frameAck.SerialNum)
	// fmt.Println((*(w.recvAckPool))[0], " && ", front.SerialNum)

	// 处理掉无用的序号 重复的ack发过来 我们要进行处理掉
	for (*(w.recvAckPool))[0] < front.SerialNum || (*(w.recvAckPool))[0] > back.SerialNum {
		fmt.Printf("Num:%d已经被丢弃", (*(w.recvAckPool))[0])
		heap.Pop(w.recvAckPool)
		if w.recvAckPool.Len() == 0 {
			return errors.New("ACK序号不符合要求")
		}
	}

	//TODO:这里为了处理时钟问题 但是感觉可以和快速重传结合起来
	//todo 并不一定需要最前面接受了才进行移动
	for iter := w.FrameWaitAck.Front(); iter != nil; iter = iter.Next() {
		frame := iter.Value.(*pdu.Frame)
		if frame.SerialNum == frameAck.SerialNum {
			frame.Done <- struct{}{}
			break
		}
	}
	// 判断栈顶元素是否相等
	for (*(w.recvAckPool))[0] == front.SerialNum {
		fmt.Println((*(w.recvAckPool))[0], " && ", front.SerialNum)
		log.Printf("serialNum:%d 成功发送", front.SerialNum)
		// * 对 front 进行更新
		w.FrameWaitAck.Remove(w.FrameWaitAck.Front())
		w.adjust()

		// 如果下一个序号和top相等 说明序号发多了 我们要去掉
		top := heap.Pop(w.recvAckPool).(uint64)
		for w.recvAckPool.Len() != 0 {
			if (*(w.recvAckPool))[0] != top {
				break
			}
			heap.Pop(w.recvAckPool)
		}
		// 此时所有ack序号都已经被接受
		if w.recvAckPool.Len() == 0 {
			break
		}
		front = w.FrameWaitAck.Front().Value.(*pdu.Frame)
	}

	return nil
}

func (w *Window) SlideWindowForErr(frameErr *pdu.Frame) error {
	if frameErr == nil {
		return errors.New("frameErr is nil")
	}

	if w.FrameWaitAck.Len() == 0 {
		return errors.New("frameErr is empty")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	for iter := w.FrameWaitAck.Front(); iter != nil; iter = iter.Next() {
		frame := iter.Value.(*pdu.Frame)
		if frame.SerialNum == frameErr.SerialNum {
			frame.Done <- struct{}{}
			w.FrameNeedReSend.PushBack(frame)
			return nil
		}
	}
	return errors.New("ERR序号不符合要求")
}

func (w *Window) StartClock(frame *pdu.Frame) {
	fmt.Printf("serialNum:%d start clock\n", frame.SerialNum)
	select {
	case <-time.After(time.Millisecond * time.Duration(frame.Duration)):
		w.mutex.Lock()
		w.FrameNeedReSend.PushFront(frame)
		w.mutex.Unlock()
		fmt.Printf("serialNum:%d data:%s out of time\n", frame.SerialNum, frame.Data)
	case <-frame.Done:
		fmt.Printf("serialNum:%d data:%s send successfully\n", frame.SerialNum, frame.Data)
	}
}
