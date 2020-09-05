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
		t.remove(e)
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

func (t *Table) remove(e *TableEntry) {
	n := t.Len()
	if n == 1 {
		t.head = nil
		t.tail = nil
		t.len--
		return
	}
	if e == t.head {
		t.pop()
		return
	}
	delete(t.table, e.ZHash)
	e.prev.next = e.next
	if e == t.tail {
		t.tail = e.prev
	}
	t.len--
}

func (t *Table) pop() {
	n := t.Len()
	if n < 2 {
		if n == 0 {
			return
		}
		if n == 1 {
			t.head = nil
			t.tail = nil
			t.len--
			return
		}
	}
	delete(t.table, t.head.ZHash)
	t.head = t.head.next
	t.head.prev = nil
	if n == 2 {
		t.tail = t.head
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

func (t *Table) Store(e *TableEntry) {
	old, ok := t.table[e.ZHash]
	if ok && old.Depth >= e.Depth {
		return
	}
	t.table[e.ZHash] = e
	for t.Len() >= t.Cap() {
		t.pop()
	}
	t.emplaceBack(e)
}
