package gopdf

func convertTypoUnit(val float64, unitsPerEm uint, fontSize float64) float64 {
	val = val * 1000.00 / float64(unitsPerEm)
	return val * fontSize / 1000.0
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
