package gopdf

type Unit uint8

// The units that can be used in the document
const (
	UnitUnset Unit = iota // No units were set, when conversion is called on nothing will happen
	UnitMM                // Millimeters

	// The math needed to convert units to points
	conversionUnitMM = 72.0 / 25.4
)

// Config static config
type Config struct {
	PageSize Rect // The default page size for all pages in the document
	Unit     Unit
}

// UnitsToPoints converts units of the provided type to points
func UnitsToPoints(unit Unit, u float64) float64 {
	return unitsToPoints(unit, u)
}

func unitsToPoints(unit Unit, u float64) float64 {
	switch unit {
	case UnitMM:
		return u * conversionUnitMM
	default:
		return u
	}
}

// PointsToUnits converts points to the provided units
func PointsToUnits(unit Unit, u float64) float64 {
	return pointsToUnits(unit, u)
}

func pointsToUnits(unit Unit, u float64) float64 {
	switch unit {
	case UnitMM:
		return u / conversionUnitMM
	default:
		return u
	}
}

// UnitsToPointsVar converts units of the provided type to points for all variables supplied
func UnitsToPointsVar(unit Unit, u ...*float64) {
	unitsToPointsVar(unit, u...)
}

func unitsToPointsVar(unit Unit, u ...*float64) {
	for x := 0; x < len(u); x++ {
		*u[x] = unitsToPoints(unit, *u[x])
	}
}

// PointsToUnitsVar converts points to the provided units for all variables supplied
func PointsToUnitsVar(unit Unit, u ...*float64) {
	pointsToUnitsVar(unit, u...)
}

func pointsToUnitsVar(unit Unit, u ...*float64) {
	for x := 0; x < len(u); x++ {
		*u[x] = pointsToUnits(unit, *u[x])
	}
}
