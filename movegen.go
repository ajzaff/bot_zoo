package zoo

var (
	stepsB       [64]Bitboard
	rabbitStepsB [2][64]Bitboard
)

func init() {
	for i := Square(0); i < 64; i++ {
		b := i.Bitboard()
		steps := b.Neighbors()
		stepsB[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b >> 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		rabbitStepsB[Gold][i] = grSteps
		rabbitStepsB[Silver][i] = srSteps
	}
}

// func (p *Pos) pieceSteps(piece Piece, table [64]Bitboard) []Step {
// 	var steps []Step
// 	piece = piece.MakeColor(p.side)
// 	p.bitboards[piece].Each(func(b Bitboard) {
// 		if p.frozenB(b) {
// 			return
// 		}
// 		src := b.Square()
// 		stepsB[src].Each(func(d Bitboard) {
// 			if p.bitboards[Empty]&d == 0 {
// 				return
// 			}
// 			dest := d.Square()
// 			steps = append(steps, p.completeCapture(Step{
// 				Src:    src,
// 				Dest:   dest,
// 				Alt:    invalidSquare,
// 				Piece1: piece,
// 			}))
// 		})
// 	})
// 	return steps
// }

// func (p *Pos) getPulls(steps *[]Step) {
// 	s2 := p.side.Opposite()

// 	if len(p.steps) == 0 {
// 		return
// 	}
// 	i := len(p.steps) - 1
// 	for ; i >= 0 && p.steps[i].Capture(); i-- {
// 	}
// 	assert("i < 0", i >= 0)
// 	t1 := p.steps[i].Piece1
// 	if t1.SameType(GRabbit) {
// 		return
// 	}
// 	dest := p.steps[i].Src
// 	db := dest.Bitboard()

// 	if p.presence[s2].Neighbors()&db == 0 {
// 		return
// 	}

// 	for t2 := GRabbit.MakeColor(s2); t2 < t1; t2++ {
// 		t2b := p.bitboards[t2]
// 		if t2b == 0 || t2b.Neighbors()&db == 0 {
// 			continue
// 		}
// 		for t2b > 0 {
// 			b := t2b & -t2b
// 			t2b &= ^b
// 			src := b.Square()
// 			assert("src == dest", src != dest)
// 			*steps = append(*steps, p.completeCapture(Step{
// 				Src:    src,
// 				Dest:   dest,
// 				Alt:    invalidSquare,
// 				Piece1: t2,
// 			}))
// 		}
// 	}
// }

// func (p *Pos) getPushes(steps *[]Step) {
// 	s1, s2 := p.side, p.side.Opposite()
// 	s2n := p.presence[s2].Neighbors()
// 	for p1 := GCat.MakeColor(s1); p1 <= GElephant.MakeColor(s1); p1++ {
// 		p1b := p.bitboards[p1]
// 		if p1b == 0 || p1b&s2n == 0 {
// 			continue
// 		}
// 		p1n := p1b.Neighbors()
// 		for p2 := GRabbit.MakeColor(s2); p2 < p1.MakeColor(s2); p2++ {
// 			p2b := p.bitboards[p2]
// 			if p2b == 0 {
// 				continue
// 			}
// 			(p1n & p2b).Each(func(b Bitboard) {
// 				if p.frozenB(b) {
// 					return
// 				}
// 				src := b.Square()
// 				sB := stepsB
// 				if p2.SameType(GRabbit) {
// 					sB = rabbitStepsB[p2.Color()]
// 				} else {
// 					sB = stepsB
// 				}
// 				sB[src].Each(func(d Bitboard) {
// 					if p.bitboards[Empty]&d == 0 {
// 						return
// 					}
// 					dest := d.Square()
// 					assert("src == dest", src != dest)
// 					*steps = append(*steps, Step{
// 						Src:    src,
// 						Dest:   dest,
// 						Piece1: p2,
// 					})
// 				})
// 			})
// 		}
// 	}
// }

// func (p *Pos) completePush(steps *[]Step) {
// 	i := len(p.steps) - 1
// 	for ; i >= 0 && p.steps[i].Capture(); i-- {
// 	}
// 	assert("i < 0", i >= 0)
// 	push := p.steps[i]
// 	p2 := push.Piece1
// 	dest := push.Src
// 	destB := dest.Bitboard()

// 	for p1 := p2.MakeColor(p.side) + 1; p1 < GElephant.MakeColor(p.side); p1++ {
// 		p1b := p.bitboards[p1]
// 		if p1b == 0 || p1b.Neighbors()&destB == 0 {
// 			continue
// 		}
// 		for p1b > 0 {
// 			b := p1b & -p1b
// 			p1b &= ^b
// 			src := b.Square()
// 			assert("src == dest", src != dest)
// 			*steps = append(*steps, Step{
// 				Src:    src,
// 				Dest:   dest,
// 				Piece1: p2,
// 			})
// 		}
// 	}
// }

func unguardedB(b, presence Bitboard) Bitboard {
	return b & ^presence.Neighbors()
}

func trappedB(b, presence Bitboard) Bitboard {
	return b & unguardedB(Traps, presence)
}

// captureB determines if a move from srcB to destB would result in a capture
// without making the move. If so, the piece and square of the captured piece is returned.
func (p *Pos) capture(presence, srcB, destB Bitboard) Capture {
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		return Capture{
			Piece: p.atB(srcB),
			Src:   destB.Square(),
		}
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		return Capture{
			Piece: p.atB(b),
			Src:   b.Square(),
		}
	}
	return Capture{}
}

