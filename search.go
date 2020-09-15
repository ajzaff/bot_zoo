package zoo

const maxPly uint8 = 255

type searchResult struct {
	Depth uint8
	Value Value
}

// Stack is a structure maintained for each ply of search.
// It contains all the information relevant to that ply.
// The stack is used for reconstructing the searched PV.
type Stack struct {
	Depth uint8
	Step  Step
	Eval  Value
	Nodes int
}

func (e *Engine) searchRoot() searchResult {
	var best searchResult

	var stepList StepList
	stack := make([]Stack, 0)
	p := e.Pos()

	for d := uint8(0); d < maxPly; d++ {
		s := Stack{
			Depth: d + 1,
		}
		e.search(&s, p, &stepList)
		stack = append(stack, s)
	}

	return best
}

func (e *Engine) search(stack *Stack, p *Pos, steps *StepList) {
	l := steps.Len()
	steps.Generate(p)
	defer steps.Truncate(l)
}
