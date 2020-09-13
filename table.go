package zoo

import (
	"container/list"
	"fmt"
	"sync"
)

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
	Depth int16
	Value Value
	Step  *Step
}

type Table struct {
	cap   int
	list  *list.List              // guarded by m
	table map[int64]*list.Element // zhash => *list.Element
	m     sync.Mutex
}

func NewTable(cap int) *Table {
	return &Table{
		cap:   cap,
		list:  list.New(),
		table: make(map[int64]*list.Element),
	}
}

func (t *Table) Clear() {
	t.table = make(map[int64]*list.Element)
	t.list.Init()
}

func (t *Table) ProbeDepth(key int64, depth int16) (e *TableEntry, ok bool) {
	t.m.Lock()
	defer t.m.Unlock()
	elem, ok := t.table[key]
	if !ok {
		return nil, false
	}
	t.list.MoveToBack(elem)
	if e := elem.Value.(*TableEntry); depth <= e.Depth {
		return e, true
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

// locks excluded: t.m
func (t *Table) remove(e *list.Element) {
	delete(t.table, t.list.Remove(e).(*TableEntry).ZHash)
}

// locks excluded: t.m
func (t *Table) pop() {
	t.remove(t.list.Front())
}

func (t *Table) Store(e *TableEntry) {
	t.m.Lock()
	defer t.m.Unlock()
	if t.Cap() == 0 {
		return
	}
	if elem, ok := t.table[e.ZHash]; ok {
		t.list.MoveToBack(elem)
		// TODO(ajzaff): I just experimented with using an always rewrite strategy
		// which seemed to reduce the EBF somewhat. More testing is needed to determine
		// the right strategy.
		if step := elem.Value.(*TableEntry).Step; step != nil && e.Step == nil {
			// Preserve any old steps for this position.
			e.Step = step
		}
		elem.Value = e
		return
	}
	for t.Len() >= t.Cap() {
		t.pop()
	}
	t.table[e.ZHash] = t.list.PushBack(e)
}

func (t *Table) StoreMove(p *Pos, depth int16, value Value, move []Step) {
	for _, step := range move {
		entry := &TableEntry{
			Bound: ExactBound,
			ZHash: p.zhash,
			Depth: depth,
			Value: value,
			Step:  new(Step),
		}
		*entry.Step = step
		t.Store(entry)
		defer func() {
			if err := p.Unstep(); err != nil {
				panic(fmt.Sprintf("store_unstep: %v", err))
			}
		}()
		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("store_step: %s: %s: %v", move, step, err))
		}
	}
}

// Best returns the best move by probing the table.
// This is similar to PV but only returns steps for the current side.
func (t *Table) Best(p *Pos) (move []Step, score Value, err error) {
	initSide := p.Side()
	for i := 0; initSide == p.Side(); i++ {
		e, ok := t.ProbeDepth(p.zhash, 0)
		if !ok || e.Step == nil {
			break
		}
		if i == 0 {
			score = e.Value
		}
		if err := p.Step(*e.Step); err != nil {
			return move, score, err
		}
		move = append(move, *e.Step)
		defer func() {
			if err := p.Unstep(); err != nil {
				panic(fmt.Errorf("best_unstep: %v", err))
			}
		}()
	}
	return move, score, nil
}
