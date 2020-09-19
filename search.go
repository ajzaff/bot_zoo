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

func (e *Engine) searchRoot(ponder bool) {
	p := e.Pos()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	defer e.Stop()

	if e.UseTranspositionTable {
		e.tt.NewSearch()
	}

	var bestMove Move

	for p.stepsLeft > 0 {
		var (
			bestStep  Step
			bestValue = -Inf
			stepList  StepList
		)
		stepList.Generate(p)
		for j := 0; j < stepList.Len(); j++ {
			step := stepList.At(j)

			var stepValue Value
			for k := 0; k < trials; k++ {
				if err := p.Step(step.Step); err != nil {
					ppanic(p, err)
				}
				value := e.search(p, r, &stepList)
				if err := p.Unstep(); err != nil {
					ppanic(p, err)
				}
				stepValue += value
			}

			if stepValue > bestValue {
				bestValue = stepValue
				bestStep = step.Step
			}
			stepList.SetValue(j, stepValue/trials)
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
		if err := p.Step(bestStep); err != nil {
			ppanic(p, err)
		}
		defer func() {
			if err := p.Unstep(); err != nil {
				ppanic(p, err)
			}
		}()
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

	m := Value(1) // side multiplier

	for {
		// Is this a terminal node? Return the value immediately.
		// eval := 0 //p.Value()
		// if eval.Terminal() {
		// 	return m * eval
		// }

		// Generate the steps for the next node.
		ls = append(ls, steps.Len())
		steps.Generate(p)

		// If immobilized return a losing score.
		numSteps := steps.Len() - ls[len(ls)-1]
		if numSteps <= 0 {
			return m * -Win
		}

		// Choose a next step at random.
		i := ls[len(ls)-1] + r.Intn(numSteps)
		step := steps.At(i)

		initSide := p.Side()

		// Make the step now.
		if err := p.Step(step.Step); err != nil {
			ppanic(p, err)
		}

		if p.Side() != initSide {
			m = -m
		}

		// Defer the unstep as well as backpropagation.
		defer func(i int) {
			if err := p.Unstep(); err != nil {
				ppanic(p, err)
			}
		}(i)
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
