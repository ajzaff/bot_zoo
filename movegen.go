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

func unguardedB(b, presence Bitboard) Bitboard {
	return b & ^presence.Neighbors()
}

func trappedB(b, presence Bitboard) Bitboard {
	return b & unguardedB(Traps, presence)
}

// capture computes statically the capture resulting from a move from src to dest if any.
// The possible types of captures are abandoning a trapped piece or
// capturing ourselves by stepping onto an unguarded trap square.
func (p *Pos) capture(presence Bitboard, src, dest Square) Capture {
	srcB := src.Bitboard()
	destB := dest.Bitboard()
	newPresence := presence ^ (srcB | destB)
	if b := unguardedB(destB&Traps, newPresence); b != 0 {
		// Capture of piece pushed from srcB to destB.
		return Capture{
			Piece: p.board[src],
			Src:   dest,
		}
	}
	if b := trappedB(newPresence&srcB.Neighbors(), newPresence); b != 0 {
		// Capture of piece left next to src.
		// TODO(ajzaff): It would be more clear to have a adjacentTrap helper for this.
		return Capture{
			Piece: p.board[b.Square()],
			Src:   b.Square(),
		}
	}
	return Capture{}
}

// generateSteps appends all legal steps to a.
// a legal step is any sliding step, push or
// pull step in which the Src, Dest, or Alt is occupied
// and the Src piece is not frozen, and all captures
// are completed.
func (p *Pos) generateSteps(a *[]*Step) {
	if p.stepsLeft < 4 {
		*a = append(*a, &Step{Pass: true})
	}
	if p.stepsLeft == 0 {
		return
	}
	c := p.Side()
	presence := p.Presence(c)
	enemyPresence := p.Presence(c.Opposite())
	empty := p.Empty()
	for b := presence; b > 0; b &= b - 1 {
		sb := b & -b
		src := b.Square()
		t := p.board[src]

		if p.frozenB(t, sb) {
			continue
		}
		var db Bitboard
		if t.SameType(GRabbit) {
			db = rabbitStepsB[c][src]
		} else {
			db = stepsB[src]
		}
		emptyDB := db & empty

		// Generate default step from src to dest with possible capture.
		for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			*a = append(*a, &Step{
				Piece1: t,
				Src:    src,
				Dest:   dest,
				Alt:    invalidSquare,
				Cap:    p.capture(presence, src, dest),
			})
		}
		// Pushing and pulling is not possible.
		if p.stepsLeft < 2 || t.SameType(GRabbit) {
			continue
		}
		// Generate pushes from src to dest (to alt) with possible capture.
		for b2 := db & enemyPresence & p.Weaker(t); b2 > 0; b2 &= b2 - 1 {
			dest := b2.Square()
			for ab := stepsB[dest] & ^sb & empty; ab > 0; ab &= ab - 1 {
				alt := ab.Square()
				step := &Step{
					Piece1: t,
					Piece2: p.At(dest),
					Src:    src,
					Dest:   dest,
					Alt:    alt,
				}
				if cap := p.capture(presence, src, dest); cap.Valid() {
					step.Cap = cap
				} else if cap := p.capture(enemyPresence, dest, alt); cap.Valid() {
					step.Cap = cap
				}
				*a = append(*a, step)
			}
		}
		// Generate pulls from alt to src (to dest) with possible capture.
		for ab := stepsB[src] & enemyPresence & p.Weaker(t); ab > 0; ab &= ab - 1 {
			for b2 := emptyDB; b2 > 0; b2 &= b2 - 1 {
				dest := b2.Square()
				alt := ab.Square()
				step := &Step{
					Piece1: t,
					Piece2: p.At(alt),
					Src:    src,
					Dest:   dest,
					Alt:    alt,
				}
				if cap := p.capture(presence, src, dest); cap.Valid() {
					step.Cap = cap
				} else if cap := p.capture(enemyPresence, alt, src); cap.Valid() {
					step.Cap = cap
				}
				*a = append(*a, step)
			}
		}
	}
}

func (p *Pos) getRootMovesLenInternal(set map[uint64]bool, prefix []Step, moves *[][]Step, stepsLeft int) {
	if stepsLeft == 0 {
		move := make([]Step, len(prefix), len(prefix)+1)
		copy(move, prefix)
		if !prefix[len(prefix)-1].Pass {
			move = append(move, Step{Pass: true})
		}
		if !set[p.zhash] {
			set[p.zhash] = true
			if !Recurring(move) {
				*moves = append(*moves, move)
			}
		}
		return
	}
	assert("movesLeft < 0", stepsLeft > 0)
	n := len(prefix)
	var stepList StepList
	stepList.Generate(p)
	for i := 0; i < stepList.Len(); i++ {
		step := stepList.At(i)
		if step.Len() > stepsLeft {
			continue
		}
		prefix = append(prefix, step)
		if err := p.Step(step); err != nil {
			panic(err)
		}
		if step.Pass {
			if !set[p.zhash] {
				set[p.zhash] = true
				move := make([]Step, len(prefix))
				copy(move, prefix)
				if !Recurring(move) {
					*moves = append(*moves, move)
				}
			}
		} else {
			p.getRootMovesLenInternal(set, prefix, moves, stepsLeft-step.Len())
		}
		if err := p.Unstep(); err != nil {
			panic(err)
		}
		prefix = prefix[:n]
	}
}

func (e *Engine) getRootMovesLen(p *Pos, depth int) [][]Step {
	if depth <= 0 || depth > 4 {
		panic("depth <= 0 || depth > 4")
	}
	if depth > p.stepsLeft {
		depth = p.stepsLeft
	}
	var moves [][]Step
	var prefix []Step
	set := map[uint64]bool{p.zhash: true}
	for i := 1; i <= depth; i++ {
		p.getRootMovesLenInternal(set, prefix, &moves, i)
	}
	return moves
}

func recurring2(ka, kb StepKind, a, b Step) bool {
	// Three cases:
	// 	a and b are default steps that cancel eachother out
	//	a and b are push & pull
	//	a and b are pull & push
	return (ka == KindDefault && kb == KindDefault ||
		ka == KindPush && kb == KindPull ||
		ka == KindPull && kb == KindPush) &&
		a.Src == b.Dest && a.Dest == b.Src
}

func recurring4(a, b, c, d Step) bool {
	ka, kb, kc, kd := a.Kind(), b.Kind(), c.Kind(), d.Kind()
	return recurring2(ka, kb, a, b) || recurring2(kb, kc, b, c) ||
		recurring2(ka, kc, a, c) || recurring2(kb, kd, b, d) ||
		recurring2(ka, kd, a, d) || recurring2(kb, kc, b, c)
}

// Recurring evaluates statically whether the move leads to a recurring position
// Or whether a move contains intermediate redundancies like a unnecessary pivot
// or push pull pairs. The provided move should be terminated with a (pass) move.
func Recurring(move []Step) bool {
	switch len(move) {
	case 1: // (pass)
		return true
	case 3: // default (pass)
		return recurring2(move[0].Kind(), move[1].Kind(), move[0], move[1])
	case 5: // step step (pass)
		return recurring4(move[0], move[1], move[2], move[3])
	default: // 1, 3, other lengths
		return false
	}
}
