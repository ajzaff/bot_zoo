package zoo

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseMove(t *testing.T) {
	for _, tc := range []struct {
		input   string
		want    []Step
		wantErr bool
	}{{
		input: "Da4n",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n Da5e",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Src:    ParseSquare("a5"),
			Dest:   ParseSquare("b5"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n Ra3n",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Src:    ParseSquare("a3"),
			Dest:   ParseSquare("a4"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n ra3n",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Alt:    ParseSquare("a3"),
			Piece1: GDog,
			Piece2: SRabbit,
		}, {
			Pass: true,
		}},
	}, {
		input: "Dh3s Rh2n Rg1e Rf1e",
		want: []Step{{
			Src:    ParseSquare("h3"),
			Dest:   ParseSquare("h2"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Src:    ParseSquare("h2"),
			Dest:   ParseSquare("h3"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Src:    ParseSquare("g1"),
			Dest:   ParseSquare("h1"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Src:    ParseSquare("f1"),
			Dest:   ParseSquare("g1"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Pass: true,
		}},
	}, {
		input: "Md2n Dh2n Md3n Md4s",
		want: []Step{{
			Src:    ParseSquare("d2"),
			Dest:   ParseSquare("d3"),
			Alt:    invalidSquare,
			Piece1: GCamel,
		}, {
			Src:    ParseSquare("h2"),
			Dest:   ParseSquare("h3"),
			Alt:    invalidSquare,
			Piece1: GDog,
		}, {
			Src:    ParseSquare("d3"),
			Dest:   ParseSquare("d4"),
			Alt:    invalidSquare,
			Piece1: GCamel,
		}, {
			Src:    ParseSquare("d4"),
			Dest:   ParseSquare("d3"),
			Alt:    invalidSquare,
			Piece1: GCamel,
		}, {
			Pass: true,
		}},
	}, {
		input: "Hc4s rc5s Hc3w rc4s rc3x",
		want: []Step{{
			Src:    ParseSquare("c4"),
			Dest:   ParseSquare("c3"),
			Alt:    ParseSquare("c5"),
			Piece1: GHorse,
			Piece2: SRabbit,
		}, {
			Src:    ParseSquare("c3"),
			Dest:   ParseSquare("b3"),
			Alt:    ParseSquare("c4"),
			Piece1: GHorse,
			Piece2: SRabbit,
			Cap: Capture{
				Piece: SRabbit,
				Src:   ParseSquare("c3"),
			},
		}, {
			Pass: true,
		}},
	}, {
		input: "Rf1w Rg1w Hg2w Hb3w Cc3x",
		want: []Step{{
			Src:    ParseSquare("f1"),
			Dest:   ParseSquare("e1"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Src:    ParseSquare("g1"),
			Dest:   ParseSquare("f1"),
			Alt:    invalidSquare,
			Piece1: GRabbit,
		}, {
			Src:    ParseSquare("g2"),
			Dest:   ParseSquare("f2"),
			Alt:    invalidSquare,
			Piece1: GHorse,
		}, {
			Src:    ParseSquare("b3"),
			Dest:   ParseSquare("a3"),
			Alt:    invalidSquare,
			Piece1: GHorse,
			Cap: Capture{
				Piece: GCat,
				Src:   ParseSquare("c3"),
			},
		}, {
			Pass: true,
		}},
	}} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseMove(tc.input)
			if tc.wantErr != (err != nil) {
				t.Errorf("ParseMove(): got err = %v, want err = %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseMove() got diff (-want, +got):\n%s", diff)
			}
		})
	}
}
