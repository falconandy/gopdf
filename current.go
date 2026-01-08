package gopdf

// Current current state
type Current struct {
	X float64
	Y float64

	CountOfFont int

	FontSize      float64
	FontStyle     int // Regular|Bold|Italic|Underline
	FontFontCount int

	CharSpacing float64

	FontISubset *SubsetFontObj // FontType == CURRENT_FONT_TYPE_SUBSET

	//text color mode
	txtColorMode string //color, gray

	//text color
	txtColor ICacheColorText

	lineWidth float64

	//current page size
	pageSize *Rect
}

func (c *Current) setTextColor(color ICacheColorText) {
	c.txtColor = color
}

func (c *Current) textColor() ICacheColorText {
	return c.txtColor
}
