package zoo

var goalRanks = [2]Bitboard{
	^NotRank8, // Gold
	^NotRank1, // Silver
}

var goalRange = [2]Bitboard{
	^NotRank8 | ^NotRank8>>8 | ^NotRank8>>16, // Gold
	^NotRank1 | ^NotRank1<<8 | ^NotRank1<<16, // Silver
}

func scoreStep(p *Pos, step Step) Value {
	src := step.Src()
	if !src.Valid() {
		// Ignore likely setup move.
		return 0
	}
	dest := step.Dest()
	assert("dest is invalid", dest.Valid())
	piece1 := p.At(src)
	side := piece1.Color()
	var value Value
	// Add +Inf - |step| for goal moves.
	if piece1.SameType(GRabbit) && dest.Bitboard()&goalRanks[side] != 0 {
		value += +Inf - Value(step.Len()) // find shortest mate
	}
	// Add O(large) for rabbit moves close to goal.
	if piece1.SameType(GRabbit) && dest.Bitboard()&goalRange[side] != 0 { // Coarse goal threat:
		value += 2000
	}
	// Add static value of capture:
	if step.Capture() {
		if t := step.Cap(); t.Color() == side {
			value -= 1000
		} else {
			value += 1000
		}
	}
	return value
}
