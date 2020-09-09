package zoo

import "testing"

func BenchmarkJunke11(b *testing.B) {
	e := NewEngine(1337)
	p, err := ParseShortPosition("g [rr        c M       r rr  r  E      rHe  D  HrR R   CmRrRR  RRdR]")
	if err != nil {
		b.Fatal(err)
	}
	e.SetPos(p)
	for n := 0; n < b.N; n++ {
		e.GoFixed(10)
	}
}
