package zoo

import (
	"fmt"
	"math/rand"
	"time"
)

const trials = 2000

func (e *Engine) searchRoot() ExtStep {
	p := e.Pos()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	defer e.Stop()

	if p.moveNum == 1 {
		move := e.randomSetup(r)
		if !e.ponder {
			e.writef("bestmove %s\n", MoveString(move))
		}
		return ExtStep{}
	}

	var best []Step

	for p.stepsLeft > 0 {
		var (
			bestStep  Step
			bestValue = -Inf
			stepList  StepList
		)
		stepList.Generate(p)
		for j := 0; j < stepList.Len(); j++ {
			step := stepList.At(j)

			if step.Len() > p.stepsLeft {
				step.Value = -1
				continue
			}

			var stepValue Value
			for k := 0; k < trials; k++ {
				if err := p.Step(step.Step); err != nil {
					ppanic(p, err)
				}
				value := e.search(p, r, &stepList, uint8(step.Step.Len()))
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
		if bestStep.Pass() {
			break
		}

		best = append(best, bestStep)
		if err := p.Step(bestStep); err != nil {
			ppanic(p, err)
		}
		defer func() {
			if err := p.Unstep(); err != nil {
				ppanic(p, err)
			}
		}()
	}
	best = append(best, Pass)
	if !e.ponder && MoveLen(best) > 0 {
		e.writef("bestmove %s\n", MoveString(best))
	}
	return ExtStep{}
}

// search performs a Monte Carlo Tree Search from the position p.
// search always plays until the end of the game with the result
// being a Win for one side or the other. This value is a single
// stochastic outcome and must be repeated many times to hone in
// on a true result.
func (e *Engine) search(p *Pos, r *rand.Rand, steps *StepList, depth uint8) Value {
	// Maintain a stack of move lengths which will allow us to backprop the search value up.
	ls := []int{steps.Len()}
	defer steps.Truncate(ls[0])

	m := Value(1) // side multiplier

	for {
		// Is this a terminal node? Return the value immediately.
		eval := p.Value()
		if eval.Terminal() {
			return m * eval
		}

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
