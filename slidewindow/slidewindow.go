package slidewindow

import (
	"errors"
	"fmt"
	"log"
	"network/pdu"
	"sync"
	"time"
)

const (
	windowLen = 10
)

var (
	ErrEmpty     = errors.New("slideWindow has nothing")
	ErrWatingAck = errors.New("slideWindow is waiting ack")
	ErrFrame     = errors.New("frame is nil or ackWindow is empty")
)

type Window struct {
	FrameWaitAck    []*pdu.Frame
	FrameNeedSend   []*pdu.Frame
	FrameStored     []*pdu.Frame
	FrameNeedReSend []*pdu.Frame

	WindowLen int

	mu sync.Mutex
}

func New() *Window {
	w := &Window{
		FrameWaitAck:    []*pdu.Frame{},
		FrameNeedSend:   []*pdu.Frame{},
		FrameStored:     []*pdu.Frame{},
		FrameNeedReSend: []*pdu.Frame{},
		WindowLen:       windowLen,
	}
	go w.adjust()
	return w
}

// a goroutine to adjust the window
func (w *Window) adjust() {
	for {
		w.mu.Lock()
		for len(w.FrameNeedSend)+len(w.FrameWaitAck) < w.WindowLen &&
			len(w.FrameStored) > 0 {
			w.FrameNeedSend = append(w.FrameNeedSend, w.FrameStored[0])
			w.FrameStored = w.FrameStored[1:]
		}
		w.mu.Unlock()
		time.Sleep(200 * time.Millisecond)
	}
}

func (w *Window) AddFrame(frameList ...*pdu.Frame) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, frame := range frameList {
		if frame == nil {
			panic("AddFrame must not be nil")
		}
		w.FrameStored = append(w.FrameStored, frame)
	}
}

func (w *Window) SlideWindowForSend() (*pdu.Frame, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 先取出待重发的
	if len(w.FrameNeedReSend) > 0 {
		frame := w.FrameNeedReSend[0]
		w.FrameNeedReSend = w.FrameNeedReSend[1:]
		log.Printf("SerialNum[%d] has resend\n", frame.SerialNum)
		return frame, nil
	}

	if len(w.FrameNeedSend) > 0 {
		frame := w.FrameNeedSend[0]
		w.FrameNeedSend = w.FrameNeedSend[1:]
		w.FrameWaitAck = append(w.FrameWaitAck, frame)
		return frame, nil
	}

	if len(w.FrameStored) > 0 {
		return nil, ErrWatingAck
	}

	return nil, ErrEmpty
}

// 接受道 `ackFrame` 需要滑动窗口
func (w *Window) SlideWindowForAck(frameAck *pdu.Frame) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if frameAck == nil || len(w.FrameWaitAck) == 0 {
		return ErrFrame
	}

	index, err := binarySearch(w.FrameWaitAck, frameAck.SerialNum)
	if err != nil {
		return err
	}

	// 取消对应 `frame` 的时钟
	w.FrameWaitAck[index].Done <- struct{}{}

	w.FrameWaitAck = append(w.FrameWaitAck[:index], w.FrameWaitAck[index+1:]...)
	return nil
}

func (w *Window) StartClock(frame *pdu.Frame) string {
	select {
	case <-time.After(time.Millisecond * time.Duration(frame.Duration)):
		w.mu.Lock()
		w.FrameNeedReSend = append(w.FrameNeedReSend, frame)
		w.mu.Unlock()
		return fmt.Sprintf("serialNum:%d data:%s out of time\n", frame.SerialNum, frame.Data)
	case <-frame.Done:
		return fmt.Sprintf("serialNum:%d data:%s send successfully\n", frame.SerialNum, frame.Data)
	}
}

func binarySearch(arr []*pdu.Frame, num uint64) (int, error) {
	l := 0
	r := len(arr)

	for l < r {
		m := (l + r) / 2
		if arr[m].SerialNum == num {
			return m, nil
		} else if arr[m].SerialNum > num {
			r = m
		} else {
			l = m + 1
		}
	}
	return -1, fmt.Errorf("frame[%d] not find", num)
}
