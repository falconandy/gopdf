package gopdf

import (
	"fmt"
	"io"
)

// ContentObj content object
type ContentObj struct { //impl IObj
	listCache listCacheContent
	//text bytes.Buffer
	getRoot func() *GoPdf
}

func (c *ContentObj) init(funcGetRoot func() *GoPdf) {
	c.getRoot = funcGetRoot
}

func (c *ContentObj) write(w io.Writer, objID int) error {
	buff := GetBuffer()
	defer PutBuffer(buff)

	if err := c.listCache.write(buff); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "<<\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "/Length %d\n", buff.Len()); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ">>\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "stream\n"); err != nil {
		return err
	}

	if _, err := buff.WriteTo(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "endstream\n"); err != nil {
		return err
	}

	return nil
}

func (c *ContentObj) getType() string {
	return "Content"
}

// AppendStreamSubsetFont add stream of text
func (c *ContentObj) AppendStreamSubsetFont(rectangle *Rect, text string, cellOpt CellOption) error {

	textColor := c.getRoot().curr.textColor()
	fontCountIndex := c.getRoot().curr.FontFontCount + 1
	fontSize := c.getRoot().curr.FontSize
	fontStyle := c.getRoot().curr.FontStyle
	charSpacing := c.getRoot().curr.CharSpacing
	x := c.getRoot().curr.X
	y := c.getRoot().curr.Y
	fontSubset := c.getRoot().curr.FontISubset

	cache := cacheContentText{
		fontSubset:     fontSubset,
		rectangle:      rectangle,
		textColor:      textColor,
		fontCountIndex: fontCountIndex,
		fontSize:       fontSize,
		fontStyle:      fontStyle,
		charSpacing:    charSpacing,
		x:              x,
		y:              y,
		pageheight:     c.getRoot().curr.pageSize.H,
		contentType:    ContentTypeCell,
		cellOpt:        cellOpt,
		lineWidth:      c.getRoot().curr.lineWidth,
		txtColorMode:   c.getRoot().curr.txtColorMode,
	}
	var err error
	c.getRoot().curr.X, c.getRoot().curr.Y, err = c.listCache.appendContentText(cache, text)
	if err != nil {
		return err
	}
	return nil
}

// AppendStreamLine append line
func (c *ContentObj) AppendStreamLine(x1 float64, y1 float64, x2 float64, y2 float64) {
	//h := c.getRoot().config.PageSize.H
	//c.stream.WriteString(fmt.Sprintf("%0.2f %0.2f m %0.2f %0.2f l s\n", x1, h-y1, x2, h-y2))
	var cache cacheContentLine
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x1 = x1
	cache.y1 = y1
	cache.x2 = x2
	cache.y2 = y2
	c.listCache.append(&cache)
}

// AppendStreamSetLineWidth : set line width
func (c *ContentObj) AppendStreamSetLineWidth(w float64) {
	var cache cacheContentLineWidth
	cache.width = w
	c.listCache.append(&cache)
}

// AppendStreamSetLineType : Set linetype [solid, dashed, dotted]
func (c *ContentObj) AppendStreamSetLineType(t string) {
	var cache cacheContentLineType
	cache.lineType = t
	c.listCache.append(&cache)
}

// AppendStreamSetColorStroke  set the color stroke
func (c *ContentObj) AppendStreamSetColorStroke(r uint8, g uint8, b uint8) {
	var cache cacheContentColorRGB
	cache.colorType = colorTypeStrokeRGB
	cache.r = r
	cache.g = g
	cache.b = b
	c.listCache.append(&cache)
}

// AppendStreamSetColorFill  set the color fill
func (c *ContentObj) AppendStreamSetColorFill(r uint8, g uint8, b uint8) {
	var cache cacheContentColorRGB
	cache.colorType = colorTypeFillRGB
	cache.r = r
	cache.g = g
	cache.b = b
	c.listCache.append(&cache)
}

// ContentObjCalTextHeight : calculates height of text.
func ContentObjCalTextHeight(fontsize int) float64 {
	return ContentObjCalTextHeightPrecise(float64(fontsize))
}

// ContentObjCalTextHeightPrecise : like ContentObjCalTextHeight,
// but fontsize float64
func ContentObjCalTextHeightPrecise(fontsize float64) float64 {
	return (float64(fontsize) * 0.7)
}

// When setting colour and grayscales the value has to be between 0.00 and 1.00
// This function takes a float64 and returns 0.0 if it is less than 0.0 and 1.0 if it
// is more than 1.0
func fixRange10(val float64) float64 {
	if val < 0.0 {
		return 0.0
	}
	if val > 1.0 {
		return 1.0
	}
	return val
}

func convertTTFUnit2PDFUnit(n int, upem int) int {
	var ret int
	if n < 0 {
		rest1 := n % upem
		storrest := 1000 * rest1
		//ledd2 := (storrest != 0 ? rest1 / storrest : 0);
		ledd2 := 0
		if storrest != 0 {
			ledd2 = rest1 / storrest
		} else {
			ledd2 = 0
		}
		ret = -((-1000*n)/upem - int(ledd2))
	} else {
		ret = (n/upem)*1000 + ((n%upem)*1000)/upem
	}
	return ret
}
