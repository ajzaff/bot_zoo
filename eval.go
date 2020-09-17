package zoo

// Value is a score assigned to a position or move to represent its goodness.
// Higher numbers are better. Depending on the context Values may range from
// [0,1] (as in policies) or [-1,1] as in position eval. Inf is provided as
// value outside the range useful for initialization.
type Value float64

// Inf is an infinite value outside the range of normal use.
const Inf Value = 999

// Win represents a winning evaluation in the case of positional eval or an
// "only move" value in the policy setting.
const Win Value = 1

// Loss is equal to -Win and represents a losing position in positional eval.
// It is more canonical than stating "-Win". Loss is not used in policy scoring.
const Loss Value = -Win

// None returns whether the Value v has valid information. It is functionally
// the same as `!v.Infinite()``.
func (v Value) None() bool {
	return !v.Infinite()
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
	return v >= Win
}

// Loss returns whether the Value v is a loss.
func (v Value) Loss() bool {
	return v <= Loss
}
