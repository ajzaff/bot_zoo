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
	index int
}

type Table struct {
	table map[int64]*Entry
	data  []*Entry
	moves [][]Step
	epoch uint64
	cap   int
}

func NewTable(cap int) *Table {
	return &Table{
		table: make(map[int64]*Entry),
		cap:   cap,
	}
}

func (t *Table) Clear() {
	t.table = make(map[int64]*Entry)
	t.data = make([]*Entry, 0)
	t.moves = make([][]Step, 0)
	t.epoch = 0
}

func (t *Table) ProbeDepth(key int64, depth int) (e *Entry, ok bool) {
	if e, ok := t.table[key]; ok {
		e.Epoch = t.epoch
		t.epoch++
		heap.Fix(t, e.index)
		// TODO(ajzaff): Change this to `e.Depth >= depth`
		// once quiescence search is added.
		if e.Depth == depth {
			return e, true
		}
	}
	return nil, false
}

func (t *Table) Len() int {
	return len(t.data)
}

func (t *Table) Swap(i, j int) {
	t.table[t.data[i].ZHash], t.table[t.data[j].ZHash] = t.data[j], t.data[i]
	t.data[i], t.data[j] = t.data[j], t.data[i]
	t.data[i].index, t.data[j].index = j, i
	t.moves[i], t.moves[j] = t.moves[j], t.moves[i]
}

func (t *Table) Less(i, j int) bool {
	return t.data[i].Epoch < t.data[j].Epoch
}

func (t *Table) Push(x interface{}) {
	e := x.(*Entry)
	oldEntry, ok := t.table[e.ZHash]
	if ok && oldEntry.Depth >= e.Depth {
		return
	}
	e.Epoch = t.epoch
	t.epoch++
	t.table[e.ZHash] = e
	if t.Len() >= t.cap {
		for t.Len() > t.cap {
			heap.Pop(t)
		}
		e.index = 0
		t.data[0] = e
		t.moves[0] = e.Move
		heap.Fix(t, 0)
		return
	}
	e.index = t.Len()
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
	heap.Push(t, e)
}
