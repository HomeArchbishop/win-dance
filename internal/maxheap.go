package internal

// RectHeap is a max-heap of *RectInfo based on UnusedSqr
type RectHeap []*RectInfo

func (h RectHeap) Len() int           { return len(h) }
func (h RectHeap) Less(i, j int) bool { return h[i].UnusedSqr > h[j].UnusedSqr }
func (h RectHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *RectHeap) Push(x interface{}) {
	*h = append(*h, x.(*RectInfo))
}

func (h *RectHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
