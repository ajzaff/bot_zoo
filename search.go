package zoo

import (
	"math/rand"
	"sync"
	"time"
)

const trials = 2000

const maxPly = 1024

type searchState struct {
	tt TranspositionTable

	wg         sync.WaitGroup
	resultChan chan Move

	// semi-atomic
	stopping int32
	running  int32
}

func (s *searchState) Reset() {
	s.tt.Resize(50)
	s.wg = sync.WaitGroup{}
	s.resultChan = make(chan Move)
	s.stopping = 0
	s.running = 0
}

func (e *Engine) searchRoot(ponder bool) {
	p := e.Pos
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	defer e.Stop()

	if e.UseTranspositionTable {
		e.tt.NewSearch()
	}

	var bestMove Move

	n := p.stepsLeft
	for i := 0; i < n; i++ {
		var (
			best     ExtStep
			stepList StepList
		)
		best.Value = -Inf
		stepList.Generate(p)
		for j := 0; j < stepList.Len(); j++ {

			var (
				step ExtStep
				m    = Value(1)
			)

			step = stepList.At(j)

			if !p.Legal(step.Step) {
				continue
			}

			initSide := p.Side()

			p.Step(step.Step)

			if p.Side() != initSide {
				m = -1
			}

			step.Value = 0
			for k := 0; k < trials; k++ {
				step.Value += e.search(p, r, &stepList)
			}
			step.Value /= trials
			step.Value *= m

			stepList.SetValue(j, step.Value)
			if step.Value > best.Value {
				best = step
			}

			p.Unstep()
		}

		stepList.Sort(0)
		for i := 0; i < stepList.Len(); i++ {
			step := stepList.At(i)
			if cap := step.DebugCaptureContext(p); cap != 0 {
				e.Debugf("[%f] %s %s", stepList.At(i).Value, step.Step, cap)
			} else {
				e.Debugf("[%f] %s", stepList.At(i).Value, step.Step)
			}
		}
		e.Debugf("---")

		// No available moves.
		if best.Value == -Inf {
			break
		}

		cap := p.Step(best.Step)
		bestMove = append(bestMove, best.Step)
		if cap != 0 {
			bestMove = append(bestMove, cap)
		}

		defer func() { p.Unstep() }()

		// Win achieved.
		if p.Terminal().Win() && p.CanPass() {
			break
		}
	}
	if !ponder {
		e.Outputf("bestmove %s", bestMove.String())
	}
}

// search performs a Monte Carlo Tree Search from the position p.
// search always plays until the end of the game with the result
// being a Win for one side or the other. This value is a single
// stochastic outcome and must be repeated many times to hone in
// on a true result.
func (e *Engine) search(p *Pos, r *rand.Rand, steps *StepList) Value {
	// Maintain a stack of move lengths which will allow us to backprop the search value up.
	ls := []int{steps.Len()}
	defer steps.Truncate(ls[0])

	for m := Value(1); ; {
		// Is this a terminal node? Return the value immediately.
		if v := p.Terminal(); v != 0 {
			return m * v
		}

		// Generate the steps for the next node.
		l := steps.Len()
		ls = append(ls, l)
		steps.Generate(p)

		// Test and truncate illegal steps:
		j := l
		for i := j; i < steps.Len(); i++ {
			step := steps.At(i)
			if p.Legal(step.Step) {
				steps.Swap(i, j)
				j++
			}
		}
		steps.Truncate(j)

		numSteps := j - l

		// Immobilized? Return a terminal loss.
		if numSteps <= 0 {
			return m * Loss
		}

		// Choose a next step at random.
		i := l + r.Intn(numSteps)

		step := steps.At(i)

		initSide := p.Side()

		// Make the step now.
		p.Step(step.Step)

		if p.Side() != initSide {
			m = -m
		}

		defer func() { p.Unstep() }()
	}
}

func searchRateKNps(nodes int, start time.Time) int64 {
	return int64(float64(nodes) / (float64(time.Now().Sub(start)) / float64(time.Second)) / 1000)
}

func printSearchInfo(e *Engine, nodes int, depth uint8, start time.Time, best ExtStep) {
	e.Outputf("info depth %d", depth)
	e.Outputf("info time %d", int(time.Now().Sub(start).Seconds()))
	e.Outputf("info score %f", best.Value)
	e.Outputf("info nodes %d", nodes)
	e.Logf("rate %d kN/s", searchRateKNps(nodes, start))
	e.Logf("hashfull %d", e.tt.Hashfull())
}
