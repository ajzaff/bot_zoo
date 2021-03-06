package zoo

import (
	"testing"
)

type legalTestCase struct {
	name          string
	shortPosition string
	steps         []Step
	moveNum       int
	inputStep     Step
	input         string
	want          bool
}

func runLegalTestCase(t *testing.T, tc legalTestCase) {
	p, err := ParseShortPosition(tc.shortPosition)
	if err != nil {
		t.Fatalf("ParseShortPosition(%q): %v", tc.shortPosition, err)
	}
	if tc.moveNum != 0 {
		p.moveNum = tc.moveNum
		if p.moveNum == 1 {
			p.stepsLeft = 16
		}
	}
	for _, step := range tc.steps {
		if !p.Legal(step) {
			t.Log(p)
			t.Fatalf("Intermediate step is not legal: %s", step)
		}
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
		t.Log(p)
		t.Errorf("Legal(%q): got legal=%v, want legal=%v", tc.input, got, tc.want)
	}
}

func TestIllegalSteps(t *testing.T) {
	for _, tc := range []legalTestCase{{
		name:          "steps left",
		shortPosition: "g [                      r                               R         ]",
		steps: []Step{
			MakeStep(GRabbit, G2, G3),
			MakeStep(GRabbit, G3, G4),
			MakeStep(GRabbit, G4, F4),
			MakeStep(GRabbit, F4, E4),
		},
		input: "Re4n",
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
		name:          "abandon push",
		shortPosition: "g [                                  r       Cr      D             ]",
		steps: []Step{
			MakeStep(SRabbit, D3, D4),
		},
		input: "rc4w",
	}, {
		name:          "incomplete push",
		shortPosition: "s [            C      De                                           ]",
		steps: []Step{
			MakeStep(GCat, E7, E8),
		},
		input: "ee6e",
	}, {
		name:          "incomplete push frozen",
		shortPosition: "s [          eD       r                                            ]",
		input:         "rd6e",
	}, {
		name:          "incomplete push 3",
		shortPosition: "s [  dchehm   c Rrrd  Rr  R r R DR r       H  D  H   CEMC   RRR    ]",
		steps: []Step{
			MakeStep(GRabbit, F7, F6),
			MakeStep(SElephant, F8, F7),
			MakeStep(GRabbit, D6, C6),
		},
		input: "rb5s",
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
		name:          "push and pull not shared",
		shortPosition: "g [                                  r       Cr      D             ]",
		steps: []Step{
			MakeStep(SRabbit, D3, D4),
			MakeStep(GCat, C3, D3),
		},
		input: "rc4s",
	}, {
		name:          "push and pull not shared 2",
		shortPosition: "s [    rrhem    r   d  c  r Rh R R   R H   rC     rR D  E R  M   CR]",
		steps: []Step{
			MakeStep(GRabbit, C4, C3),
			MakeStep(SHorse, C5, C4),
			MakeStep(GRabbit, B5, C5),
		},
		input: "ma7s",
	}, {
		name:          "new push has stronger unfrozen adjacent piece",
		shortPosition: "g [                           r       D       h                    ]",
		input:         "rd5n",
	}, {
		name:          "turn repetition",
		shortPosition: "g [                                                                ]",
		moveNum:       1,
		steps: []Step{
			MakeSetup(GRabbit, A1),
			MakeSetup(GRabbit, B1),
			MakeSetup(GRabbit, C1),
			MakeSetup(GRabbit, D1),
			MakeSetup(GRabbit, E1),
			MakeSetup(GRabbit, F1),
			MakeSetup(GRabbit, G1),
			MakeSetup(GRabbit, H1),
			MakeSetup(GDog, A2),
			MakeSetup(GHorse, B2),
			MakeSetup(GCat, C2),
			MakeSetup(GCat, D2),
			MakeSetup(GElephant, E2),
			MakeSetup(GCamel, F2),
			MakeSetup(GHorse, G2),
			MakeSetup(GDog, H2),
			MakeSetup(SRabbit, A8),
			MakeSetup(SRabbit, B8),
			MakeSetup(SRabbit, C8),
			MakeSetup(SRabbit, D8),
			MakeSetup(SRabbit, E8),
			MakeSetup(SRabbit, F8),
			MakeSetup(SRabbit, G8),
			MakeSetup(SRabbit, H8),
			MakeSetup(SDog, A7),
			MakeSetup(SHorse, B7),
			MakeSetup(SCat, C7),
			MakeSetup(SCat, D7),
			MakeSetup(SElephant, E7),
			MakeSetup(SCamel, F7),
			MakeSetup(SHorse, G7),
			MakeSetup(SDog, H7),
			MakeStep(GDog, A2, A3),
			MakeStep(GHorse, G2, G3),
			MakeStep(GHorse, G3, G2),
		},
		input: "Da3s",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			runLegalTestCase(t, tc)
		})
	}
}

