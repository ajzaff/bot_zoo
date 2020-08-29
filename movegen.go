package zoo

import (
	"math/rand"
)

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

var r = rand.New(rand.NewSource(1337))

func (e *Engine) BestMove() []Step {
	p := e.Pos()
	if p.MoveNum == 1 {
		return p.RandomSetup()
	}
	for i := 0; i < 6; i++ { // try 6 times
		var move []Step
		for j := 0; j < 4; j++ {
			steps := p.GetSteps(true)
			if len(steps) == 0 {
				return move
			}
			scores := e.Sort(steps)
			bestScore := scores[0]
			n := 1
			for ; n < len(steps); n++ {
				if scores[n] != bestScore {
					break
				}
			}
			step := steps[r.Intn(n)]
			move = append(move, step)
			var cap Step
			p, cap, _ = p.Step(step)
			if cap.Capture() {
				move = append(move, cap)
			}
		}
		if _, _, err := p.Move(move, false); err == nil {
			return move
		}
	}
	return nil
}

func (p *Pos) RandomSetup() []Step {
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
	r.Shuffle(len(ps), func(i, j int) {
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
