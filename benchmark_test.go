package zoo

import "testing"

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkJunke11(b *testing.B) {
	e := NewEngine(1337)
	e.SetLog(nopWriter{})
	p, err := ParseShortPosition("g [rr        c M       r rr  r  E      rHe  D  HrR R   CmRrRR  RRdR]")
	if err != nil {
		b.Fatal(err)
	}
	e.SetPos(p)
	for n := 0; n < b.N; n++ {
		e.GoFixed(10)
	}
}
