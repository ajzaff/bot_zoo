package zoo

import (
	"testing"
	"time"
)

func BenchmarkOpening(b *testing.B) {
	engine := NewEngine()

	engine.EngineSettings.Seed = 1337
	engine.AEISettings.LogProtocolTraffic = true
	engine.AEISettings.LogVerbosePosition = true

	if err := engine.ExecuteCommand("setposition g [rrrrrrrrhdcemcdh                                HDCMECDHRRRRRRRR]"); err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		engine.searchRoot(false)
		time.Sleep(1 * time.Second)
		engine.wg.Wait()
	}
}
