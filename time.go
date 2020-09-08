package zoo

import (
	"math"
	"time"
)

// TimeInfo is the current time state for the engine.
type TimeInfo struct {
	// GameStart is the time of game start from after receiving newgame.
	GameStart time.Time

	// Start time of the last turn for gold and silver.
	Start [2]time.Time

	// Move time for gold and silver.
	Move [2]time.Duration

	// Reserve remaining for gold and silver.
	Reserve [2]time.Duration
}

// TimeControl configures game timing control
// and some move limits. All times in seconds.
// 0 means unlimited and is the default value.
type TimeControl struct {
	// Move is the time control per-move time.
	// Set using tcmove [seconds].
	Move time.Duration

	// Reserve is the time control initial reserve time.
	// Set using tcreserve [seconds].
	Reserve time.Duration

	// MoveReservePercent of move time added to reserve.
	// Set using tcpercent [%].
	MoveReservePercent int

	// MaxTurn is the max turn time.
	// Set using tcturn [seconds].
	MaxTurn time.Duration

	// MaxReserve reserve time.
	// Set using tcmax [seconds].
	MaxReserve time.Duration

	// GameTotal max game duration.
	// Set using tctotal [seconds].
	GameTotal time.Duration

	// Turns is the max number of turns in a game.
	// Set using tcturns [N].
	Turns int
}

// makeTimeControl creates a default blitz game time control equal to "30s/3/100/5/8".
// See http://arimaa.com/arimaa/learn/matchRules.html for time control notation.
func makeTimeControl() TimeControl {
	return TimeControl{
		Move:               60 * time.Second,
		Reserve:            3 * time.Minute,
		MoveReservePercent: 100,
		MaxReserve:         5 * time.Minute,
		GameTotal:          8 * time.Hour,
	}
}

func (tc TimeControl) newTimeInfo() *TimeInfo {
	now := time.Now()
	return &TimeInfo{
		GameStart: now,
		Start:     [2]time.Time{now, now},
		Move:      [2]time.Duration{tc.Move, tc.Move},
		Reserve:   [2]time.Duration{tc.Reserve, tc.Reserve},
	}
}

func (tc TimeControl) resetTurn(t *TimeInfo, c Color) {
	t.Move[c] = tc.Move
	t.Start[c] = time.Now()
}

func (tc TimeControl) gameTimeRemaining(t *TimeInfo, c Color, now time.Time) time.Duration {
	used := now.Sub(t.Start[c])
	rem := time.Duration(math.MaxInt64)

	// Check the remaining time against the turn limit timer.
	// If set, this is a hard limit.
	if tc.MaxTurn != 0 {
		rem = tc.MaxTurn - used
	}

	// Check the total game time.
	if tc.GameTotal != 0 {
		if r := tc.GameTotal - used; r < rem {
			rem = r
		}
	}

	// Check the move time remaining plus reserves.
	turn := t.Move[c] + t.Reserve[c]
	if r := turn - used; r < rem {
		rem = r
	}

	return rem
}

// GameTimeRemaining returns the maximum amount of time left for side c.
// This generally assumes side c is the side to move.
// If HardTimeRemaining elapses side c will lose on time.
func (tc TimeControl) GameTimeRemaining(t *TimeInfo, c Color) time.Duration {
	now := time.Now()
	return tc.gameTimeRemaining(t, c, now)
}

func (tc TimeControl) turnTimeRemaining(t *TimeInfo, c Color, now time.Time) time.Duration {
	used := now.Sub(t.Start[c])
	rem := tc.gameTimeRemaining(t, c, now)

	// Check the move time remaining without reserves.
	if r := t.Move[c] - used; r < rem {
		rem = r
	}

	return rem
}

// TurnTimeRemaining returns the amount of time remaining for the turn.
// This is a tighter bound than GameTimeRemaining, if this time elpases
// side c will not neccessarily lose the game so it's still necessary to
// fall back to GameTimeRemaining.
func (tc TimeControl) TurnTimeRemaining(t *TimeInfo, c Color) time.Duration {
	now := time.Now()
	return tc.turnTimeRemaining(t, c, now)
}

// FixedOptimalTimeRemaining tries to take a middle ground between game time
// and turn time with a reasonable fixed maximum move time.
func (tc TimeControl) FixedOptimalTimeRemaining(t *TimeInfo, c Color) time.Duration {
	now := time.Now()
	turn := tc.turnTimeRemaining(t, c, now)
	if turn < 0 {
		return turn
	}
	game := tc.gameTimeRemaining(t, c, now)
	if game < 30*time.Second {
		return game
	}
	game /= 3
	if game > 120*time.Second {
		return 120 * time.Second
	}
	return game
}
