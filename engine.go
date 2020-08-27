package zoo

type Engine struct {
	p *Pos
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) SetPos(p *Pos) {
	e.p = p
}
