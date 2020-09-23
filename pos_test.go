package zoo

import "testing"

func TestLegal(t *testing.T) {
	for _, tc := range []struct {
		name          string
		shortPosition string
		steps         []Step
		moveNum       int
		inputStep     Step
		input         string
		want          bool
	}{{
		name:          "steps left",
		shortPosition: "g [                      r                               R         ]",
		steps: []Step{
			MakeStep(GRabbit, B2, D2),
			MakeStep(GRabbit, D2, E2),
			MakeStep(GRabbit, E2, F2),
			MakeStep(GRabbit, F2, G2),
		},
		input: "Rg2n",
	}, {
		name:          "valid piece",
		shortPosition: "g [                                                                ]",
		input:         "Dd2s",
	}, {
		name:          "setup after first move",
		shortPosition: "g [                      r                               R         ]",
		input:         "Dd1",
	}, {
		name:          "setup piece count",
		shortPosition: "g [                                                              DD]",
		moveNum:       1,
		input:         "Dc1",
	}, {
		name:          "setup square",
		shortPosition: "g [                                                              DD]",
		moveNum:       1,
		input:         "Dc3",
	}, {
		name:          "valid src",
		shortPosition: "s [                      r                               R         ]",
		inputStep:     MakeStep(SRabbit, 64, G5),
	}, {
		name:          "empty dest",
		shortPosition: "s [                      r       r                       R         ]",
		input:         "Rc6s",
	}, {
		name:          "frozen src",
		shortPosition: "s [                      r       rM                      R         ]",
		input:         "Rg5w",
	}, {
		name:          "capture validation",
		shortPosition: "s [                  r         R                                   ]",
		input:         "Rc6x",
	}, {
		name:          "abandon push",
		shortPosition: "s [            C      De                                           ]",
		steps: []Step{
			MakeStep(GCat, E7, E8),
		},
		input: "Dd7w",
	}, {
		name:          "incomplete push",
		shortPosition: "s [            C      De                                           ]",
		steps: []Step{
			MakeStep(GCat, E7, E8),
		},
		input: "ee6e",
	}, {
		name:          "too weak push",
		shortPosition: "g [                                                   rD      R    ]",
		steps: []Step{
			MakeStep(SRabbit, D2, D3),
		},
		input: "Rd1n",
	}, {
		name:          "begin push on last step",
		shortPosition: "s [       r                 h       D                              ]",
		steps: []Step{
			MakeStep(SRabbit, H8, H7),
			MakeStep(SRabbit, H7, H6),
			MakeStep(SRabbit, H6, H5),
		},
		input: "Db4e",
	}, {
		name:          "too weak pull",
		shortPosition: "s [                      rR                                        ]",
		steps: []Step{
			MakeStep(SRabbit, G6, G5),
		},
		input: "Rh6w",
	}, {
		name:          "new push has stronger unfrozen adjacent piece 1",
		shortPosition: "g [                           r       D       h                    ]",
		input:         "rd5n",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParseShortPosition(tc.shortPosition)
			if err != nil {
				t.Fatalf("ParseShortPosition(%q): %v", tc.shortPosition, err)
			}
			if tc.moveNum != 0 {
				p.moveNum = tc.moveNum
			}
			for _, step := range tc.steps {
				p.Step(step)
			}
			s := tc.inputStep
			if s == 0 {
				s, err = ParseStep(tc.input)
				if err != nil {
					t.Fatalf("ParseStep(%q): %v", tc.input, err)
				}
			}
			if got := p.Legal(s); got != tc.want {
				t.Errorf("Legal(%q): got legal=%v, want legal=%v", tc.input, got, tc.want)
			}
		})
	}
}
