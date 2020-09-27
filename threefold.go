package zoo

import (
	"log"
)

// Threefold tracks the number of repetitions of a given position.
// It provides the methods Lookup, Increment, and Decrement which
// manage the number of times the given position key occurs and Store
// incrementing after a move is made, and decrementing when unmade.
type Threefold struct {
	m map[Hash]int // Pos hash => repetition count
}

// NewThreefold creates a new threefold repetition map.
func NewThreefold() *Threefold {
	return &Threefold{
		m: make(map[Hash]int),
	}
}

// Clone returns a deep copy of t.
func (t *Threefold) Clone() *Threefold {
	tf := NewThreefold()
	for k, v := range t.m {
		tf.m[k] = v
	}
	return tf
}

// Clear reinitializes the repetition map removing all stored positions.
func (t *Threefold) Clear() {
	t.m = make(map[Hash]int)
}

// Lookup the repetition count for the given position key.
func (t *Threefold) Lookup(key Hash) int {
	v, ok := t.m[key]
	if !ok {
		return 0
	}
	return v
}

// Increment the position key and return the number of repetitions.
func (t *Threefold) Increment(key Hash) int {
	v := t.m[key] + 1
	t.m[key] = v
	return v
}

// Decrement the position key and return the number of repetitions.
func (t *Threefold) Decrement(key Hash) int {
	v := t.m[key] - 1
	if v <= 0 {
		v = 0
		delete(t.m, key)
	} else {
		t.m[key] = v
	}
	return v
}

// Debug prints the position information to the logger l.
func (t *Threefold) Debug(l *log.Logger) {
	l.Println("threefold (>=2):")
	for k, v := range t.m {
		if v >= 2 {
			l.Printf("  %d=%d", k, v)
		}
	}
}
