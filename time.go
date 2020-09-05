package zoo

import "time"

// TimeLimits configures game timing limits
// and some move limits. All times in seconds.
// 0 means unlimited and is the default value.
type TimeLimits struct {
	// Move is the time config per-move time.
	// Set using tcmove [seconds].
	Move time.Duration

	// Turn is the max turn time.
	// Set using tcturn [seconds].
	Turn time.Duration

	// Reserve is the time config initial reserve time.
	// Set using tcreserve [seconds].
	Reserve time.Duration

	// Percent of move time added to reserve.
	// Set using tcpercent [%].
	Percent int

	// Max reserve time.
	// Set using tcmax [seconds].
	Max time.Duration

	// Total max game duration.
	// Set using tctotal [seconds].
	Total time.Duration

	// Turns is the max number of turns in a game.
	// Set using tcturns [N].
	Turns int
}

// TimeInfo is the current time state for the engine.
type TimeInfo struct {
	// MoveUsed is the amount of time used on the current move.
	MoveUsed time.Duration

	// LastMoveUsed is the amount of time used on the last move.
	LastMoveUsed time.Duration

	// Start time of the last turn for gold and silver.
	Start [2]time.Time

	// Used time of the last turn for gold and silver.
	Used [2]time.Duration

	// Reserve remaining for gold and silver.
	Reserve [2]time.Duration

	// Nodes is the number of nodes seen at the last three
	// depths. This is used to estimate the EBF and the amount
	// of time to use on search.
	Nodes [3]int
}

func (e *Engine) RemainingTime() time.Duration {
	return 0
}
