package scaling

import (
	"runtime"
	"testing"
)

func TestCoreScaling(t *testing.T) {
	max := runtime.NumCPU()

	for i := 1; i <= max; i++ {
		runtime.GOMAXPROCS(i)

		t.Logf("running scaling experiment with %d cores", i)
	}
}
