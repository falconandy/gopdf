package gopdf

// Rect defines a rectangle.
type Rect struct {
	W float64
	H float64
}

// UnitsToPoints converts the rectanlges width and height to Points. When this is called it is assumed the values of the rectangle are in Units
func (rect *Rect) UnitsToPoints(unit int) (r *Rect) {
	if rect == nil {
		return
	}

	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unit, &r.W, &r.H)
	return
}
