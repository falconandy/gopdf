package gopdf

// The units that can be used in the document
const (
	UnitMM = 1 // Millimeters

	// The math needed to convert units to points
	conversionUnitMM = 72.0 / 25.4
)

// Config static config
type Config struct {
	Unit     int  // The unit type to use when composing the document.
	PageSize Rect // The default page size for all pages in the document
}

func (c Config) getUnit() int {
	return c.Unit
}

func unitsToPoints(unit int, u float64) float64 {
	switch unit {
	case UnitMM:
		return u * conversionUnitMM
	default:
		return u
	}
}

func unitsToPointsVar(unit int, u ...*float64) {
	for x := 0; x < len(u); x++ {
		*u[x] = unitsToPoints(unit, *u[x])
	}
}
