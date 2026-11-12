// Package sky holds asterism's constellation catalog and the pure rendering
// core: projecting stars from celestial coordinates onto a tangent plane,
// fitting them into a character grid, rasterizing the figure lines, and
// plotting the stars. Nothing in this package touches a real terminal or the
// process environment, so every function here is deterministic and can be
// unit-tested offline.
//
// Coordinate and magnitude values in the catalog are approximate J2000 catalog
// figures (right ascension in hours, declination in degrees, apparent visual
// magnitude). They are chosen so the projected stick figures are recognizable,
// not for astrometric precision. The figure line segments follow the common
// IAU / Stellarium "stick figure" for each constellation.
package sky

import (
	"sort"
	"strings"
)

// Star is a single catalog star. Name is the proper or Bayer name for the
// notable stars and empty for anonymous ones.
type Star struct {
	Name string  // proper/Bayer name, empty if unnamed
	RA   float64 // right ascension, hours in [0, 24)
	Dec  float64 // declination, degrees in [-90, 90]
	Mag  float64 // apparent visual magnitude
}

// Constellation is one catalog entry: its identity, best-viewing metadata, the
// bright stars, and the figure line segments that connect them (each segment is
// a pair of indices into Stars).
type Constellation struct {
	Latin      string   // Latin name, e.g. "Orion"
	Abbr       string   // IAU 3-letter abbreviation, e.g. "Ori"
	Common     string   // common / English name
	Hemisphere string   // "Northern", "Southern", or "Equatorial"
	BestMonth  string   // month it culminates in the evening sky
	Stars      []Star   // bright stars, in a stable order
	Lines      [][2]int // figure segments as pairs of Stars indices
}

// BrightestStar returns the star with the smallest (brightest) magnitude.
// It assumes the constellation has at least one star, which catalog integrity
// guarantees.
func (c Constellation) BrightestStar() Star {
	best := c.Stars[0]
	for _, s := range c.Stars[1:] {
		if s.Mag < best.Mag {
			best = s
		}
	}
	return best
}

