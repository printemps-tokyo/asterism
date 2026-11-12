package sky

import (
	"fmt"
	"math"
	"time"
)

// Observer is a place on Earth and a moment, for working out where a
// constellation stands in the local sky.
type Observer struct {
	Lat  float64   // latitude, degrees, north positive
	Lon  float64   // longitude, degrees, east positive
	Time time.Time // the moment; its location is used to print local times
}

// siderealRate is sidereal days per solar day.
const siderealRate = 1.00273790935

// JulianDay returns the Julian Day of t (UTC). Seconds and nanoseconds are
// converted separately, since UnixNano only covers 1678 to 2262.
func JulianDay(t time.Time) float64 {
	return float64(t.Unix())/86400 + float64(t.Nanosecond())/86400e9 + 2440587.5
}

// MinYear and MaxYear bound the moments asterism accepts: its J2000
// positions ignore precession, which moves stars about 1.4 degrees a century.
const (
	MinYear = 1900
	MaxYear = 2100
)

// GMST returns Greenwich mean sidereal time at t, in degrees [0, 360), from
// the IAU 1982 expression as given by Meeus, Astronomical Algorithms (12.4).
func GMST(t time.Time) float64 {
	jd := JulianDay(t)
	T := (jd - 2451545.0) / 36525
	theta := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*T*T - T*T*T/38710000
	return math.Mod(math.Mod(theta, 360)+360, 360)
}

// AltAz returns the altitude and azimuth (degrees; azimuth from north through
// east) of a J2000 position for an observer. Precession since J2000, nutation,
// and refraction are ignored: about half a degree off now and up to about 1.5
// degrees at the ends of MinYear..MaxYear, plenty for "where do I look".
func AltAz(raHours, dec float64, o Observer) (alt, az float64) {
	v := horizonVector(raHours, dec, o)
	alt = deg(math.Asin(math.Max(-1, math.Min(1, v[2]))))
	az = math.Mod(deg(math.Atan2(v[0], v[1]))+360, 360)
	return alt, az
}

// horizonVector is the unit vector of a J2000 position in the observer's
// east-north-up frame.
func horizonVector(raHours, dec float64, o Observer) [3]float64 {
	lst := GMST(o.Time) + o.Lon
	h := rad(lst - raHours*15)
	d, phi := rad(dec), rad(o.Lat)
	return [3]float64{
		-math.Cos(d) * math.Sin(h),
		math.Sin(d)*math.Cos(phi) - math.Cos(d)*math.Cos(h)*math.Sin(phi),
		math.Sin(d)*math.Sin(phi) + math.Cos(d)*math.Cos(h)*math.Cos(phi),
	}
}

// Center returns the direction of the figure's center: the mean of its stars'
// unit vectors, so figures that straddle RA 0h average correctly.
func (c Constellation) Center() (raHours, dec float64) {
	var x, y, z float64
	for _, s := range c.Stars {
		a, d := rad(s.RA*15), rad(s.Dec)
		x += math.Cos(d) * math.Cos(a)
		y += math.Cos(d) * math.Sin(a)
		z += math.Sin(d)
	}
	ra := math.Mod(deg(math.Atan2(y, x))/15+24, 24)
	return ra, deg(math.Atan2(z, math.Hypot(x, y)))
}

// Visibility is where a constellation stands for an observer.
type Visibility struct {
	Alt, Az     float64   // the center's altitude and azimuth, degrees
	Above       int       // stars above the horizon
	Total       int       // stars in the figure
	Transit     time.Time // the center's meridian crossing nearest to the moment
	Rise, Set   time.Time // around that transit; zero when Circumpolar or NeverRises
	Circumpolar bool      // the center never sets
	NeverRises  bool      // the center never rises
}

// See works out the constellation's visibility for an observer. Rise and set
// are for the center at altitude 0, without refraction.
func See(c Constellation, o Observer) Visibility {
	v := Visibility{Total: len(c.Stars)}
	for _, s := range c.Stars {
		if alt, _ := AltAz(s.RA, s.Dec, o); alt > 0 {
			v.Above++
		}
	}
	ra, dec := c.Center()
	v.Alt, v.Az = AltAz(ra, dec, o)

	lstHours := (GMST(o.Time) + o.Lon) / 15
	untilTransit := math.Mod(ra-lstHours, 24)
	switch {
	case untilTransit > 12:
		untilTransit -= 24
	case untilTransit <= -12:
		untilTransit += 24
	}
	v.Transit = o.Time.Add(hoursToDuration(untilTransit / siderealRate))

	cosH0 := -math.Tan(rad(o.Lat)) * math.Tan(rad(dec))
	switch {
	case cosH0 < -1:
		v.Circumpolar = true
	case cosH0 > 1:
		v.NeverRises = true
	default:
		half := hoursToDuration(deg(math.Acos(cosH0)) / 15 / siderealRate)
		v.Rise = v.Transit.Add(-half)
		v.Set = v.Transit.Add(half)
	}
	return v
}

