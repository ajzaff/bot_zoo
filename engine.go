package zoo

type Engine struct {
	p *Pos
}

func NewEngine() *Engine {
	return &Engine{
		p: new(Pos),
	}
}

func (e *Engine) Pos() *Pos {
	return e.p
}

func (e *Engine) SetPos(p *Pos) {
	*e.p = *p
}
