package gopdf

// Current current state
type Current struct {
	FontSize    float64
	CharSpacing float64
	FontISubset *SubsetFontObj // FontType == CURRENT_FONT_TYPE_SUBSET
}
