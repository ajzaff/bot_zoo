package zoo

import (
	"container/heap"
)

type Bound int

const (
	NoBound Bound = iota
	UpperBound
	LowerBound
	ExactBound
)

type Entry struct {
	Bound
	ZHash int64
	Epoch uint64
	Depth int
	Value int
	Move  []Step
}

type Table struct {
	table map[int64]int
	data  []*Entry
	moves [][]Step
	epoch uint64
	cap   int
}

func NewTable(cap int) *Table {
	return &Table{
		table: make(map[int64]int),
		cap:   cap,
	}
}

func (t *Table) Clear() {
	t.table = make(map[int64]int)
	t.data = make([]*Entry, 0)
	t.moves = make([][]Step, 0)
	t.epoch = 0
}

func (t *Table) ProbeDepth(key int64, depth int) (e *Entry, ok bool) {
	if i, ok := t.table[key]; ok {
		e = t.data[i]
		e.Epoch = t.epoch
		t.epoch++
		heap.Fix(t, i)
		if e.Depth >= depth {
			return e, true
		}
	}
	return nil, false
}

func (t *Table) Len() int {
	return len(t.data)
}

func (t *Table) Swap(i, j int) {
	t.table[t.data[i].ZHash] = j
	t.table[t.data[j].ZHash] = i
	t.data[i], t.data[j] = t.data[j], t.data[i]
	t.moves[i], t.moves[j] = t.moves[j], t.moves[i]
}

func (t *Table) Less(i, j int) bool {
	return t.data[i].Epoch < t.data[j].Epoch
}

func (t *Table) Push(x interface{}) {
	e := x.(*Entry)
	t.table[e.ZHash] = len(t.data)
	t.data = append(t.data, e)
	t.moves = append(t.moves, e.Move)
}

func (t *Table) Pop() interface{} {
	e := t.data[t.Len()-1]
	delete(t.table, e.ZHash)
	t.data = t.data[:t.Len()-1]
	t.moves = t.moves[:t.Len()-1]
	return e
}

func (t *Table) Store(e *Entry) {
	for t.cap <= t.Len() {
		heap.Pop(t)
	}
	e.Epoch = t.epoch
	t.epoch++
	heap.Push(t, e)
}
