package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
)

const (
	auto = iota
	roll
	no
)

type Masking struct {
	Free int
	Throw int
	Time Hours
}

type Hours struct {
	At05G float64
	At1G float64
	At2G float64
}

var masking_table = map[string]Masking {
	"O5":				Masking{ Free: auto,	Throw: 0,		Time: Hours { At05G: 336,			At1G: 240,		At2G: 168,		} },
	"B0": 			Masking{ Free: auto,	Throw: 0,		Time: Hours { At05G: 216,			At1G: 151,		At2G: 108,		} },
	"B5": 			Masking{ Free: auto,	Throw: 0,		Time: Hours { At05G: 153.6,		At1G: 108,		At2G: 76.8,		} },
	"A0": 			Masking{ Free: roll,	Throw: 17,	Time: Hours { At05G: 124.8,		At1G: 89,			At2G: 62.4,		} },
	"A5": 			Masking{ Free: roll,	Throw: 16,	Time: Hours { At05G: 103.2,		At1G: 72,			At2G: 50.4,		} },
	"F0": 			Masking{ Free: roll,	Throw: 15, 	Time: Hours { At05G: 91.2,		At1G: 64.8,		At2G: 45,			} },
	"F5": 			Masking{ Free: roll,	Throw: 14, 	Time: Hours { At05G: 86.4,		At1G: 62.4,		At2G: 43,			} },
	"G0": 			Masking{ Free: roll,	Throw: 12, 	Time: Hours { At05G: 81.6,		At1G: 57.6,		At2G: 40,			} },
	"G5": 			Masking{ Free: no,		Throw: 0,		Time: Hours { At05G: 76.8,		At1G: 52.8,		At2G: 38,			} },
	"K0": 			Masking{ Free: no,		Throw: 0, 	Time: Hours { At05G: 72,			At1G: 50.4,		At2G: 36,			} },
	"K5": 			Masking{ Free: no,		Throw: 0, 	Time: Hours { At05G: 67.2,		At1G: 48,			At2G: 34,			} },
	"M0": 			Masking{ Free: no,		Throw: 0, 	Time: Hours { At05G: 62.4,		At1G: 44,			At2G: 31,			} },
	"M5": 			Masking{ Free: no,		Throw: 0, 	Time: Hours { At05G: 45,			At1G: 32,			At2G: 22,			} },
	"M9": 			Masking{ Free: no,		Throw: 0, 	Time: Hours { At05G: 29,			At1G: 20,			At2G: 14,			} },
	"Unknown":	Masking{ Free: roll,	Throw: 8,		Time: Hours { At05G: 62.4,		At1G: 44,			At2G: 31,			} },
}

var free_jump_table = map[string]Hours {
	"Asteroid":					Hours { At05G: 0.9,		At1G: 0.7,	At2G: 0.5,	},
	"1,000 miles":			Hours { At05G: 2.7,		At1G: 1.9,	At2G: 1.3,	},
	"2,000 miles":			Hours { At05G: 3.8,		At1G: 2.7,	At2G: 1.9,	},
	"3,000 miles":			Hours { At05G: 4.6,		At1G: 3.3,	At2G: 2.3,	},
	"4,000 miles":			Hours { At05G: 5.4,		At1G: 3.8,	At2G: 2.7,	},
	"5,000 miles":			Hours { At05G: 6.0,		At1G: 4.2,	At2G: 3.0,	},
	"6,000 miles":			Hours { At05G: 6.6,		At1G: 4.6,	At2G: 3.3,	},
	"7,000 miles":			Hours { At05G: 7.1,		At1G: 5.0,	At2G: 3.5,	},
	"8,000 miles":			Hours { At05G: 7.6,		At1G: 5.4,	At2G: 3.8,	},
	"9,000 miles":			Hours { At05G: 8.0,		At1G: 5.7,	At2G: 4.0,	},
	"10,000 miles":			Hours { At05G: 8.5,		At1G: 6.0,	At2G: 4.2,	},
	"Small Gas Giant":	Hours { At05G: 14.7,	At1G: 10.4,	At2G: 7.3,	},
	"Medium Gas Giant":	Hours { At05G: 19.0,	At1G: 13.4,	At2G: 9.5,	},
	"Large Gas Giant":	Hours { At05G: 24.0,	At1G: 17.0,	At2G: 12.0,	},
}

var time_factor_table_1 = map[int]float64 {
	1: 0.2,
	2: 0.4,
	3: 0.6,
	4: 0.8,
	5: 0.9,
	6: 1.0,
}

var time_factor_table_2 = map[string]float64 {
	"G5": 0.4,
	"K0": 0.5,
	"K5": 0.8,
	"M0": 0.8,
	"M5": 0.8,
	"M9": 0.9,
}

func main() {	
	dest := flag.String("dest", "Unknown", "The spectral class of the destination system primary.")
	acceleration := flag.Float64("accel", 1.0, "Acceleration in gees, 0.5, 1, or 2")
	world_diameter := flag.String("wd", "8,000 miles", "Main world diameter.")

	flag.Parse()
	
	masking_row, ok := masking_table[*dest]
	var is_free bool
	
	if ok {
		fmt.Printf("Destination: %s\n", *dest)

		var hours float64
		
		switch masking_row.Free {
		case auto:
			is_free = true
			free_jump_row := free_jump_table[*world_diameter]
			hours = get_time(free_jump_row, *acceleration)
		case roll:
			is_free = throw(3) >= masking_row.Throw
			if is_free {
				free_jump_row := free_jump_table[*world_diameter]
				hours = get_time(free_jump_row, *acceleration)
			} else {
				factor := time_factor_table_1[throw(1)]
				hours = factor * get_time(masking_row.Time, *acceleration)
			}
		case no:
			is_free = false
			is_near_side := throw(1) < 4
			primary_factor := time_factor_table_2[*dest]
			world_factor := time_factor_table_1[throw(1)]
			
			if is_near_side {
				hours = primary_factor * get_time(masking_row.Time, *acceleration)
			} else {
				hours = math.Max(primary_factor, world_factor) * get_time(masking_row.Time, *acceleration)
			}
		}

		if is_free {
			fmt.Printf("Free Jump\n")
		} else {
			fmt.Printf("Masked Jump\n")
		}

		fmt.Printf("Travel Time: %v (hours)\n", hours)

	} else {
		fmt.Printf("Invalid destination: %s\n", *dest)
	}

}

func throw(num int) int {
	sum := 0
	for i := 0; i < num; i++ {
		sum += (rand.Intn(6) + 1)
	}
	return sum
}

func get_time(hours Hours, acceleration float64) float64 {
	switch acceleration {
	case 0.5:
		return hours.At05G
	case 1:
		return hours.At1G
	case 2:
		return hours.At2G
	default:
		return hours.At1G
	}
}
