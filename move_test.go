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
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n Da5e",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("a5"),
			Dest:   ParseSquare("b5"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n Ra3n",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("a3"),
			Dest:   ParseSquare("a4"),
			Piece1: GRabbit,
			Alt:    invalidSquare,
		}, {
			Pass: true,
		}},
	}, {
		input: "Da4n ra3n",
		want: []Step{{
			Src:    ParseSquare("a4"),
			Dest:   ParseSquare("a5"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("a3"),
			Dest:   ParseSquare("a4"),
			Piece1: SRabbit,
			Alt:    invalidSquare,
		}, {
			Pass: true,
		}},
	}, {
		input: "Dh3s Rh2n Rg1e Rf1e",
		want: []Step{{
			Src:    ParseSquare("h3"),
			Dest:   ParseSquare("h2"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("h2"),
			Dest:   ParseSquare("h3"),
			Piece1: GRabbit,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("g1"),
			Dest:   ParseSquare("h1"),
			Piece1: GRabbit,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("f1"),
			Dest:   ParseSquare("g1"),
			Piece1: GRabbit,
			Alt:    invalidSquare,
		}, {
			Pass: true,
		}},
	}, {
		input: "Md2n Dh2n Md3n Md4s",
		want: []Step{{
			Src:    ParseSquare("d2"),
			Dest:   ParseSquare("d3"),
			Piece1: GCamel,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("h2"),
			Dest:   ParseSquare("h3"),
			Piece1: GDog,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("d3"),
			Dest:   ParseSquare("d4"),
			Piece1: GCamel,
			Alt:    invalidSquare,
		}, {
			Src:    ParseSquare("d4"),
			Dest:   ParseSquare("d3"),
			Piece1: GCamel,
			Alt:    invalidSquare,
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
