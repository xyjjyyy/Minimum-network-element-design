package pdu

type FrameHeap []*Frame

func (h FrameHeap) Len() int           { return len(h) }
func (h FrameHeap) Less(i, j int) bool { return h[i].SerialNum < h[j].SerialNum }
func (h FrameHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *FrameHeap) Push(x interface{}) {
	*h = append(*h, x.(*Frame))
}

func (h *FrameHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
