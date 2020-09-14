package zoo

type Bound uint8

const (
	BoundNone Bound = iota
	BoundUpper
	BoundLower
	BoundExact
)

// TableEntry is the transposition table entry.
// The layout is made as compact as possible to save memory.
type TableEntry struct {
	// Key is the upper part of the ZHash (16 bits).
	Key16 uint16
	// Value of the entry (16 bits).
	Value Value
	// Eval of the entry (16 bits).
	Eval Value
	// Depth of the entry (8 bits).
	Depth uint8
	// Gen8Bound packs the bound type, PV, and aging parameter (8 bits).
	Gen8Bound uint8
	// Step packed into a uint32 (32 bits).
	Step Step
}

func (e *TableEntry) Clear() {
	e.Key16 = 0
	e.Value = 0
	e.Eval = 0
	e.Depth = 0
	e.Gen8Bound = 0
	e.Step = 0
}

// PV returns whether the entry is a PV entry.
func (e *TableEntry) PV() bool {
	return e.Gen8Bound&4 != 0
}

// Bound extracts the bound from this entry.
func (e *TableEntry) Bound() Bound {
	return Bound(e.Gen8Bound & 0x3)
}

// Save the information into the TableEntry if it is more valuable.
func (e *TableEntry) Save(key uint64, v, ev Value, pv bool, b Bound, gen, depth uint8, step Step) {
	key16 := uint16(key >> 48)
	if step.Kind() != KindInvalid || key16 != e.Key16 {
		// Preserve step information. Only reset step if valid
		// and on key change (modulo Type 1 key errors).
		e.Step = step
	}
	// Overwrite more valuable entries.
	if key16 != e.Key16 || depth > e.Depth-4 || b == BoundExact {
		e.Key16 = key16
		e.Value = v
		e.Eval = ev
		g := gen
		if pv {
			g |= 1 << 2
		}
		e.Gen8Bound = g | uint8(b)
		e.Depth = depth
	}
}

// Table is the main clustered transposition table for the engine.
type Table struct {
	cap          int
	clusterCount int
	gen8         uint8 // aging parameter
	table        []*tableCluster
}

const clusterSize = 3

// TableCluster is a cluster of the table storing the entries.
type tableCluster struct {
	entries [clusterSize]TableEntry
}

// cluster uses the 32 lowest order bits of the key to determine the cluster index.
func (t *Table) cluster(key uint64) *tableCluster {
	return t.table[(uint64(uint32(key))*uint64(t.clusterCount))>>32]
}

// NewTableSize returns a table with the given mbSize.
func NewTableSize(mbSize int) *Table {
	t := &Table{}
	t.Resize(defaultSizeMB)
	return t
}

const defaultSizeMB = 50

// NewTable returns a table with the default size.
func NewTable() *Table {
	return NewTableSize(defaultSizeMB)
}

// Clear clears the hash table.
// A call to Clear during active search is problematic and should be prevented.
func (t *Table) Clear() {
	// TODO(ajzaff): Clear using multiple goroutines.
	for i := 0; i < t.clusterCount; i++ {
		for j := 0; j < clusterSize; j++ {
			t.table[i].entries[j].Clear()
		}
	}
}

// Resize the table by reallocating a new table of the specified size in MB.
// A call to Resize during active search is problematic and should be prevented.
func (t *Table) Resize(mbSize int) {
	t.Clear()
	t.clusterCount = mbSize * 1024 * 1024 / 36 // sizeof(tableCluster)
	t.table = make([]*tableCluster, t.clusterCount)
	for i := 0; i < t.clusterCount; i++ {
		t.table[i] = new(tableCluster)
	}
}

func (t *Table) NewSearch() {
	t.gen8 += 8
}

// Probe returns the entry matching the key and found = true, or returns
// the least valuable entry (to be overwritten with a call to Save) and found = false.
func (t *Table) Probe(key uint64) (e *TableEntry, found bool) {
	key16 := uint16(key >> 48)
	cluster := t.cluster(key)

	for i := 0; i < clusterSize; i++ {
		e := &cluster.entries[i]
		if e.Key16 == 0 || e.Key16 == key16 {
			e.Gen8Bound = uint8(t.gen8 | (e.Gen8Bound & 0x7)) // Refresh
			return e, e.Key16 == key16
		}
	}

	// Find an entry to be replaced according to the replacement strategy.
	replace := &cluster.entries[0]
	for i := 1; i < clusterSize; i++ {
		e := &cluster.entries[i]
		// Pick least valuable entry whilst handling cyclic generation overflow.
		// See stockfish/tt.cpp for explaination.
		if e.Depth-((uint8(263+int(t.gen8))-e.Gen8Bound)&0xf8) >
			e.Depth-((uint8(263+int(t.gen8))-e.Gen8Bound)&0xf8) {
			replace = e
		}
	}
	return replace, false
}

// Hashfull approximates the hashtable fullness (per million sampled entries).
func (t *Table) Hashfull() int {
	cnt := 0
	for i := 0; i < 1000/clusterSize; i++ {
		for j := 0; j < clusterSize; j++ {
			if t.table[i].entries[j].Gen8Bound&0xf8 == t.gen8 {
				cnt++
			}
		}
	}
	return cnt * 1000 / (clusterSize * (1000 / clusterSize))
}
