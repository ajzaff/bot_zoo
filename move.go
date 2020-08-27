package zoo

var (
	Steps       [64]Bitboard
	RabbitSteps [2][64]Bitboard
)

func init() {
	for i := 0; i < 64; i++ {
		b := Bitboard(1) << i
		steps := b.Neighbors()
		Steps[i] = steps
		grSteps := steps
		if b&NotRank1 != 0 { // rabbits can't move backwards.
			grSteps ^= b << 8
		}
		srSteps := steps
		if b&NotRank8 != 0 {
			srSteps ^= b << 8
		}
		RabbitSteps[0][i] = grSteps
		RabbitSteps[1][i] = srSteps
	}
}

type Step struct {
	Src, Dest uint8
}