// catalog is the embedded constellation data. It is unexported; callers reach
// it through Catalog, Lookup, and Names so lookups stay case-insensitive and
// abbreviation-aware.
//
// Data note: RA/Dec are approximate J2000 catalog values and magnitudes are
// approximate apparent visual magnitudes, curated for recognizable rendering.
var catalog = []Constellation{
	{
		Latin: "Orion", Abbr: "Ori", Common: "The Hunter",
		Hemisphere: "Equatorial", BestMonth: "January",
		Stars: []Star{
			{"Betelgeuse", 5.919, 7.407, 0.50}, // 0
			{"Rigel", 5.242, -8.202, 0.18},     // 1
			{"Bellatrix", 5.418, 6.350, 1.64},  // 2
			{"Mintaka", 5.533, -0.299, 2.23},   // 3 belt
			{"Alnilam", 5.604, -1.202, 1.69},   // 4 belt
			{"Alnitak", 5.679, -1.943, 1.77},   // 5 belt
			{"Saiph", 5.796, -9.670, 2.06},     // 6
			{"Meissa", 5.585, 9.934, 3.39},     // 7 head
		},
		Lines: [][2]int{
			{0, 2},         // shoulders
			{0, 5},         // right shoulder to belt
			{2, 3},         // left shoulder to belt
			{3, 4}, {4, 5}, // the belt
			{5, 6},         // belt to Saiph
			{3, 1},         // belt to Rigel
			{0, 7}, {2, 7}, // head
		},
	},
	{
		Latin: "Ursa Major", Abbr: "UMa", Common: "The Great Bear (Big Dipper)",
		Hemisphere: "Northern", BestMonth: "April",
		Stars: []Star{
			{"Dubhe", 11.062, 61.751, 1.79},  // 0
			{"Merak", 11.031, 56.383, 2.37},  // 1
			{"Phecda", 11.897, 53.695, 2.44}, // 2
			{"Megrez", 12.257, 57.033, 3.31}, // 3
			{"Alioth", 12.900, 55.960, 1.77}, // 4
			{"Mizar", 13.399, 54.925, 2.23},  // 5
			{"Alkaid", 13.792, 49.313, 1.86}, // 6
		},
		Lines: [][2]int{
			{0, 1}, {1, 2}, {2, 3}, {3, 0}, // bowl of the Dipper
			{3, 4}, {4, 5}, {5, 6}, // handle
		},
	},
	{
		Latin: "Ursa Minor", Abbr: "UMi", Common: "The Little Bear (Little Dipper)",
		Hemisphere: "Northern", BestMonth: "June",
		Stars: []Star{
			{"Polaris", 2.530, 89.264, 1.98},      // 0 (the North Star)
			{"Kochab", 14.845, 74.156, 2.07},      // 1
			{"Pherkad", 15.345, 71.834, 3.00},     // 2
			{"Yildun", 17.537, 86.586, 4.36},      // 3
			{"Epsilon UMi", 16.766, 82.037, 4.23}, // 4
			{"Zeta UMi", 15.734, 77.794, 4.32},    // 5
			{"Eta UMi", 16.291, 75.755, 4.95},     // 6
		},
		Lines: [][2]int{
			{0, 3}, {3, 4}, {4, 5}, // handle
			{5, 6}, {6, 1}, {1, 2}, {2, 5}, // bowl
		},
	},
	{
		Latin: "Cassiopeia", Abbr: "Cas", Common: "The Queen",
		Hemisphere: "Northern", BestMonth: "November",
		Stars: []Star{
			{"Schedar", 0.675, 56.537, 2.24},   // 0
			{"Caph", 0.153, 59.150, 2.28},      // 1
			{"Gamma Cas", 0.945, 60.717, 2.47}, // 2
			{"Ruchbah", 1.430, 60.235, 2.68},   // 3
			{"Segin", 1.906, 63.670, 3.35},     // 4
		},
		Lines: [][2]int{
			{1, 0}, {0, 2}, {2, 3}, {3, 4}, // the W
		},
	},
	{
		Latin: "Cygnus", Abbr: "Cyg", Common: "The Swan (Northern Cross)",
		Hemisphere: "Northern", BestMonth: "September",
		Stars: []Star{
			{"Deneb", 20.690, 45.280, 1.25},     // 0
			{"Albireo", 19.512, 27.960, 3.18},   // 1
			{"Sadr", 20.371, 40.257, 2.23},      // 2
			{"Gienah", 20.770, 33.970, 2.48},    // 3
			{"Delta Cyg", 19.750, 45.131, 2.87}, // 4
		},
		Lines: [][2]int{
			{0, 2}, {2, 1}, // spine of the cross
			{4, 2}, {2, 3}, // crossbar
		},
	},
	{
		Latin: "Lyra", Abbr: "Lyr", Common: "The Lyre",
		Hemisphere: "Northern", BestMonth: "August",
		Stars: []Star{
			{"Vega", 18.615, 38.784, 0.03},      // 0
			{"Sheliak", 18.834, 33.363, 3.52},   // 1
			{"Sulafat", 18.982, 32.690, 3.25},   // 2
			{"Delta Lyr", 18.908, 36.899, 4.30}, // 3
			{"Zeta Lyr", 18.746, 37.605, 4.36},  // 4
		},
		Lines: [][2]int{
			{0, 4}, {0, 1}, // Vega to the frame
			{4, 3}, {3, 2}, {2, 1}, {1, 4}, // parallelogram
		},
	},
	{
		Latin: "Aquila", Abbr: "Aql", Common: "The Eagle",
		Hemisphere: "Equatorial", BestMonth: "August",
		Stars: []Star{
			{"Altair", 19.846, 8.868, 0.76},         // 0
			{"Tarazed", 19.771, 10.613, 2.72},       // 1
			{"Alshain", 19.922, 6.407, 3.71},        // 2
			{"Deneb el Okab", 19.090, 13.863, 2.99}, // 3
			{"Theta Aql", 20.188, -0.821, 3.23},     // 4
			{"Delta Aql", 19.425, 3.115, 3.36},      // 5
			{"Lambda Aql", 19.104, -4.882, 3.43},    // 6
		},
		Lines: [][2]int{
			{3, 1}, {1, 0}, {0, 2}, // head and shoulders
			{0, 4},         // to the tail
			{0, 5}, {5, 6}, // wing
		},
	},
	{
		Latin: "Scorpius", Abbr: "Sco", Common: "The Scorpion",
		Hemisphere: "Southern", BestMonth: "July",
		Stars: []Star{
			{"Antares", 16.490, -26.432, 1.06},     // 0
			{"Shaula", 17.560, -37.104, 1.62},      // 1 stinger
			{"Sargas", 17.622, -42.998, 1.86},      // 2
			{"Dschubba", 16.005, -22.622, 2.29},    // 3
			{"Pi Sco", 15.981, -26.114, 2.89},      // 4
			{"Acrab", 16.090, -19.805, 2.56},       // 5
			{"Tau Sco", 16.598, -28.216, 2.82},     // 6
			{"Epsilon Sco", 16.836, -34.293, 2.29}, // 7
			{"Mu Sco", 16.864, -38.017, 3.00},      // 8
			{"Zeta Sco", 16.909, -42.362, 3.62},    // 9
			{"Eta Sco", 17.203, -43.239, 3.32},     // 10
			{"Kappa Sco", 17.708, -39.030, 2.41},   // 11
			{"Lesath", 17.512, -37.296, 2.70},      // 12
		},
		Lines: [][2]int{
			{5, 3}, {3, 4}, {3, 0}, // head and claws
			{0, 6}, {6, 7}, {7, 8}, {8, 9}, {9, 10}, // curving body
			{10, 2}, {2, 11}, {11, 12}, {12, 1}, // tail to the stinger
		},
	},
	{
		Latin: "Sagittarius", Abbr: "Sgr", Common: "The Archer (Teapot)",
		Hemisphere: "Southern", BestMonth: "August",
		Stars: []Star{
			{"Kaus Australis", 18.403, -34.385, 1.85}, // 0
			{"Nunki", 18.921, -26.297, 2.05},          // 1
			{"Ascella", 19.044, -29.880, 2.60},        // 2
			{"Kaus Media", 18.350, -29.828, 2.70},     // 3
			{"Kaus Borealis", 18.466, -25.422, 2.81},  // 4
			{"Phi Sgr", 18.763, -26.991, 3.17},        // 5
			{"Alnasl", 18.097, -30.424, 2.98},         // 6
			{"Tau Sgr", 19.115, -27.670, 3.32},        // 7
		},
		Lines: [][2]int{
			{6, 3}, {3, 0}, {0, 2}, // spout and base
			{2, 1}, {1, 5}, {5, 4}, {4, 3}, // body and lid
			{1, 7}, {7, 2}, // handle
		},
	},
	{
		Latin: "Leo", Abbr: "Leo", Common: "The Lion",
		Hemisphere: "Northern", BestMonth: "April",
		Stars: []Star{
			{"Regulus", 10.139, 11.967, 1.36},    // 0
			{"Denebola", 11.818, 14.572, 2.14},   // 1
			{"Algieba", 10.333, 19.842, 2.08},    // 2
			{"Zosma", 11.235, 20.524, 2.56},      // 3
			{"Epsilon Leo", 9.764, 23.774, 2.98}, // 4
			{"Eta Leo", 10.123, 16.763, 3.48},    // 5
			{"Theta Leo", 11.237, 15.430, 3.34},  // 6
			{"Adhafera", 10.278, 23.417, 3.44},   // 7
			{"Mu Leo", 9.880, 26.007, 3.88},      // 8
		},
		Lines: [][2]int{
			{0, 5}, {5, 2}, {2, 7}, {7, 8}, {8, 4}, // the Sickle (head)
			{0, 6}, {6, 1}, {1, 3}, {3, 2}, // body and hindquarters
		},
	},
	{
		Latin: "Taurus", Abbr: "Tau", Common: "The Bull",
		Hemisphere: "Northern", BestMonth: "January",
		Stars: []Star{
			{"Aldebaran", 4.599, 16.509, 0.85},   // 0
			{"Elnath", 5.438, 28.607, 1.65},      // 1
			{"Zeta Tau", 5.627, 21.143, 3.00},    // 2
			{"Theta Tau", 4.477, 15.871, 3.40},   // 3
			{"Gamma Tau", 4.330, 15.628, 3.65},   // 4
			{"Epsilon Tau", 4.478, 19.180, 3.53}, // 5
			{"Lambda Tau", 4.011, 12.490, 3.47},  // 6
			{"Alcyone", 3.791, 24.105, 2.87},     // 7 (Pleiades)
		},
		Lines: [][2]int{
			{1, 5}, {5, 0}, {0, 4}, {4, 6}, // face V and horn
			{0, 2}, // second horn
			{5, 7}, // toward the Pleiades
		},
	},
	{
		Latin: "Gemini", Abbr: "Gem", Common: "The Twins",
		Hemisphere: "Northern", BestMonth: "February",
		Stars: []Star{
			{"Pollux", 7.755, 28.026, 1.14},    // 0
			{"Castor", 7.577, 31.888, 1.58},    // 1
			{"Alhena", 6.629, 16.399, 1.93},    // 2
			{"Wasat", 7.335, 21.982, 3.53},     // 3
			{"Mebsuta", 6.732, 25.131, 3.06},   // 4
			{"Mekbuda", 7.068, 20.570, 3.79},   // 5
			{"Tejat", 6.383, 22.514, 2.87},     // 6
			{"Propus", 6.248, 22.507, 3.28},    // 7
			{"Xi Gem", 6.755, 12.896, 3.35},    // 8
			{"Kappa Gem", 7.740, 24.398, 3.57}, // 9
		},
		Lines: [][2]int{
			{1, 4}, {4, 6}, {6, 7}, // Castor's body
			{0, 3}, {3, 5}, {5, 2}, // Pollux's body
			{1, 0}, // heads
			{4, 3}, // waist link
			{3, 8}, // foot
			{0, 9}, // Pollux head branch
		},
	},
	{
		Latin: "Canis Major", Abbr: "CMa", Common: "The Great Dog",
		Hemisphere: "Southern", BestMonth: "February",
		Stars: []Star{
			{"Sirius", 6.752, -16.716, -1.46},      // 0
			{"Mirzam", 6.378, -17.956, 1.98},       // 1
			{"Wezen", 7.140, -26.393, 1.83},        // 2
			{"Adhara", 6.977, -28.972, 1.50},       // 3
			{"Aludra", 7.402, -29.303, 2.45},       // 4
			{"Furud", 6.338, -30.063, 3.02},        // 5
			{"Muliphein", 7.063, -15.633, 4.11},    // 6
			{"Omicron2 CMa", 7.048, -23.833, 3.02}, // 7
		},
		Lines: [][2]int{
			{1, 0}, {0, 6}, // nose and head
			{0, 7}, {7, 2}, // neck to the body
			{2, 4}, {4, 3}, {3, 2}, // hindquarters triangle
			{3, 5}, // rear leg
		},
	},
	{
		Latin: "Bootes", Abbr: "Boo", Common: "The Herdsman",
		Hemisphere: "Northern", BestMonth: "June",
		Stars: []Star{
			{"Arcturus", 14.261, 19.182, -0.05}, // 0
			{"Nekkar", 15.032, 40.390, 3.49},    // 1
			{"Seginus", 14.535, 38.308, 3.03},   // 2
			{"Izar", 14.750, 27.074, 2.35},      // 3
			{"Rho Boo", 14.535, 30.371, 3.58},   // 4
			{"Muphrid", 13.912, 18.398, 2.68},   // 5
			{"Delta Boo", 15.258, 33.315, 3.47}, // 6
		},
		Lines: [][2]int{
			{0, 3}, {3, 6}, {6, 1}, {1, 2}, {2, 4}, {4, 0}, // the kite
			{0, 5}, // to Muphrid
		},
	},
	{
		Latin: "Crux", Abbr: "Cru", Common: "The Southern Cross",
		Hemisphere: "Southern", BestMonth: "May",
		Stars: []Star{
			{"Acrux", 12.443, -63.099, 0.77},  // 0
			{"Mimosa", 12.795, -59.689, 1.25}, // 1
			{"Gacrux", 12.519, -57.113, 1.63}, // 2
			{"Imai", 12.253, -58.749, 2.79},   // 3
		},
		Lines: [][2]int{
			{0, 2}, // long axis
			{1, 3}, // cross axis
		},
	},
	{
		Latin: "Andromeda", Abbr: "And", Common: "The Chained Maiden",
		Hemisphere: "Northern", BestMonth: "November",
		Stars: []Star{
			{"Alpheratz", 0.140, 29.091, 2.06}, // 0
			{"Mirach", 1.162, 35.621, 2.05},    // 1
			{"Almach", 2.065, 42.330, 2.10},    // 2
			{"Delta And", 0.655, 30.861, 3.27}, // 3
			{"Nembus", 1.638, 48.628, 3.57},    // 4
			{"Mu And", 0.947, 38.499, 3.86},    // 5
		},
		Lines: [][2]int{
			{0, 3}, {3, 1}, {1, 2}, // the chain
			{1, 5}, {5, 4}, // branch
		},
	},
	{
		Latin: "Pegasus", Abbr: "Peg", Common: "The Winged Horse",
		Hemisphere: "Northern", BestMonth: "October",
		Stars: []Star{
			{"Markab", 23.079, 15.205, 2.48},   // 0
			{"Scheat", 23.063, 28.083, 2.42},   // 1
			{"Algenib", 0.221, 15.184, 2.83},   // 2
			{"Enif", 21.736, 9.875, 2.39},      // 3
			{"Homam", 22.691, 10.831, 3.40},    // 4
			{"Matar", 22.717, 30.221, 2.94},    // 5
			{"Baham", 22.170, 6.198, 3.53},     // 6
			{"Alpheratz", 0.140, 29.091, 2.06}, // 7 (shared corner)
		},
		Lines: [][2]int{
			{0, 1}, {1, 7}, {7, 2}, {2, 0}, // the Great Square
			{0, 4}, {4, 3}, // neck and head
			{4, 6}, // muzzle
			{1, 5}, // foreleg
		},
	},
	{
		Latin: "Perseus", Abbr: "Per", Common: "The Hero",
		Hemisphere: "Northern", BestMonth: "December",
		Stars: []Star{
			{"Mirfak", 3.405, 49.861, 1.79},      // 0
			{"Algol", 3.136, 40.956, 2.12},       // 1
			{"Gamma Per", 3.081, 53.506, 2.91},   // 2
			{"Delta Per", 3.715, 47.788, 3.01},   // 3
			{"Epsilon Per", 3.964, 40.010, 2.88}, // 4
			{"Zeta Per", 3.902, 31.884, 2.84},    // 5
			{"Rho Per", 3.086, 38.840, 3.39},     // 6
			{"Eta Per", 2.845, 55.895, 3.76},     // 7
		},
		Lines: [][2]int{
			{7, 2}, {2, 0}, // upper body
			{0, 3}, {3, 4}, {4, 5}, // right leg
			{0, 1}, {1, 6}, // left leg to Algol
		},
	},
	{
		Latin: "Auriga", Abbr: "Aur", Common: "The Charioteer",
		Hemisphere: "Northern", BestMonth: "January",
		Stars: []Star{
			{"Capella", 5.278, 45.998, 0.08},    // 0
			{"Menkalinan", 5.992, 44.947, 1.90}, // 1
			{"Mahasim", 5.995, 37.213, 2.62},    // 2
			{"Hassaleh", 4.950, 33.166, 2.69},   // 3
			{"Almaaz", 5.033, 43.823, 2.99},     // 4
			{"Elnath", 5.438, 28.607, 1.65},     // 5 (shared with Taurus)
		},
		Lines: [][2]int{
			{0, 1}, {1, 2}, {2, 5}, {5, 3}, {3, 4}, {4, 0}, // the pentagon
		},
	},
	{
		Latin: "Corona Borealis", Abbr: "CrB", Common: "The Northern Crown",
		Hemisphere: "Northern", BestMonth: "July",
		Stars: []Star{
			{"Alphecca", 15.578, 26.715, 2.22},    // 0
			{"Nusakan", 15.464, 29.106, 3.66},     // 1
			{"Theta CrB", 15.549, 31.359, 4.14},   // 2
			{"Gamma CrB", 15.712, 26.296, 3.80},   // 3
			{"Delta CrB", 15.827, 26.068, 4.59},   // 4
			{"Epsilon CrB", 15.958, 26.878, 4.13}, // 5
			{"Iota CrB", 16.024, 29.851, 4.98},    // 6
		},
		Lines: [][2]int{
			{2, 1}, {1, 0}, {0, 3}, {3, 4}, {4, 5}, {5, 6}, // the arc
		},
	},
	{
		Latin: "Cepheus", Abbr: "Cep", Common: "The King",
		Hemisphere: "Northern", BestMonth: "October",
		Stars: []Star{
			{"Alderamin", 21.309, 62.585, 2.45}, // 0
			{"Alfirk", 21.478, 70.561, 3.23},    // 1
			{"Errai", 23.656, 77.632, 3.21},     // 2 (peak)
			{"Zeta Cep", 22.181, 58.201, 3.35},  // 3
			{"Iota Cep", 22.828, 66.200, 3.52},  // 4
			{"Delta Cep", 22.477, 58.415, 4.07}, // 5
		},
		Lines: [][2]int{
			{0, 1}, {1, 2}, {2, 4}, {4, 1}, // roof to the peak
			{4, 3}, {3, 0}, // house body
		},
	},
	{
		Latin: "Draco", Abbr: "Dra", Common: "The Dragon",
		Hemisphere: "Northern", BestMonth: "July",
		Stars: []Star{
			{"Eltanin", 17.943, 51.489, 2.23},   // 0
			{"Rastaban", 17.507, 52.301, 2.79},  // 1
			{"Thuban", 14.073, 64.376, 3.65},    // 2
			{"Altais", 19.209, 67.661, 3.07},    // 3
			{"Nu Dra", 17.539, 55.185, 4.88},    // 4
			{"Xi Dra", 17.892, 56.873, 3.75},    // 5
			{"Zeta Dra", 17.146, 65.715, 3.17},  // 6
			{"Eta Dra", 16.400, 61.514, 2.74},   // 7
			{"Edasich", 15.415, 58.966, 3.29},   // 8
			{"Kappa Dra", 12.558, 69.788, 3.82}, // 9
		},
		Lines: [][2]int{
			{0, 1}, {1, 4}, {4, 5}, {5, 0}, // the head quadrilateral
			{0, 3}, {3, 6}, {6, 7}, {7, 8}, {8, 2}, {2, 9}, // winding body
		},
	},
	{
		Latin: "Cancer", Abbr: "Cnc", Common: "The Crab",
		Hemisphere: "Northern", BestMonth: "March",
		Stars: []Star{
			{"Altarf", 8.275, 9.186, 3.53},             // 0
			{"Asellus Australis", 8.745, 18.154, 3.94}, // 1
			{"Asellus Borealis", 8.721, 21.469, 4.66},  // 2
			{"Iota Cnc", 8.778, 28.759, 4.02},          // 3
			{"Acubens", 8.975, 11.858, 4.25},           // 4
		},
		Lines: [][2]int{
			{0, 1}, {1, 2}, {2, 3}, // the stem
			{1, 4}, // the branch
		},
	},
	{
		Latin: "Aries", Abbr: "Ari", Common: "The Ram",
		Hemisphere: "Northern", BestMonth: "December",
		Stars: []Star{
			{"Hamal", 2.119, 23.462, 2.00},     // 0
			{"Sheratan", 1.911, 20.808, 2.64},  // 1
			{"Mesarthim", 1.892, 19.294, 3.86}, // 2
			{"Bharani", 2.833, 27.261, 3.61},   // 3
		},
		Lines: [][2]int{
			{3, 0}, {0, 1}, {1, 2}, // the horn curve
		},
	},
}

