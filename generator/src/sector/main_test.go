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

func TestGetHzVar(t *testing.T) {
	star := star{
		class: spectralClass{
			letter:  "M",
			numeral: 4,
		},
		size: "II",
	}

	var results = map[int]int{
		-2: 0,
		-1: 0,
		0:  0,
		1:  0,
		2:  0,
	}

	for i := 0; i < 500; i++ {
		actual := getHzVar(star)
		switch actual {
		case -2, -1, 0, 1, 2:
			results[actual] = results[actual] + 1
		default:
			t.Fatalf("Bad: %d", actual)
		}
	}

	t.Logf("results: %v", results)
}
