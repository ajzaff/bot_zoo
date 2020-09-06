package zoo

import "container/list"

type Bound int

const (
	NoBound Bound = iota
	UpperBound
	LowerBound
	ExactBound
)

// EntrySize is the approximate TableEntry size in bytes.
// This can be calculated using the unsafe package:
//	unsafe.Sizeof(TableEntry{})
const EntrySize = 40

type TableEntry struct {
	Bound
	ZHash int64
	Depth int
	Value int
	Step  *Step
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
	if elem, ok := t.table[e.ZHash]; ok {
		elem.Value = e
		t.list.MoveToBack(elem)
		return
	}
	elem := t.list.PushBack(e)
	t.table[e.ZHash] = elem
}

// Best returns the best move by probing the table.
// This is similar to PV but only returns steps for the current side.
func (t *Table) Best(p *Pos) (move []Step, score int, err error) {
	initSide := p.Side()
	for i := 0; initSide == p.Side(); i++ {
		e, ok := t.ProbeDepth(p.zhash, 0)
		if !ok || e.Bound != ExactBound || e.Step == nil {
			break
		}
		if i == 0 {
			score = e.Value
		}
		if err := p.Step(*e.Step); err != nil {
			return nil, 0, err
		}
		move = append(move, *e.Step)
	}
	for range move {
		if err := p.Unstep(); err != nil {
			return nil, 0, err
		}
	}
	return move, score, nil
}

// PV returns the principal variation by probing the table.
// The PV has a maximum length of 50 steps.
func (t *Table) PV(p *Pos) (pv []Step, score int, err error) {
	for i := 0; i < 50; i++ {
		e, ok := t.ProbeDepth(p.zhash, 0)
		if !ok || e.Bound != ExactBound || e.Step == nil {
			break
		}
		if i == 0 {
			score = e.Value
		}
		if err := p.Step(*e.Step); err != nil {
			return nil, 0, err
		}
		pv = append(pv, *e.Step)
	}
	for range pv {
		if err := p.Unstep(); err != nil {
			return nil, 0, err
		}
	}
	return pv, score, nil
}
