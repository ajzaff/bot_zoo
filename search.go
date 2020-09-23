package zoo

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const trials = 2000

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
			bestStep  Step
			bestValue = -Inf
			stepList  StepList
		)
		stepList.Generate(p)
		for j := 0; j < stepList.Len(); j++ {
			step := stepList.At(j)

			if ok, _ := p.Legal(step.Step); !ok {
				continue
			}

			initSide := p.Side()

			p.Step(step.Step)

			var (
				m     Value = 1
				value Value
			)

			if p.Side() != initSide {
				m = -1
			}

			for k := 0; k < trials; k++ {
				value += m * e.search(p, r, &stepList)
			}
			value /= trials

			stepList.SetValue(j, value)
			if value > bestValue {
				bestValue = value
				bestStep = step.Step
			}

			p.Unstep()
		}

		stepList.Sort(0)
		for i := 0; i < stepList.Len(); i++ {
			fmt.Println("log ", stepList.At(i).Value, stepList.At(i).Step)
		}
		fmt.Println("log ---")
		stepList.Truncate(0)

		if bestValue == -Inf {
			break
		}

		bestMove = append(bestMove, bestStep)
		p.Step(bestStep)
		defer func() { p.Unstep() }()
	}
	if !ponder {
		e.Outputf("bestmove %s", bestMove.WithCaptureContext(p).String())
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
			if ok, _ := p.Legal(step.Step); ok {
				steps.Swap(i, j)
				j++
			}
		}
		steps.Truncate(j)

		// Immobilized? Return a terminal loss.
		numSteps := steps.Len() - l
		if numSteps <= 0 {
			return m * Loss
		}

		initSide := p.Side()

		// Choose a next step at random.
		i := ls[len(ls)-1] + r.Intn(numSteps)
		step := steps.At(i)

		// Make the step now.
		p.Step(step.Step)

		if p.Side() != initSide {
			m = -m
		}

		// Defer the unstep and backpropagation.
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
