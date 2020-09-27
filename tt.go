package zoo

// TTEntry is the transposition table entry.
// The layout is made as compact as possible to save memory.
type TTEntry struct {
	// Key is the upper part of the ZHash (16 bits).
	Key16 uint16

	// Value is the estimated value of this node from the alpha network (16 bits).
	Value Value

	// Playouts is the number of playouts that passed through this node divided by 256 (8 bits).
	Playouts uint8

	// Gen8 contains the aging parameter (8 bits).
	Gen8 uint8

	// Best step packed into a uint16 (16 bits).
	Step Step
}

// Clear zeroes the TableEntry and resets to its initial state.
func (e *TTEntry) Clear() {
	e.Key16 = 0
	e.Value = 0
	e.Playouts = 0
	e.Gen8 = 0
	e.Step = 0
}

// Save the information into the TableEntry if it is more valuable.
func (e *TTEntry) Save(key uint64, v Value, pv bool, gen, runs uint8, step Step) {
	key16 := uint16(key >> 48)
	if step != 0 || key16 != e.Key16 {
		// Preserve step information. Only reset step if valid
		// and on key change (modulo Type 1 key errors).
		e.Step = step
	}
	// Overwrite more valuable entries.
	if key16 != e.Key16 || runs > e.Playouts {
		e.Key16 = key16
		e.Value = v
		g := gen
		if pv {
			g |= 1 << 2
		}
		e.Gen8 = g
	}
}

// TranspositionTable is the main clustered transposition table for the engine.
// The clustered table structure was adapted from Stockfish but the functional
// parts have been retrofitted from AlphaBeta to suit Monte Carlo. Instead of
// move bounds, we cache the results of experiment runs. Depth is replaced with
// the number of trials used to achieve the value.
type TranspositionTable struct {
	cap          int
	clusterCount int
	gen8         uint8 // aging parameter
	table        []tableCluster
}

const clusterSize = 3

// TableCluster is a cluster of the table storing the entries.
type tableCluster struct {
	entries [clusterSize]TTEntry
}

// cluster uses the 32 lowest order bits of the key to determine the cluster index.
func (t *TranspositionTable) cluster(key uint64) *tableCluster {
	return &t.table[(uint64(uint32(key))*uint64(t.clusterCount))>>32]
}

// Clear clears the hash table.
// A call to Clear during active search is problematic and should be prevented.
func (t *TranspositionTable) Clear() {
	// TODO(ajzaff): Clear using multiple goroutines.
	for i := 0; i < t.clusterCount; i++ {
		for j := 0; j < clusterSize; j++ {
			t.table[i].entries[j].Clear()
		}
	}
}

// Resize the table by reallocating a new table of the specified size in MB.
// A call to Resize during active search is problematic and should be prevented.
func (t *TranspositionTable) Resize(mbSize int) {
	t.Clear()
	t.clusterCount = mbSize * 1024 * 1024 / 36 // sizeof(tableCluster)
	t.table = make([]tableCluster, t.clusterCount)
}

// GlobalAge returns the global cyclic age parameter of the table which affects how entries are evicted.
func (t *TranspositionTable) GlobalAge() uint8 {
	return t.gen8
}

// NewSearch is called before a new search to increase the GlobalAge of the table.
func (t *TranspositionTable) NewSearch() {
	t.gen8 += 8
}

// Probe returns the entry matching the key and found = true, or returns
// the least valuable entry (to be overwritten with a call to Save) and found = false.
func (t *TranspositionTable) Probe(key uint64) (e *TTEntry, found bool) {
	key16 := uint16(key >> 48)
	cluster := t.cluster(key)

	for i := 0; i < clusterSize; i++ {
		e := &cluster.entries[i]
		if e.Key16 == key16 {
			e.Gen8 = t.gen8 // Refresh
			return e, true
		}
	}

	// Find an entry to be replaced according to the replacement strategy.
	replace := &cluster.entries[0]
	for i := 1; i < clusterSize; i++ {
		e := &cluster.entries[i]
		// Pick least valuable entry whilst handling cyclic generation overflow.
		// See stockfish/tt.cpp for explaination.
		if replace.Playouts-((uint8(263+int(t.gen8))-e.Gen8)&0xf8) >
			e.Playouts-((uint8(263+int(t.gen8))-e.Gen8)&0xf8) {
			replace = e
		}
	}
	return replace, false
}

// Hashfull approximates the hashtable fullness (per million sampled entries).
func (t *TranspositionTable) Hashfull() int {
	cnt := 0
	for i := 0; i < 1000/clusterSize; i++ {
		for j := 0; j < clusterSize; j++ {
			if t.table[i].entries[j].Gen8&0xf8 == t.gen8 {
				cnt++
			}
		}
	}
	return cnt * 1000 / (clusterSize * (1000 / clusterSize))
}
