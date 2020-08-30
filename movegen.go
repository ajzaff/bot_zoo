package zoo

func (p *Pos) GetSteps(check bool) []Step {
	var res []Step
	for t := GRabbit; t <= SElephant; t++ {
		bs := p.Bitboards[t]
		for bs > 0 {
			b := bs & -bs
			bs &= ^b
			src := b.Square()
			ds := StepsFor(t, src)
			for ds > 0 {
				d := ds & -ds
				ds &= ^d
				dest := d.Square()
				step := Step{
					Src:   src,
					Dest:  dest,
					Piece: t,
					Dir:   NewDelta(src.Delta(dest)),
				}
				if p.Bitboards[Empty]&d == 0 {
					continue
				}
				if check {
					if ok, _ := p.CheckStep(step); !ok {
						continue
					}
				}
				res = append(res, step)
			}
		}
	}
	return res
}

func (p *Pos) getMoves(transpose map[int64]bool, prefix []Step, moves *[][]Step, depth int) {
	if depth <= 0 {
		if !transpose[p.ZHash] && !p.Push {
			transpose[p.ZHash] = true
			*moves = append(*moves, prefix)
		}
		return
	}
	for _, step := range p.GetSteps(true) {
		move := append(prefix, step)
		t, cap, _ := p.Step(step)
		if cap.Capture() {
			move = append(move, cap)
		}
		newPrefix := make([]Step, len(move))
		copy(newPrefix, move)
		t.getMoves(transpose, newPrefix, moves, depth-1)
	}
}

func (e *Engine) getMovesLen(p *Pos, n int) [][]Step {
	transpose := map[int64]bool{p.ZHash: true}
	if n <= 0 || n > 4 {
		panic("n <= 0 || n > 4")
	}
	var moves [][]Step
	for i := 1; i <= n; i++ {
		p.getMoves(transpose, nil, &moves, i)
	}
	return moves
}

func (e *Engine) RandomSetup() []Step {
	p := e.Pos()
	c := p.Side
	rank := 7
	if c == Gold {
		rank = 1
	}
	ps := []Piece{
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GRabbit.MakeColor(c),
		GCat.MakeColor(c),
		GCat.MakeColor(c),
		GDog.MakeColor(c),
		GDog.MakeColor(c),
		GHorse.MakeColor(c),
		GHorse.MakeColor(c),
		GCamel.MakeColor(c),
		GElephant.MakeColor(c),
	}
	e.r.Shuffle(len(ps), func(i, j int) {
		ps[i], ps[j] = ps[j], ps[i]
	})
	var setup []Step
	for i := rank; i >= rank-1; i-- {
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			setup = append(setup, Step{
				Src:   invalidSquare,
				Dest:  at,
				Piece: ps[0],
			})
			ps = ps[1:]
		}
	}
	return setup
}
