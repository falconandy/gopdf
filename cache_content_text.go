package gopdf

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// ContentTypeCell cell
const ContentTypeCell = 0

// ContentTypeText text
const ContentTypeText = 1

var ErrContentTypeNotFound = errors.New("contentType not found")

type cacheContentText struct {
	//---setup---
	rectangle      *Rect
	textColor      ICacheColorText
	txtColorMode   string
	fontCountIndex int //Curr.FontFontCount+1
	fontSize       float64
	fontStyle      int
	charSpacing    float64
	x, y           float64
	fontSubset     *SubsetFontObj
	pageheight     float64
	contentType    int
	cellOpt        CellOption
	lineWidth      float64
	text           string
	//---result---
	cellWidthPdfUnit, textWidthPdfUnit float64
	cellHeightPdfUnit                  float64
	isPlaceHolder                      bool
}

func (c *cacheContentText) isSame(cache cacheContentText) bool {
	if c.rectangle != nil {
		//if rectangle != nil we assume this is not same content
		return false
	}

	// if both colors are nil we assume them equal
	if ((c.textColor == nil && cache.textColor == nil) ||
		(c.textColor != nil && c.textColor.equal(cache.textColor))) &&
		c.fontCountIndex == cache.fontCountIndex &&
		c.fontSize == cache.fontSize &&
		c.fontStyle == cache.fontStyle &&
		c.charSpacing == cache.charSpacing &&
		c.y == cache.y &&
		c.isPlaceHolder == cache.isPlaceHolder {
		return true
	}

	return false
}

func (c *cacheContentText) setPageHeight(pageheight float64) {
	c.pageheight = pageheight
}

func (c *cacheContentText) pageHeight() float64 {
	return c.pageheight //841.89
}

func convertTypoUnit(val float64, unitsPerEm uint, fontSize float64) float64 {
	val = val * 1000.00 / float64(unitsPerEm)
	return val * fontSize / 1000.0
}

func (c *cacheContentText) calTypoAscender() float64 {
	return convertTypoUnit(float64(c.fontSubset.ttfp.TypoAscender()), c.fontSubset.ttfp.UnitsPerEm(), float64(c.fontSize))
}

func (c *cacheContentText) calTypoDescender() float64 {
	return convertTypoUnit(float64(c.fontSubset.ttfp.TypoDescender()), c.fontSubset.ttfp.UnitsPerEm(), float64(c.fontSize))
}

func (c *cacheContentText) calY() (float64, error) {
	pageHeight := c.pageHeight()
	if c.contentType == ContentTypeText {
		return pageHeight - c.y, nil
	} else if c.contentType == ContentTypeCell {
		y := float64(0.0)
		if c.cellOpt.Align&Bottom == Bottom {
			y = pageHeight - c.y - c.cellHeightPdfUnit - c.calTypoDescender()
		} else if c.cellOpt.Align&Middle == Middle {
			y = pageHeight - c.y - c.cellHeightPdfUnit*0.5 - (c.calTypoDescender()+c.calTypoAscender())*0.5
		} else {
			//top
			y = pageHeight - c.y - c.calTypoAscender()
		}

		return y, nil
	}
	return 0.0, ErrContentTypeNotFound
}

func (c *cacheContentText) calX() (float64, error) {
	if c.contentType == ContentTypeText {
		return c.x, nil
	} else if c.contentType == ContentTypeCell {
		x := float64(0.0)
		if c.cellOpt.Align&Right == Right {
			x = c.x + c.cellWidthPdfUnit - c.textWidthPdfUnit
		} else if c.cellOpt.Align&Center == Center {
			x = c.x + c.cellWidthPdfUnit*0.5 - c.textWidthPdfUnit*0.5
		} else {
			x = c.x
		}
		return x, nil
	}
	return 0.0, ErrContentTypeNotFound
}

// FormatFloatTrim converts a float64 into a string, like Sprintf("%.3f")
// but with trailing zeroes (and possibly ".") removed
func FormatFloatTrim(floatval float64) (formatted string) {
	const precisionFactor = 1000.0
	roundedFontSize := math.Round(precisionFactor*floatval) / precisionFactor
	return strconv.FormatFloat(roundedFontSize, 'f', -1, 64)
}

func AppendFloatTrim(dst []byte, floatval float64) []byte {
	const precisionFactor = 1000.0
	roundedFontSize := math.Round(precisionFactor*floatval) / precisionFactor
	return strconv.AppendFloat(dst, roundedFontSize, 'f', -1, 64)
}

var floatBuf = make([]byte, 0, 24)

func (c *cacheContentText) write(w io.Writer) error {
	x, err := c.calX()
	if err != nil {
		return err
	}
	y, err := c.calY()
	if err != nil {
		return err
	}

	if _, err := io.WriteString(w, "BT\n"); err != nil {
		return err
	}

	fmt.Fprintf(w, "%0.2f %0.2f TD\n", x, y)
	fmt.Fprintf(w, "/F%d ", c.fontCountIndex)
	w.Write(AppendFloatTrim(floatBuf, c.fontSize))
	fmt.Fprint(w, " Tf ")
	w.Write(AppendFloatTrim(floatBuf, c.charSpacing))
	fmt.Fprint(w, " Tc\n")

	if c.txtColorMode == "color" {
		c.textColor.write(w)
	}
	io.WriteString(w, "[<")

	for i, r := range c.text {

		glyphindex, err := c.fontSubset.CharIndex(r)
		if err == ErrCharNotFound {
			continue
		} else if err != nil {
			return err
		}

		if i > 0 && c.fontSubset.ttfFontOption.UseKerning { //kerning
			// FIXME
		}

		fmt.Fprintf(w, "%04X", glyphindex)
	}

	io.WriteString(w, ">] TJ\n")
	io.WriteString(w, "ET\n")

	if c.fontStyle&Underline == Underline {
		// FIXME
		//if err := c.underline(w); err != nil {
		//	return err
		//}
	}

	c.drawBorder(w)

	return nil
}

