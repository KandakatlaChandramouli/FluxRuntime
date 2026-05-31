package batching

import "testing"

func TestAdaptiveBatching(t *testing.T) {
	sizes := []int{
		1,
		2,
		4,
		8,
		16,
		32,
		64,
	}

	for _, s := range sizes {
		t.Logf("evaluating batch size=%d", s)
	}
}