// HorizonView returns a copy of c whose star coordinates are replaced so that
// rendering it shows the figure as it stands in the sky: the observer faces
// the figure's center, the zenith is up, and the right hand is to the right.
// The sky is rotated so the center lands at RA 0h, Dec 0, with "up" toward the
// zenith; this stays seamless for figures at the zenith or across north.
func HorizonView(c Constellation, o Observer) Constellation {
	ra, dec := c.Center()
	center := horizonVector(ra, dec, o)
	// Up: the zenith direction with its component along the center removed.
	// At the zenith itself, use north.
	up := sub([3]float64{0, 0, 1}, scale(center, center[2]))
	if norm(up) < 1e-6 {
		up = sub([3]float64{0, 1, 0}, scale(center, center[1]))
	}
	up = scale(up, 1/norm(up))
	right := cross(center, up) // facing the center with up overhead

	out := c
	out.Stars = make([]Star, len(c.Stars))
	for i, s := range c.Stars {
		v := horizonVector(s.RA, s.Dec, o)
		x, y, z := dot(v, right), dot(v, up), dot(v, center)
		lon := deg(math.Atan2(x, z)) // positive to the right
		lat := deg(math.Asin(math.Max(-1, math.Min(1, y))))
		// ProjectStars puts larger RA to the left, so rightward is -RA.
		out.Stars[i] = Star{Name: s.Name, RA: math.Mod(-lon/15+24, 24), Dec: lat, Mag: s.Mag}
	}
	return out
}

func dot(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func norm(a [3]float64) float64   { return math.Sqrt(dot(a, a)) }
func scale(a [3]float64, k float64) [3]float64 {
	return [3]float64{a[0] * k, a[1] * k, a[2] * k}
}
func sub(a, b [3]float64) [3]float64 { return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func cross(a, b [3]float64) [3]float64 {
	return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

// Compass names an azimuth on a 16-point compass rose.
func Compass(az float64) string {
	points := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	i := int(math.Floor(math.Mod(az+11.25+360, 360) / 22.5))
	return points[i%16]
}

// ObserverLines describes a constellation's place in the observer's sky, for
// the info panel.
func ObserverLines(c Constellation, o Observer) []string {
	v := See(c, o)
	where := fmt.Sprintf("%s at %s", FormatLatLon(o.Lat, o.Lon), o.Time.Format("2006-01-02 15:04 MST"))
	var state string
	switch {
	case v.Alt > 0:
		state = fmt.Sprintf("up, %.0f° high in the %s (azimuth %.0f°)", v.Alt, Compass(v.Az), v.Az)
	default:
		state = fmt.Sprintf("below the horizon (%.0f°, %s)", v.Alt, Compass(v.Az))
	}
	lines := []string{
		"Seen from:      " + where,
		"Center:         " + state,
		fmt.Sprintf("Stars up:       %d of %d", v.Above, v.Total),
	}
	clock := func(t time.Time) string { return t.In(o.Time.Location()).Format("15:04") }
	switch {
	case v.Circumpolar:
		lines = append(lines, "Rise/set:       never sets here (circumpolar); highest at "+clock(v.Transit))
	case v.NeverRises:
		lines = append(lines, "Rise/set:       never rises here")
	default:
		lines = append(lines, fmt.Sprintf("Rise/set:       rises %s, highest %s, sets %s", clock(v.Rise), clock(v.Transit), clock(v.Set)))
	}
	return lines
}

// FormatLatLon renders a position like "35.68N 139.77E".
func FormatLatLon(lat, lon float64) string {
	ns, ew := "N", "E"
	if lat < 0 {
		ns, lat = "S", -lat
	}
	if lon < 0 {
		ew, lon = "W", -lon
	}
	return fmt.Sprintf("%.2f%s %.2f%s", lat, ns, lon, ew)
}

func hoursToDuration(h float64) time.Duration {
	return time.Duration(h * float64(time.Hour))
}

func rad(d float64) float64 { return d * math.Pi / 180 }
func deg(r float64) float64 { return r * 180 / math.Pi }
