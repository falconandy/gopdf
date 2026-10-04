package gopdf

// Rect defines a rectangle.
type Rect struct {
	W    float64
	H    float64
	unit Unit
}

// UnitsToPoints converts the rectanlges width and height to Points. When this is called it is assumed the values of the rectangle are in Units
func (rect *Rect) UnitsToPoints(t int) (r *Rect) {
	if rect == nil {
		return
	}

	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(rect.unit, &r.W, &r.H)
	return
}

func (rect *Rect) unitsToPoints(unit Unit) (r *Rect) {
	if rect == nil {
		return
	}
	if rect.unit != UnitUnset {
		unit = rect.unit
	}
	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unit, &r.W, &r.H)
	return
}