func TestLegalSteps(t *testing.T) {
	for _, tc := range []legalTestCase{{
		name:          "push on third step",
		shortPosition: "g [       r                            Ed     Cr           R       ]",
		steps: []Step{
			MakeStep(GCat, D3, D2),
			MakeStep(SRabbit, E3, D3),
		},
		input: "df4s",
		want:  true,
	}, {
		name:          "pull and push same piece",
		shortPosition: "g [       r                            Ed     Cr           R       ]",
		steps: []Step{
			MakeStep(GCat, D3, D2),
			MakeStep(SRabbit, E3, D3),
		},
		input: "rd3e",
		want:  true,
	}, {
		name:          "pull and push same piece 2",
		shortPosition: "g [       r                            Ed     Cr           R       ]",
		steps: []Step{
			MakeStep(GCat, D3, D2),
			MakeStep(SRabbit, E3, D3),
			MakeStep(SRabbit, D3, E3),
		},
		input: "Ra1n",
	}, {
		name:          "completed push",
		shortPosition: "g [          cD       r                                            ]",
		steps: []Step{
			MakeStep(SRabbit, D6, E6),
		},
		input: "Dd7s",
		want:  true,
	}, {
		name:          "pull to capture",
		shortPosition: "g [                                  rr      C       D             ]",
		steps: []Step{
			MakeStep(GCat, C3, D3),
		},
		input: "rc4s",
		want:  true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			runLegalTestCase(t, tc)
		})
	}
}

type canPassTestCase struct {
	name          string
	shortPosition string
	steps         []Step
	moveNum       int
	want          bool
}

func runCanPassTestCase(t *testing.T, tc canPassTestCase) {
	p, err := ParseShortPosition(tc.shortPosition)
	if err != nil {
		t.Fatalf("ParseShortPosition(%q): %v", tc.shortPosition, err)
	}
	if tc.moveNum != 0 {
		p.moveNum = tc.moveNum
		if p.moveNum == 1 {
			p.stepsLeft = 16
		}
	}
	for _, step := range tc.steps {
		p.Step(step)
	}
	if got := p.CanPass(); got != tc.want {
		t.Errorf("CanPass(): got can_pass=%v, want can_pass=%v", got, tc.want)
	}
}

func TestCanPass(t *testing.T) {
	for _, tc := range []canPassTestCase{{
		name:          "no steps taken",
		shortPosition: "g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]",
	}, {
		name:          "pass setup",
		shortPosition: "g [                                                                ]",
		moveNum:       1,
		steps: []Step{
			MakeSetup(GRabbit, A1),
			MakeSetup(GRabbit, B1),
			MakeSetup(GRabbit, C1),
			MakeSetup(GRabbit, D1),
			MakeSetup(GRabbit, E1),
			MakeSetup(GRabbit, F1),
			MakeSetup(GRabbit, G1),
			MakeSetup(GRabbit, H1),
			MakeSetup(GDog, A2),
			MakeSetup(GHorse, B2),
			MakeSetup(GCat, C2),
			MakeSetup(GCat, D2),
			MakeSetup(GElephant, E2),
			MakeSetup(GCamel, F2),
			MakeSetup(GHorse, G2),
		},
	}, {
		name:          "incomplete push",
		shortPosition: "s [          eD                                                    ]",
		steps: []Step{
			MakeStep(GDog, D7, D6),
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			runCanPassTestCase(t, tc)
		})
	}
}
