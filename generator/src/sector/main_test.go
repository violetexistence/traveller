package sector

import (
	"testing"
)

func TestGetGovernment(t *testing.T) {
	for i := population_0; i <= population_A; i++ {
		for j := 0; j < 1000; j++ {
			actual := getGovernment(i)
			if actual < government_0 || actual > government_D {
				t.Fatalf("getGovernment(%d): %d?? Min: %d, Max: %d", i, actual, government_0, government_D)
			}
		}
	}
}

func TestDefineTradeCode(t *testing.T) {
}