// curated is the hand-curated part of the catalog, above; generated (in
// catalog_gen.go) holds the rest of the 88 IAU constellations.
var curated = catalog

func init() {
	catalog = append(append([]Constellation{}, curated...), generated...)
	sort.SliceStable(catalog, func(i, j int) bool { return catalog[i].Latin < catalog[j].Latin })
}

// Catalog returns all constellations in catalog order (alphabetical by
// Latin name).
func Catalog() []Constellation {
	out := make([]Constellation, len(catalog))
	copy(out, catalog)
	return out
}

// Names returns the Latin names of all constellations, sorted alphabetically.
func Names() []string {
	names := make([]string, 0, len(catalog))
	for _, c := range catalog {
		names = append(names, c.Latin)
	}
	sort.Strings(names)
	return names
}

// Lookup resolves a constellation by Latin name, common name, or IAU
// abbreviation, case-insensitively. Spaces and hyphens are treated as
// equivalent, so "ursa major", "Ursa-Major", and "URSAMAJOR" all match.
// The second return value reports whether a match was found.
func Lookup(query string) (Constellation, bool) {
	want := normalize(query)
	for _, c := range catalog {
		if normalize(c.Latin) == want ||
			normalize(c.Abbr) == want ||
			normalize(c.Common) == want {
			return c, true
		}
	}
	return Constellation{}, false
}

// CloseMatches returns Latin names whose normalized form contains the
// normalized query (or vice versa), for "did you mean" hints on a miss.
func CloseMatches(query string) []string {
	want := normalize(query)
	if want == "" {
		return nil
	}
	var out []string
	for _, c := range catalog {
		n := normalize(c.Latin)
		a := normalize(c.Abbr)
		if strings.Contains(n, want) || strings.Contains(want, n) ||
			strings.HasPrefix(a, want) || strings.HasPrefix(want, a) {
			out = append(out, c.Latin)
		}
	}
	sort.Strings(out)
	return out
}

// normalize lowercases s and strips spaces and hyphens so lookups are lenient.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r == ' ' || r == '-' || r == '_':
			// skip separators
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
