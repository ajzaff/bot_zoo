package zoo

import "container/list"

type Bound int

const (
	NoBound Bound = iota
	UpperBound
	LowerBound
	ExactBound
)

type TableEntry struct {
	Bound
	ZHash int64
	Depth int
	Value int
	Move  []Step
	prev  *TableEntry
	next  *TableEntry
}

type Table struct {
	cap   int
	list  list.List
	table map[int64]*list.Element
}

func NewTable(cap int) *Table {
	return &Table{
		cap:   cap,
		table: make(map[int64]*list.Element),
		list:  list.List{},
	}
}

func (t *Table) Clear() {
	t.table = make(map[int64]*list.Element)
	t.list = list.List{}
}

func (t *Table) ProbeDepth(key int64, depth int) (e *TableEntry, ok bool) {
	if e, ok := t.table[key]; ok {
		t.list.MoveToBack(e)
		// TODO(ajzaff): Currently the table is cleared after each search so entry.Depth > depth is not possible.
		if entry := e.Value.(*TableEntry); entry.Depth >= depth {
			return entry, true
		}
	}
	return nil, false
}

func (t *Table) Len() int {
	return t.list.Len()
}

func (t *Table) Cap() int {
	return t.cap
}

func (t *Table) SetCap(cap int) {
	t.cap = cap
}

func (t *Table) remove(e *list.Element) {
	t.list.Remove(e)
	delete(t.table, e.Value.(*TableEntry).ZHash)
}

func (t *Table) pop() {
	e := t.list.Remove(t.list.Front())
	delete(t.table, e.(*TableEntry).ZHash)
}

func (t *Table) Store(e *TableEntry) {
	for t.Len() >= t.Cap() {
		t.pop()
	}
	elem := t.list.PushBack(e)
	t.table[e.ZHash] = elem
}
