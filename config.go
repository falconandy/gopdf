package gopdf

// The units that can be used in the document
const (
	UnitUnset = iota // No units were set, when conversion is called on nothing will happen
	UnitMM           // Millimeters

	// The math needed to convert units to points
	conversionUnitMM = 72.0 / 25.4
)

// Config static config
type Config struct {
	Unit int // The unit type to use when composing the document.
	//Value that use to convert units to points.
	//If this variable is not 0. This value will be used to calculate the unit conversion instead of the existing const value in the system.
	//And if this variable is not 0. Value ​​in Config.Unit will not be used.
	PageSize Rect // The default page size for all pages in the document
}

func (c Config) getUnit() int {
	return c.Unit
}

// UnitsToPoints converts units of the provided type to points
func UnitsToPoints(t int, u float64) float64 {
	return unitsToPoints(defaultUnitConfig{Unit: t}, u)
}

func unitsToPoints(unitCfg unitConfigurator, u float64) float64 {
	switch unitCfg.getUnit() {
	case UnitMM:
		return u * conversionUnitMM
	default:
		return u
	}
}

// PointsToUnits converts points to the provided units
func PointsToUnits(t int, u float64) float64 {
	return pointsToUnits(defaultUnitConfig{Unit: t}, u)
}

func pointsToUnits(unitCfg unitConfigurator, u float64) float64 {
	switch unitCfg.getUnit() {
	case UnitMM:
		return u / conversionUnitMM
	default:
		return u
	}
}

// UnitsToPointsVar converts units of the provided type to points for all variables supplied
func UnitsToPointsVar(t int, u ...*float64) {
	unitsToPointsVar(defaultUnitConfig{Unit: t}, u...)
}

func unitsToPointsVar(unitCfg unitConfigurator, u ...*float64) {
	for x := 0; x < len(u); x++ {
		*u[x] = unitsToPoints(unitCfg, *u[x])
	}
}

// PointsToUnitsVar converts points to the provided units for all variables supplied
func PointsToUnitsVar(t int, u ...*float64) {
	pointsToUnitsVar(defaultUnitConfig{Unit: t}, u...)
}

func pointsToUnitsVar(unitCfg unitConfigurator, u ...*float64) {
	for x := 0; x < len(u); x++ {
		*u[x] = pointsToUnits(unitCfg, *u[x])
	}
}

type unitConfigurator interface {
	getUnit() int
}

type defaultUnitConfig struct {
	Unit              int
	ConversionForUnit float64
}

func (d defaultUnitConfig) getUnit() int {
	return d.Unit
}
