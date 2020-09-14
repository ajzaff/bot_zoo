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
		want: []Step{
			MakeDefault(ParseSquare("a4"), ParseSquare("a5"), GDog),
			Pass,
		},
	}, {
		input: "Da4n Da5e",
		want: []Step{
			MakeDefault(ParseSquare("a4"), ParseSquare("a5"), GDog),
			MakeDefault(ParseSquare("a5"), ParseSquare("b5"), GDog),
			Pass,
		},
	}, {
		input: "Da4n Ra3n",
		want: []Step{
			MakeDefault(ParseSquare("a4"), ParseSquare("a5"), GDog),
			MakeDefault(ParseSquare("a3"), ParseSquare("a4"), GRabbit),
			Pass,
		},
	}, {
		input: "Da4n ra3n",
		want: []Step{
			MakeAlternate(ParseSquare("a4"), ParseSquare("a5"), ParseSquare("a3"), GDog, SRabbit),
			Pass,
		},
	}, {
		input: "Dh3s Rh2n Rg1e Rf1e",
		want: []Step{
			MakeDefault(ParseSquare("h3"), ParseSquare("h2"), GDog),
			MakeDefault(ParseSquare("h2"), ParseSquare("h3"), GRabbit),
			MakeDefault(ParseSquare("g1"), ParseSquare("h1"), GRabbit),
			MakeDefault(ParseSquare("f1"), ParseSquare("g1"), GRabbit),
			Pass,
		},
	}, {
		input: "Md2n Dh2n Md3n Md4s",
		want: []Step{
			MakeDefault(ParseSquare("d2"), ParseSquare("d3"), GCamel),
			MakeDefault(ParseSquare("h2"), ParseSquare("h3"), GDog),
			MakeDefault(ParseSquare("d3"), ParseSquare("d4"), GCamel),
			MakeDefault(ParseSquare("d4"), ParseSquare("d3"), GCamel),
			Pass,
		},
	}, {
		input: "Hc4s rc5s Hc3w rc4s rc3x",
		want: []Step{
			MakeAlternate(ParseSquare("c4"), ParseSquare("c3"), ParseSquare("c5"), GHorse, SRabbit),
			MakeAlternateCapture(ParseSquare("c3"), ParseSquare("b3"), ParseSquare("c4"), GHorse, SRabbit, SRabbit),
			Pass,
		},
	}, {
		input: "Rf1w Rg1w Hg2w Hb3w Cc3x",
		want: []Step{
			MakeDefault(ParseSquare("f1"), ParseSquare("e1"), GRabbit),
			MakeDefault(ParseSquare("g1"), ParseSquare("f1"), GRabbit),
			MakeDefault(ParseSquare("g2"), ParseSquare("f2"), GHorse),
			MakeDefaultCapture(ParseSquare("b3"), ParseSquare("a3"), GHorse, GCat),
			Pass,
		},
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
