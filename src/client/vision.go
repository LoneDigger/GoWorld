package client

type AreaSet map[int]bool

// 玩家註冊區塊
type Vision struct {
	main int
	m    AreaSet
}

func NewVision() *Vision {
	return &Vision{
		main: -1,
		m:    make(map[int]bool),
	}
}

func (h *Vision) Add(id int) {
	h.m[id] = false
}

func (h *Vision) Remove(id int) {
	delete(h.m, id)
}

func (h *Vision) SetMain(id int) {
	h.main = id
}

func (h *Vision) Main() int {
	return h.main
}

func (h *Vision) Areas() AreaSet {
	return h.m
}
