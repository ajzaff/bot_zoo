package zoo

import "sync"

// Threefold tracks the number of repetitions of a given position.
// It provides the methods Lookup, Increment, and Decrement which
// manage the number of times the given position key occurs and Store
// incrementing after a move is made, and decrementing when unmade.
// Threefold is safe for concurrent use.
type Threefold struct {
	m sync.Map // zhash => repetition count
}

// NewThreefold creates a new threefold repetition map.
func NewThreefold() *Threefold {
	return &Threefold{}
}

// Clone returns a deep copy of t.
func (t *Threefold) Clone() *Threefold {
	tf := &Threefold{}
	t.m.Range(func(key, value interface{}) bool {
		tf.m.Store(key, value)
		return true
	})
	return tf
}

// Clear reinitializes the repetition map removing all stored positions.
func (t *Threefold) Clear() {
	t.m = sync.Map{}
}

// Lookup the repetition count for the given position key.
func (t *Threefold) Lookup(key Hash) int {
	v, ok := t.m.Load(key)
	if !ok {
		return 0
	}
	return v.(int)
}

// Increment the position key and return the number of repetitions.
func (t *Threefold) Increment(key Hash) int {
	v, ok := t.m.LoadOrStore(key, 1)
	if !ok {
		return 1
	}
	n := v.(int) + 1
	t.m.Store(key, n)
	return n
}

// Decrement the position key and return the number of repetitions.
func (t *Threefold) Decrement(key Hash) int {
	v, ok := t.m.LoadOrStore(key, 0)
	if !ok {
		return 0
	}
	n := v.(int) - 1
	if n <= 0 {
		n = 0
		t.m.Delete(key)
	} else {
		t.m.Store(key, n)
	}
	return n
}
