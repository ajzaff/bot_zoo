package zoo

// Stack is a structure maintained for each ply of search.
// It contains all the information relevant to that ply.
// The stack is used for reconstructing the searched PV.
type Stack struct {
	PV    []Step
	Depth int
	Step  Step
	Eval  int
}

// StepList implements an efficient data structure for storing steps from search lines.
type StepList struct {
	steps []Step
	p     int
}

// Generate the moves for position p and append them to the move list.
func (l *StepList) Generate(p *Pos) {
	p.Steps(&l.steps)
}

// Truncate truncates the list to the given length n.
func (l *StepList) Truncate(n int) {
	l.steps = l.steps[:n]
}

func (l *StepList) At(i int) Step {
	return l.steps[i]
}

func (l *StepList) Len() int {
	return len(l.steps)
}
