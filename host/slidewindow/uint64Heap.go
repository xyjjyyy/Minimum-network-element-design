package slidewindow

type Uint64Heap []uint64

func (h Uint64Heap) Len() int           { return len(h) }
func (h Uint64Heap) Less(i, j int) bool { return h[i] < h[j] }
func (h Uint64Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *Uint64Heap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *Uint64Heap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
