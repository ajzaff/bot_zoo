package zoo

import (
	"fmt"
	"strings"
)

func (b Bitboard) String() string {
	var sb strings.Builder
	sb.WriteString("\n +-----------------+\n")
	for i := 7; i >= 0; i-- {
		fmt.Fprintf(&sb, "%d| ", i+1)
		for j := 0; j < 8; j++ {
			at := Square(8*i + j)
			atB := at.Bitboard()
			if b&atB != 0 {
				sb.WriteByte('X')
			} else {
				sb.WriteByte('.')
			}
			sb.WriteByte(' ')
		}
		sb.WriteString("|\n")
	}
	sb.WriteString(" +-----------------+\n   a b c d e f g h")
	return sb.String()
}
