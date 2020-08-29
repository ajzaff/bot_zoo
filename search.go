package zoo

func (e *Engine) Search() (move []Step, score int) {
	p := e.Pos()
	if p.MoveNum == 1 {
		// TODO(ajzaff): Find best setup moves using a specialized search.
		// For now, choose a random setup.
		return e.RandomSetup(), 0
	}
	for d := 0; d <= 4; d++ {
		move, score = e.search(p, d)
	}
	return move, score
}

func (e *Engine) search(p *Pos, depth int) (move []Step, score int) {
	scoreFn := func(x int) int {
		if p.Side != Gold {
			return -x
		}
		return x
	}
	if depth == 0 || p.Terminal() {
		return nil, scoreFn(p.Score())
	}
	minmaxFn := func(x, y int) bool {
		if p.Side != Gold {
			return x > y
		}
		return x < y
	}
	var best []Step
	value := scoreFn(-terminalEval)
	steps := p.GetSteps(true)
	for _, step := range steps {
		t, _, err := p.Step(step)
		if err != nil {
			continue
		}
		move, v := e.search(t, depth-1)
		if minmaxFn(value, v) {
			best = append([]Step{step}, move...)
			value = v
		}
	}
	return best, value
}
