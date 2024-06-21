package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
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

type Jump struct {
	SpectralClass string
	Diameter string
	Type string
	TravelTime float64
}

var masking_table = map[string]Masking {
	"O5":				{ Free: auto,	Throw: 0,		Time: Hours { At05G: 336,			At1G: 240,		At2G: 168,		} },
	"B0": 			{ Free: auto,	Throw: 0,		Time: Hours { At05G: 216,			At1G: 151,		At2G: 108,		} },
	"B5": 			{ Free: auto,	Throw: 0,		Time: Hours { At05G: 153.6,		At1G: 108,		At2G: 76.8,		} },
	"A0": 			{ Free: roll,	Throw: 17,	Time: Hours { At05G: 124.8,		At1G: 89,			At2G: 62.4,		} },
	"A5": 			{ Free: roll,	Throw: 16,	Time: Hours { At05G: 103.2,		At1G: 72,			At2G: 50.4,		} },
	"F0": 			{ Free: roll,	Throw: 15, 	Time: Hours { At05G: 91.2,		At1G: 64.8,		At2G: 45,			} },
	"F5": 			{ Free: roll,	Throw: 14, 	Time: Hours { At05G: 86.4,		At1G: 62.4,		At2G: 43,			} },
	"G0": 			{ Free: roll,	Throw: 12, 	Time: Hours { At05G: 81.6,		At1G: 57.6,		At2G: 40,			} },
	"G5": 			{ Free: no,		Throw: 0,		Time: Hours { At05G: 76.8,		At1G: 52.8,		At2G: 38,			} },
	"K0": 			{ Free: no,		Throw: 0, 	Time: Hours { At05G: 72,			At1G: 50.4,		At2G: 36,			} },
	"K5": 			{ Free: no,		Throw: 0, 	Time: Hours { At05G: 67.2,		At1G: 48,			At2G: 34,			} },
	"M0": 			{ Free: no,		Throw: 0, 	Time: Hours { At05G: 62.4,		At1G: 44,			At2G: 31,			} },
	"M5": 			{ Free: no,		Throw: 0, 	Time: Hours { At05G: 45,			At1G: 32,			At2G: 22,			} },
	"M9": 			{ Free: no,		Throw: 0, 	Time: Hours { At05G: 29,			At1G: 20,			At2G: 14,			} },
	"Unknown":	{ Free: roll,	Throw: 8,		Time: Hours { At05G: 62.4,		At1G: 44,			At2G: 31,			} },
}