// Steps generates steps including sliding moves, pushes and pulls.
// Captures are completed internally to Step, but you may check whether
// a step results in a capture by calling captures with the step as an
// argument.
func (p *Pos) Steps() []Step {
	if p.stepsLeft == 0 {
		return []Step{{Pass: true}}
	}
	var steps []Step
	canPush := p.stepsLeft > 1
	lo, hi := GRabbit, SElephant
	c1, c2 := p.side, p.side.Opposite()
	_ = c2
	if !canPush {
		lo, hi = GRabbit.MakeColor(c1), GElephant.MakeColor(c1)
	}
	for t := lo; t <= hi; t++ {
		if !t.Valid() {
			continue
		}
		ts := p.bitboards[t]
		if ts == 0 {
			continue
		}
		// Generate sliding steps.
		if t.Color() == c1 {
			sB := stepsB

			// My rabbits can only step forward.
			if t.SameType(GRabbit) {
				sB = rabbitStepsB[p.side]
			}
			ts.Each(func(sb Bitboard) {
				src := sb.Square()
				sB[src].Each(func(db Bitboard) {
					if p.bitboards[Empty]&db == 0 {
						return
					}
					dest := db.Square()
					step := Step{
						Src:    src,
						Dest:   dest,
						Alt:    invalidSquare,
						Piece1: t,
					}
					if cap := p.capture(p.presence[c1], sb, db); cap.Valid() {
						step.Cap = cap
					}
					steps = append(steps, step)
				})
			})
		}
	}
	if p.stepsLeft < 4 {
		steps = append(steps, Step{Pass: true})
	}
	return steps
}

func (p *Pos) getRootMoves(prefix []Step, moves *[][]Step, movesLeft int) {
	if movesLeft == 0 {
		if !prefix[len(prefix)-1].Pass {
			prefix = append(prefix, Step{Pass: true})
		}
		*moves = append(*moves, prefix)
		return
	}
	assert("movesLeft < 0", movesLeft > 0)
	for _, step := range p.Steps() {
		move := append(prefix, step)
		if step.Pass {
			*moves = append(*moves, move)
			break
		}
		if err := p.Step(step); err != nil {
			panic(err)
		}
		newPrefix := make([]Step, len(move))
		copy(newPrefix, move)
		p.getRootMoves(newPrefix, moves, movesLeft-1)
		if err := p.Unstep(); err != nil {
			panic(err)
		}
	}
}

func (e *Engine) getRootMovesLen(p *Pos, depth int) [][]Step {
	if depth <= 0 || depth > 4 {
		panic("depth <= 0 || depth > 4")
	}
	var moves [][]Step
	for i := 1; i <= depth; i++ {
		p.getRootMoves(nil, &moves, i)
	}
	return moves
}
