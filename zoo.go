package zoo

import "log"

// Value is a score assigned to a position or move to represent its goodness.
// Higher numbers are better. Values range from [-1, 1]. Win is 1 and loss is
// -1. Inf is provided as value outside the range useful for initialization.
type Value float32

// Inf is an infinite value outside the range of normal use.
const Inf Value = 999

// Win represents a winning evaluation in the case of positional eval or an
// "only move" value in the policy setting.
const Win Value = 1

// Loss is equal to -Win and represents a losing position in positional eval.
// It is more canonical than stating "-Win". Loss is not used in policy scoring.
const Loss Value = -Win

// Valid returns whether the Value v has valid information.
func (v Value) Valid() bool {
	return v >= Loss && v <= Win
}

// Infinite returns whether the Value v is at or exceeds either Inf or -Inf.
func (v Value) Infinite() bool {
	return v >= Inf || v <= -Inf
}

// Terminal returns whether the Value v is a Win or Loss value.
func (v Value) Terminal() bool {
	return v.Win() || v.Loss()
}

// Win returns whether the Value v is a win.
func (v Value) Win() bool {
	return v == Win
}

// Loss returns whether the Value v is a loss.
func (v Value) Loss() bool {
	return v == Loss
}

func assert(message string, cond bool) {
	if !cond {
		panic(message)
	}
}

func passert(p *Pos, message string, cond bool) {
	if !cond {
		ppanic(p, message)
	}
}

func ppanic(p *Pos, v interface{}) {
	log.Println(p.String())
	log.Println(p.ShortString())
	log.Println(p.moves.String())
	panic(v)
}