var free_jump_table = map[string]Hours {
	"Asteroid":					{ At05G: 0.9,		At1G: 0.7,	At2G: 0.5,	},
	"1,000 miles":			{ At05G: 2.7,		At1G: 1.9,	At2G: 1.3,	},
	"2,000 miles":			{ At05G: 3.8,		At1G: 2.7,	At2G: 1.9,	},
	"3,000 miles":			{ At05G: 4.6,		At1G: 3.3,	At2G: 2.3,	},
	"4,000 miles":			{ At05G: 5.4,		At1G: 3.8,	At2G: 2.7,	},
	"5,000 miles":			{ At05G: 6.0,		At1G: 4.2,	At2G: 3.0,	},
	"6,000 miles":			{ At05G: 6.6,		At1G: 4.6,	At2G: 3.3,	},
	"7,000 miles":			{ At05G: 7.1,		At1G: 5.0,	At2G: 3.5,	},
	"8,000 miles":			{ At05G: 7.6,		At1G: 5.4,	At2G: 3.8,	},
	"9,000 miles":			{ At05G: 8.0,		At1G: 5.7,	At2G: 4.0,	},
	"10,000 miles":			{ At05G: 8.5,		At1G: 6.0,	At2G: 4.2,	},
	"Small Gas Giant":	{ At05G: 14.7,	At1G: 10.4,	At2G: 7.3,	},
	"Medium Gas Giant":	{ At05G: 19.0,	At1G: 13.4,	At2G: 9.5,	},
	"Large Gas Giant":	{ At05G: 24.0,	At1G: 17.0,	At2G: 12.0,	},
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

var (
	origin_spectral_class string
	origin_diameter string
	destination_spectral_class string
	destination_diameter string
	g_rating float64
)

var (
	outjump_plan Jump
	breakout_plan Jump
)

func mainx() {	
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("(Origin) Spectral Class").
				Options(
					huh.NewOption("O5", "O5"),
					huh.NewOption("B0", "B0"),
					huh.NewOption("B5", "B5"),
					huh.NewOption("A0", "A0"),
					huh.NewOption("A5", "A5"),
					huh.NewOption("F0", "F0"),
					huh.NewOption("F5", "F5"),
					huh.NewOption("G0", "G0").Selected(true),
					huh.NewOption("G5", "G5"),
					huh.NewOption("K0", "K0"),
					huh.NewOption("K5", "K5"),
					huh.NewOption("M0", "M0"),
					huh.NewOption("M5", "M5"),
					huh.NewOption("M9", "M9"),
					huh.NewOption("Unknown", "Unknown"),
				).
				Value(&origin_spectral_class),

		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("(Origin) World Diameter").
				Options(
					huh.NewOption("Asteroid", "Asteroid"),
					huh.NewOption("1,000 miles", "1,000 miles"),
					huh.NewOption("2,000 miles", "2,000 miles"),
					huh.NewOption("3,000 miles", "3,000 miles"),
					huh.NewOption("4,000 miles", "4,000 miles"),
					huh.NewOption("5,000 miles", "5,000 miles"),
					huh.NewOption("6,000 miles", "6,000 miles"),
					huh.NewOption("7,000 miles", "7,000 miles"),
					huh.NewOption("8,000 miles", "8,000 miles").Selected(true),
					huh.NewOption("9,000 miles", "9,000 miles"),
					huh.NewOption("10,000 miles", "10,000 miles"),
					huh.NewOption("Small Gas Giant", "Small Gas Giant"),
					huh.NewOption("Medium Gas Giant", "Medium Gas Giant"),
					huh.NewOption("Large Gas Giant", "Large Gas Giant"),
				).
				Value(&origin_diameter),

		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("(Destination) Spectral Class").
				Options(
					huh.NewOption("O5", "O5"),
					huh.NewOption("B0", "B0"),
					huh.NewOption("B5", "B5"),
					huh.NewOption("A0", "A0"),
					huh.NewOption("A5", "A5"),
					huh.NewOption("F0", "F0"),
					huh.NewOption("F5", "F5"),
					huh.NewOption("G0", "G0").Selected(true),
					huh.NewOption("G5", "G5"),
					huh.NewOption("K0", "K0"),
					huh.NewOption("K5", "K5"),
					huh.NewOption("M0", "M0"),
					huh.NewOption("M5", "M5"),
					huh.NewOption("M9", "M9"),
					huh.NewOption("Unknown", "Unknown"),
				).
				Value(&destination_spectral_class),

		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("(Destination) World Diameter").
				Options(
					huh.NewOption("Asteroid", "Asteroid"),
					huh.NewOption("1,000 miles", "1,000 miles"),
					huh.NewOption("2,000 miles", "2,000 miles"),
					huh.NewOption("3,000 miles", "3,000 miles"),
					huh.NewOption("4,000 miles", "4,000 miles"),
					huh.NewOption("5,000 miles", "5,000 miles"),
					huh.NewOption("6,000 miles", "6,000 miles"),
					huh.NewOption("7,000 miles", "7,000 miles"),
					huh.NewOption("8,000 miles", "8,000 miles").Selected(true),
					huh.NewOption("9,000 miles", "9,000 miles"),
					huh.NewOption("10,000 miles", "10,000 miles"),
					huh.NewOption("Small Gas Giant", "Small Gas Giant"),
					huh.NewOption("Medium Gas Giant", "Medium Gas Giant"),
					huh.NewOption("Large Gas Giant", "Large Gas Giant"),
				).
				Value(&destination_diameter),

		),

		huh.NewGroup(
			huh.NewSelect[float64]().
				Title("M-Drive Acceleration").
				Options(
					huh.NewOption("0.5G", 0.5),
					huh.NewOption("1G", 1.0).Selected(true),
					huh.NewOption("2G", 2.0),
				).
				Value(&g_rating),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = spinner.New().
		Title("Calculating travel time...").
		Action(create_plan).
		Run()

	if err != nil {
		log.Fatal(err)
	}

	total_travel_time := outjump_plan.TravelTime + 168.0 + breakout_plan.TravelTime

	fmt.Printf("\nOutjump: %s, %.2f (hours)\n", outjump_plan.Type, outjump_plan.TravelTime)
	fmt.Printf("Breakout: %s, %.2f (hours)\n", breakout_plan.Type, breakout_plan.TravelTime)
	fmt.Printf("Total Travel Time Est. %.2f (hours)\n", total_travel_time)
}

func create_plan() {
	outjump_plan = compute_jump(origin_spectral_class, origin_diameter, g_rating)
	breakout_plan = compute_jump(destination_spectral_class, destination_diameter, g_rating)	
}

func compute_jump(spectral_class string, world_diameter string, acceleration float64) Jump {
	jump := Jump {
		SpectralClass: spectral_class,
		Diameter: world_diameter,
	}
	masking_row := masking_table[spectral_class]
	
	switch masking_row.Free {
	case auto:
		jump.Type = "Free"
		free_jump_row := free_jump_table[world_diameter]
		jump.TravelTime = choose_hours(free_jump_row, acceleration)
	case roll:
		if dice(3) >= masking_row.Throw {
			jump.Type = "Free"
			free_jump_row := free_jump_table[world_diameter]
			jump.TravelTime = choose_hours(free_jump_row, acceleration)
		} else {
			jump.Type = "Masked"
			factor := time_factor_table_1[dice(1)]
			jump.TravelTime = factor * choose_hours(masking_row.Time, acceleration)
		}
	case no:
		jump.Type = "Masked"
		is_near_side := dice(1) < 4
		primary_factor := time_factor_table_2[spectral_class]
		world_factor := time_factor_table_1[dice(1)]
		
		if is_near_side {
			jump.TravelTime = primary_factor * choose_hours(masking_row.Time, acceleration)
		} else {
			jump.TravelTime = math.Max(primary_factor, world_factor) * choose_hours(masking_row.Time, acceleration)
		}
	}

	return jump	
}

func dice(num int) int {
	sum := 0
	for i := 0; i < num; i++ {
		sum += (rand.Intn(6) + 1)
	}
	return sum
}

func choose_hours(hours Hours, acceleration float64) float64 {
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
