package zoo

func (e *Engine) RandomSetup() []Step {
	p := e.Pos()
	c := p.side
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
				Alt:    at,
				Piece1: ps[0],
			})
			ps = ps[1:]
		}
	}
	return setup
}
