package gopdf

// Rect defines a rectangle.
type Rect struct {
	W            float64
	H            float64
	unitOverride defaultUnitConfig
}

// UnitsToPoints converts the rectanlges width and height to Points. When this is called it is assumed the values of the rectangle are in Units
func (rect *Rect) UnitsToPoints(t int) (r *Rect) {
	if rect == nil {
		return
	}

	unitCfg := defaultUnitConfig{Unit: t}
	if rect.unitOverride.getUnit() != UnitUnset {
		unitCfg = rect.unitOverride
	}

	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unitCfg, &r.W, &r.H)
	return
}

func (rect *Rect) unitsToPoints(unitCfg unitConfigurator) (r *Rect) {
	if rect == nil {
		return
	}
	if rect.unitOverride.getUnit() != UnitUnset {
		unitCfg = rect.unitOverride
	}
	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unitCfg, &r.W, &r.H)
	return
}
