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
			MakeStep(GDog, A4, A5),
		},
	}, {
		input: "Da4n Da5e",
		want: []Step{
			MakeStep(GDog, A4, A5),
			MakeStep(GDog, A5, B5),
		},
	}, {
		input: "Da4n Ra3n",
		want: []Step{
			MakeStep(GDog, A4, A5),
			MakeStep(GRabbit, A3, A4),
		},
	}, {
		input: "Da4n ra3n",
		want: []Step{
			MakeStep(GDog, A4, A5),
			MakeStep(SRabbit, A3, A4),
		},
	}, {
		input: "Dh3s Rh2n Rg1e Rf1e",
		want: []Step{
			MakeStep(GDog, H3, H2),
			MakeStep(GRabbit, H2, H3),
			MakeStep(GRabbit, G1, H1),
			MakeStep(GRabbit, F1, G1),
		},
	}, {
		input: "Md2n Dh2n Md3n Md4s",
		want: []Step{
			MakeStep(GCamel, D2, D3),
			MakeStep(GDog, H2, H3),
			MakeStep(GCamel, D3, D4),
			MakeStep(GCamel, D4, D3),
		},
	}, {
		input: "Hc4s rc5s Hc3w rc4s rc3x",
		want: []Step{
			MakeStep(GHorse, C4, C3),
			MakeStep(SRabbit, C5, C4),
			MakeStep(GHorse, C3, B3),
			MakeStep(SRabbit, C4, C3),
			MakeCapture(SRabbit, C3),
		},
	}, {
		input: "Rf1w Rg1w Hg2w Hb3w Cc3x",
		want: []Step{
			MakeStep(GRabbit, F1, E1),
			MakeStep(GRabbit, G1, F1),
			MakeStep(GHorse, G2, F2),
			MakeStep(GHorse, B3, A3),
			MakeCapture(GCat, C3),
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
