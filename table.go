package zoo

type Entry struct {
	Best []Step
}

type Table struct {
	data map[int64]*SearchResult
}

func NewTable() Table {
	return Table{
		data: make(map[int64]*SearchResult),
	}
}

func (t Table) GetDepth(key int64, depth int) (r *SearchResult, ok bool) {
	return nil, false

	// r, ok = t.data[key]
	// if ok && r.Depth >= depth {
	// 	return r, true
	// }
	// return nil, false
}

func (t Table) Update(key int64, value *SearchResult) {
	// r, ok := t.data[key]
	// if !ok {
	// 	t.data[key] = value
	// 	return
	// }
	// if value.Depth > r.Depth {
	// 	*r = *value
	// }
}
