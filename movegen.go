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

func (p *Pos) pieceSteps(piece Piece, table [64]Bitboard) []Step {
	var steps []Step
	piece = piece.MakeColor(p.side)
	bs := p.bitboards[piece]
	for bs > 0 {
		b := bs & -bs
		bs &= ^b
		if p.frozenB(b) {
			continue
		}
		src := b.Square()
		for ds := stepsB[src]; ds > 0; {
			d := ds & -ds
			ds &= ^d
			if p.bitboards[Empty]&d == 0 {
				continue
			}
			dest := d.Square()
			steps = append(steps, p.completeCapture(Step{
				Src:    src,
				Dest:   dest,
				Piece1: piece,
			}))
		}
	}
	return steps
}

func (p *Pos) completeCapture(step Step) Step {
	if step.Capture() {
		return step
	}
	if err := p.Step(step); err != nil {
		panic(err)
	}
	defer func() {
		if err := p.Unstep(step); err != nil {
			panic(err)
		}
	}()
	c := step.Piece1.Color()
	ps := p.presence[c]
	srcB := step.Src.Bitboard()

	if b := srcB.Neighbors() & Traps & ^ps.Neighbors(); b != 0 {
		step.Cap = Capture{
			Piece: p.atB(b),
			Src:   b.Square(),
		}
	}
	return step
}

func (p *Pos) getPulls(steps *[]Step) {
	s2 := p.side.Opposite()

	if len(p.steps) == 0 {
		return
	}
	i := len(p.steps) - 1
	for ; i >= 0 && p.steps[i].Capture(); i-- {
	}
	assert("i < 0", i >= 0)
	t1 := p.steps[i].Piece1
	if t1.SamePiece(GRabbit) {
		return
	}
	dest := p.steps[i].Src
	db := dest.Bitboard()

	if p.presence[s2].Neighbors()&db == 0 {
		return
	}

	for t2 := GRabbit.MakeColor(s2); t2 < t1; t2++ {
		t2b := p.bitboards[t2]
		if t2b == 0 || t2b.Neighbors()&db == 0 {
			continue
		}
		for t2b > 0 {
			b := t2b & -t2b
			t2b &= ^b
			src := b.Square()
			assert("src == dest", src != dest)
			*steps = append(*steps, p.completeCapture(Step{
				Src:    src,
				Dest:   dest,
				Piece1: t2,
			}))
		}
	}
}

func (p *Pos) getPushes(steps *[]Step) {
	s1, s2 := p.side, p.side.Opposite()
	s2n := p.presence[s2].Neighbors()
	for p1 := GCat.MakeColor(s1); p1 <= GElephant.MakeColor(s1); p1++ {
		p1b := p.bitboards[p1]
		if p1b == 0 || p1b&s2n == 0 {
			continue
		}
		p1n := p1b.Neighbors()
		for p2 := GRabbit.MakeColor(s2); p2 < p1.MakeColor(s2); p2++ {
			p2b := p.bitboards[p2]
			if p2b == 0 {
				continue
			}
			for bs := p1n & p2b; bs > 0; {
				b := bs & -bs
				bs &= ^b
				if p.frozenB(b) {
					continue
				}
				// src := b.Square()
				// for ds := stepsFor(p2, src); ds > 0; {
				// 	d := ds & -ds
				// 	ds &= ^d
				// 	if p.Bitboards[Empty]&d == 0 {
				// 		continue
				// 	}
				// 	dest := d.Square()
				// 	assert("src == dest", src != dest)
				// 	*steps = append(*steps, Step{
				// 		Src:   src,
				// 		Dest:  dest,
				// 		Piece: p2,
				// 		Dir:   NewDelta(src.Delta(dest)),
				// 	})
				// }
			}
		}
	}
}

func (p *Pos) completePush(steps *[]Step) {
	i := len(p.steps) - 1
	for ; i >= 0 && p.steps[i].Capture(); i-- {
	}
	assert("i < 0", i >= 0)
	push := p.steps[i]
	p2 := push.Piece1
	dest := push.Src
	destB := dest.Bitboard()

	for p1 := p2.MakeColor(p.side) + 1; p1 < GElephant.MakeColor(p.side); p1++ {
		p1b := p.bitboards[p1]
		if p1b == 0 || p1b.Neighbors()&destB == 0 {
			continue
		}
		for p1b > 0 {
			b := p1b & -p1b
			p1b &= ^b
			src := b.Square()
			assert("src == dest", src != dest)
			*steps = append(*steps, Step{
				Src:    src,
				Dest:   dest,
				Piece1: p2,
			})
		}
	}
}

// GenSteps generates steps including sliding moves, pushes and pulls.
// Captures are completed internally to Step, but you may check whether
// a step results in a capture by calling captures with the step as an
// argument.
func (p *Pos) GenSteps() []Step {
	var steps []Step
	if len(p.steps) > 0 && p.steps[len(p.steps)-1].Piece1.Color() != p.side {
		p.completePush(&steps)
	} else {
		if len(p.steps) < 3 {
			p.getPushes(&steps)
		}
		steps = append(steps, p.pieceSteps(GRabbit, rabbitStepsB[p.side])...)
		for t := GCat; t <= GElephant; t++ {
			steps = append(steps, p.pieceSteps(t, stepsB)...)
		}
	}
	return steps
}

func (p *Pos) getRootMoves(prefix []Step, moves *[][]Step, depth int) {
	if depth <= 0 {
		*moves = append(*moves, prefix)
		return
	}
	steps := p.GenSteps()
	for _, step := range steps {
		move := append(prefix, step)
		if err := p.Step(step); err != nil {
			panic(err)
		}
		newPrefix := make([]Step, len(move))
		copy(newPrefix, move)
		p.getRootMoves(newPrefix, moves, depth-1)
		p.Unstep(step)
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
