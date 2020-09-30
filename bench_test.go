package zoo

import (
	"testing"
	"time"
)

func BenchmarkOpening(b *testing.B) {
	engine, err := NewEngine(&EngineSettings{
		Seed: 1337,
	}, &AEISettings{
		LogProtocolTraffic: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	if err := engine.ExecuteCommand("setposition g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]"); err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		engine.searchRoot(false)
		time.Sleep(1 * time.Second)
		engine.wg.Wait()
	}
}