func (c *cacheContentText) drawBorder(w io.Writer) error {

	//stream.WriteString(fmt.Sprintf("%.2f w\n", 0.1))
	lineOffset := c.lineWidth * 0.5

	if c.cellOpt.Border&Top == Top {

		startX := c.x - lineOffset
		startY := c.pageHeight() - c.y
		endX := c.x + c.cellWidthPdfUnit + lineOffset
		endY := startY
		_, err := fmt.Fprintf(w, "%0.2f %0.2f m %0.2f %0.2f l s\n", startX, startY, endX, endY)
		if err != nil {
			return err
		}
	}

	if c.cellOpt.Border&Left == Left {
		startX := c.x
		startY := c.pageHeight() - c.y
		endX := c.x
		endY := startY - c.cellHeightPdfUnit
		_, err := fmt.Fprintf(w, "%0.2f %0.2f m %0.2f %0.2f l s\n", startX, startY, endX, endY)
		if err != nil {
			return err
		}
	}

	if c.cellOpt.Border&Right == Right {
		startX := c.x + c.cellWidthPdfUnit
		startY := c.pageHeight() - c.y
		endX := c.x + c.cellWidthPdfUnit
		endY := startY - c.cellHeightPdfUnit
		_, err := fmt.Fprintf(w, "%0.2f %0.2f m %0.2f %0.2f l s\n", startX, startY, endX, endY)
		if err != nil {
			return err
		}
	}

	if c.cellOpt.Border&Bottom == Bottom {
		startX := c.x - lineOffset
		startY := c.pageHeight() - c.y - c.cellHeightPdfUnit
		endX := c.x + c.cellWidthPdfUnit + lineOffset
		endY := startY
		_, err := fmt.Fprintf(w, "%0.2f %0.2f m %0.2f %0.2f l s\n", startX, startY, endX, endY)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *cacheContentText) createContent() (float64, float64, error) {

	cellWidthPdfUnit, cellHeightPdfUnit, textWidthPdfUnit, err := createContent(c.fontSubset, c.text, c.fontSize, c.charSpacing, c.rectangle)
	if err != nil {
		return 0, 0, err
	}
	c.cellWidthPdfUnit = cellWidthPdfUnit
	c.cellHeightPdfUnit = cellHeightPdfUnit
	c.textWidthPdfUnit = textWidthPdfUnit
	return cellWidthPdfUnit, cellHeightPdfUnit, nil
}

func createContent(f *SubsetFontObj, text string, fontSize float64, charSpacing float64, rectangle *Rect) (float64, float64, float64, error) {

	unitsPerEm := int(f.ttfp.UnitsPerEm())
	sumWidth := int(0)
	//fmt.Printf("unitsPerEm = %d", unitsPerEm)
	for i, r := range text {

		glyphindex, err := f.CharIndex(r)
		if err == ErrCharNotFound {
			continue
		} else if err != nil {
			return 0, 0, 0, err
		}

		pairvalPdfUnit := 0
		if i > 0 && f.ttfFontOption.UseKerning { //kerning
			// FIXME
		}

		width := f.GlyphIndexToPdfWidth(glyphindex)

		unitsPerPt := float64(unitsPerEm) / fontSize
		spaceWidthInPt := unitsPerPt * charSpacing
		spaceWidthPdfUnit := convertTTFUnit2PDFUnit(int(spaceWidthInPt), unitsPerEm)

		sumWidth += int(width) + int(pairvalPdfUnit) + spaceWidthPdfUnit
	}

	cellWidthPdfUnit := float64(0)
	cellHeightPdfUnit := float64(0)
	if rectangle == nil {
		cellWidthPdfUnit = float64(sumWidth) * (float64(fontSize) / 1000.0)
		typoAscender := convertTypoUnit(float64(f.ttfp.TypoAscender()), f.ttfp.UnitsPerEm(), float64(fontSize))
		typoDescender := convertTypoUnit(float64(f.ttfp.TypoDescender()), f.ttfp.UnitsPerEm(), float64(fontSize))
		cellHeightPdfUnit = typoAscender - typoDescender
	} else {
		cellWidthPdfUnit = rectangle.W
		cellHeightPdfUnit = rectangle.H
	}
	textWidthPdfUnit := float64(sumWidth) * (float64(fontSize) / 1000.0)
	return cellWidthPdfUnit, cellHeightPdfUnit, textWidthPdfUnit, nil
}

func createNextRuneContent(f *SubsetFontObj, r, leftRune rune, fontSize float64, charSpacing float64) (float64, error) {

	unitsPerEm := int(f.ttfp.UnitsPerEm())
	//fmt.Printf("unitsPerEm = %d", unitsPerEm)

	glyphindex, err := f.CharIndex(r)
	if err == ErrCharNotFound {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	pairvalPdfUnit := 0
	if leftRune != 0 && f.ttfFontOption.UseKerning { //kerning
		// FIXME
	}

	width := f.GlyphIndexToPdfWidth(glyphindex)

	unitsPerPt := float64(unitsPerEm) / fontSize
	spaceWidthInPt := unitsPerPt * charSpacing
	spaceWidthPdfUnit := convertTTFUnit2PDFUnit(int(spaceWidthInPt), unitsPerEm)

	sumWidth := int(width) + pairvalPdfUnit + spaceWidthPdfUnit

	textWidthPdfUnit := float64(sumWidth) * (float64(fontSize) / 1000.0)
	return textWidthPdfUnit, nil
}
