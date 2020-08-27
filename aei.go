package zoo

type AEI struct {
	e *Engine
}

func NewAEI(e *Engine) *AEI {
	return &AEI{e}
}
