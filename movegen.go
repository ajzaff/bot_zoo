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
