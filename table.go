package zoo

type Bound int

const (
	NoBound Bound = iota
	UpperBound
	LowerBound
	ExactBound
)

type Entry struct {
	Bound
	Depth int
	Value int
}

type Table struct {
	table map[int64]*Entry
	data  []*Entry
	moves [][]Step
}

func NewTable() Table {
	return Table{
		table: make(map[int64]*Entry),
	}
}

func (t Table) Clear() {

}

func (t Table) ProbeDepth(key int64, depth int) (r *SearchResult, ok bool) {
	return nil, false

	// r, ok = t.data[key]
	// if ok && r.Depth >= depth {
	// 	return r, true
	// }
	// return nil, false
}

func (t Table) Store(key int64, value *SearchResult) {
	// r, ok := t.data[key]
	// if !ok {
	// 	t.data[key] = value
	// 	return
	// }
	// if value.Depth > r.Depth {
	// 	*r = *value
	// }
}
