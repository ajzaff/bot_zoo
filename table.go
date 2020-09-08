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
	Step  *Step
}

type Table struct {
	cap   int
	list  *list.List // guarded by m
	table sync.Map   // zhash => *list.Element
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
	t.table.Delete(t.list.Remove(e).(*TableEntry).ZHash)
}

// locks excluded: t.m
func (t *Table) pop() {
	t.remove(t.list.Front())
}

func (t *Table) Store(e *TableEntry) {
	if t.Cap() == 0 {
		return
	}
	if v, ok := t.table.Load(e.ZHash); ok {
		elem := v.(*list.Element)
		t.m.Lock()
		defer t.m.Unlock()
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
	t.m.Lock()
	defer t.m.Unlock()
	for t.Len() >= t.Cap() {
		t.pop()
	}
	t.table.Store(e.ZHash, t.list.PushBack(e))
}

func (t *Table) StoreMove(p *Pos, depth, score int, move []Step) {
	for _, step := range move {
		entry := &TableEntry{
			Bound: ExactBound,
			ZHash: p.zhash,
			Depth: depth,
			Value: score,
			Step:  new(Step),
		}
		*entry.Step = step
		t.Store(entry)
		if err := p.Step(step); err != nil {
			panic(fmt.Sprintf("store_step: %s: %s: %v", move, step, err))
		}
	}
	for i := len(move) - 1; i >= 0; i-- {
		if err := p.Unstep(); err != nil {
			panic(fmt.Sprintf("store_step: %s: %s: %v", move, move[i], err))
		}
	}
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
		if err := p.Step(*e.Step); err != nil {
			// Ignore this error since we might not have stored a full step.
			return pv, score, nil
		}
		pv = append(pv, *e.Step)
		defer func() {
			if err := p.Unstep(); err != nil {
				panic(fmt.Errorf("PV_unstep: %v", err))
			}
		}()
	}
	return pv, score, nil
}
