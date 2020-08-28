package zoo

type Engine struct {
	p *Pos
}

func NewEngine() *Engine {
	pos, _ := ParseShortPosition(PosEmpty)
	return &Engine{
		p: pos,
	}
}

func (e *Engine) Pos() *Pos {
	return e.p
}

func (e *Engine) SetPos(p *Pos) {
	*e.p = *p
}
