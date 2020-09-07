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
	Depth int
	Value int
	pv    bool
	Step  *Step
}

type Table struct {
	cap   int
	list  *list.List
	table sync.Map // zhash => *list.Element
	m     sync.RWMutex
}

func NewTable(cap int) *Table {
	return &Table{
		cap:  cap,
		list: list.New(),
	}
}

func (t *Table) Clear() {
	t.table = sync.Map{}
	t.list.Init()
}

func (t *Table) ProbeDepth(key int64, depth int) (e *TableEntry, ok bool) {
	elem, ok := t.table.Load(key)
	if !ok {
		return nil, false
	}
	t.m.Lock()
	defer t.m.Unlock()
	t.list.MoveToBack(elem.(*list.Element))
	if e := elem.(*list.Element).Value.(*TableEntry); depth <= e.Depth {
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
	t.list.Remove(e)
	t.table.Delete(e.Value.(*TableEntry).ZHash)
}

// locks excluded: t.m
func (t *Table) pop() {
	elem := t.list.Front()
	for elem.Value.(*TableEntry).pv {
		next := elem.Next()
		if next == nil {
			break
		}
		t.list.MoveToBack(elem)
		elem = next
	}
	e := t.list.Remove(elem)
	t.table.Delete(e.(*TableEntry).ZHash)
}

func (t *Table) Store(e *TableEntry) {
	if v, ok := t.table.Load(e.ZHash); ok {
		elem := v.(*list.Element)
		t.list.MoveToBack(elem)
		// TODO(ajzaff): I just experimented with using an always rewrite strategy
		// which seemed to reduce the EBF somewhat. More testing is needed to determine
		// the right strategy.
		elem.Value = e
		return
	}
	t.m.Lock()
	defer t.m.Unlock()
	for t.Len() >= t.Cap() {
		t.pop()
	}
	t.table.Store(e.ZHash, t.list.PushBack(e))
}

// Best returns the best move by probing the table.
// This is similar to PV but only returns steps for the current side.
func (t *Table) Best(p *Pos) (move []Step, score int, err error) {
	initSide := p.Side()
	for i := 0; initSide == p.Side(); i++ {
		e, ok := t.ProbeDepth(p.zhash, 0)
		if !ok || e.Step == nil {
			break
		}
		if i == 0 {
			score = e.Value
		}
		move = append(move, *e.Step)
		if err := p.Step(*e.Step); err != nil {
			return nil, 0, err
		}
		defer func() {
			if err := p.Unstep(); err != nil {
				panic(fmt.Errorf("best_unstep: %v", err))
			}
		}()
	}
	return move, score, nil
}

// PV returns the principal variation by probing the table.
// Despite the name it can return scores that are outside the window.
// The PV has a maximum length of 50 steps.
func (t *Table) PV(p *Pos) (pv []Step, score int, err error) {
	for i := 0; i < 50; i++ {
		e, ok := t.ProbeDepth(p.zhash, 0)
		if !ok || e.Step == nil {
			break
		}
		if i == 0 {
			score = e.Value
		}
		pv = append(pv, *e.Step)
		if err := p.Step(*e.Step); err != nil {
			return nil, 0, err
		}
		defer func() {
			if err := p.Unstep(); err != nil {
				panic(fmt.Errorf("PV_unstep: %v", err))
			}
		}()
	}
	return pv, score, nil
}
