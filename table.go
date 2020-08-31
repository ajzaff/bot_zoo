package zoo

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
	head  *TableEntry
	tail  *TableEntry
	table map[int64]*TableEntry
	len   int
	cap   int
}

func NewTable(cap int) *Table {
	return &Table{
		table: make(map[int64]*TableEntry),
		cap:   cap,
	}
}

func (t *Table) Clear() {
	t.table = make(map[int64]*TableEntry)
	t.len = 0
}

func (t *Table) ProbeDepth(key int64, depth int) (e *TableEntry, ok bool) {
	if e, ok := t.table[key]; ok {
		t.pop(e)
		t.emplaceBack(e)
		// TODO(ajzaff): Change this to `e.Depth >= depth`
		// once quiescence search is added.
		if e.Depth == depth {
			return e, true
		}
	}
	return nil, false
}

func (t *Table) Len() int {
	return t.len
}

func (t *Table) Cap() int {
	return t.cap
}

func (t *Table) pop(e *TableEntry) {
	delete(t.table, e.ZHash)
	if t.Len() == 1 {
		t.head = nil
		t.tail = nil
		t.len--
		return
	}
	if e == t.head {
		t.head = e.next
	} else if e.prev != nil {
		e.prev.next = e.next
	}
	if e == t.tail {
		t.tail = e.prev
	}
	t.len--
}

func (t *Table) emplaceBack(e *TableEntry) {
	if t.Len() == 0 {
		t.head = e
		t.tail = e
		t.tail.prev = t.head
		t.len++
		return
	}
	t.tail.next = e
	e.prev = t.tail
	e.next = nil
	t.tail = e
	t.len++
}

func (t *Table) fixedPurge(curDepth int) {
	t.pop(t.head)
	n := t.Len() / 2
	for it := t.head; it != nil && it.Depth < curDepth && n > 0; it, n = it.next, n-1 {
		t.pop(it)
	}
}

func (t *Table) Store(e *TableEntry) {
	old, ok := t.table[e.ZHash]
	if ok && old.Depth >= e.Depth {
		return
	}
	t.table[e.ZHash] = e
	if t.Len() >= t.Cap() {
		t.fixedPurge(e.Depth)
	}
	t.emplaceBack(e)
}
