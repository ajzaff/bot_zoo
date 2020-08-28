package zoo

func (p *Pos) GetSteps() []Step {
	if p.Push {
		return nil
	}
	var res []Step
	for t := GRabbit.MakeColor(p.Side); t <= GElephant.MakeColor(p.Side); t++ {
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
				if p.Bitboards[Empty]&d == 0 {
					continue
				}
				res = append(res, Step{
					Src:   src,
					Dest:  dest,
					Piece: t,
					Dir:   NewDelta(src.Delta(dest)),
				})
			}
		}
	}
	return res
}
