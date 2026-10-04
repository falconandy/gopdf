package gopdf

import (
	"errors"
	"fmt"
)

const subsetFont = "SubsetFont"

var ErrEmptyString = errors.New("empty string")

var ErrMissingFontFamily = errors.New("font family not found")

// GoPdf : A simple library for generating PDF written in Go lang
type GoPdf struct {
	pdfObjs []*SubsetFontObj
	config  Config

	//ต่ำแหน่งปัจจุบัน
	curr Current
}

// Start : init gopdf
func (gp *GoPdf) Start(config Config) {

	gp.start(config)

}

//func (gp *GoPdf) StartWithImporter(config Config, importer *gofpdi.Importer) {
//
//	gp.start(config, importer)
//
//}

func (gp *GoPdf) start(config Config) {
	gp.config = config
	gp.init()
}

// convertNumericToFloat64 : accept numeric types, return float64-value
func convertNumericToFloat64(size interface{}) (fontSize float64, err error) {
	switch size := size.(type) {
	case float32:
		return float64(size), nil
	case float64:
		return float64(size), nil
	case int:
		return float64(size), nil
	case int16:
		return float64(size), nil
	case int32:
		return float64(size), nil
	case int64:
		return float64(size), nil
	case int8:
		return float64(size), nil
	case uint:
		return float64(size), nil
	case uint16:
		return float64(size), nil
	case uint32:
		return float64(size), nil
	case uint64:
		return float64(size), nil
	case uint8:
		return float64(size), nil
	default:
		return 0.0, fmt.Errorf("fontSize must be of type (u)int* or float*, not %T", size)
	}
}

// SetFontWithStyle : set font style support Regular or Underline
// for Bold|Italic should be loaded appropriate fonts with same styles defined
// size MUST be uint*, int* or float64*
func (gp *GoPdf) SetFontWithStyle(family string, style int, size interface{}) error {
	fontSize, err := convertNumericToFloat64(size)
	if err != nil {
		return err
	}
	found := false
	i := 0
	max := len(gp.pdfObjs)
	for i < max {
		sub := gp.pdfObjs[i]
		if sub.GetFamily() == family && sub.GetTtfFontOption().Style == style&^Underline {
			gp.curr.FontSize = fontSize
			gp.curr.FontISubset = sub
			found = true
			break
		}
		i++
	}

	if !found {
		return ErrMissingFontFamily
	}

	return nil
}

// SetFont : set font style support "" or "U"
// for "B" and "I" should be loaded appropriate fonts with same styles defined
// size MUST be uint*, int* or float64*
func (gp *GoPdf) SetFont(family string, style string, size interface{}) error {
	return gp.SetFontWithStyle(family, getConvertedStyle(style), size)
}

// SetFontSize : set the font size (and only the font size) of the currently
// active font
func (gp *GoPdf) SetFontSize(fontSize float64) error {
	gp.curr.FontSize = fontSize
	return nil
}

// SetCharSpacing : set the character spacing of the currently active font
func (gp *GoPdf) SetCharSpacing(charSpacing float64) error {
	gp.UnitsToPointsVar(&charSpacing)
	gp.curr.CharSpacing = charSpacing
	return nil
}

// SplitTextWithWordWrap behaves the same way SplitText does but performs a word-wrap considering spaces in case
// a text line split would split a word.
func (gp *GoPdf) SplitTextWithWordWrap(text string, width float64) ([]string, error) {
	return gp.SplitTextWithOption(text, width, &BreakOption{
		Mode:           BreakModeIndicatorSensitive,
		BreakIndicator: ' ',
	})
}

// SplitTextWithOption splits a text into multiple lines based on the current font size of the document.
// BreakOptions allow to define the behavior of the split (strict or sensitive). For more information see BreakOption.
func (gp *GoPdf) SplitTextWithOption(text string, width float64, opt *BreakOption) ([]string, error) {
	// fallback to default break option
	if opt == nil {
		opt = &DefaultBreakOption
	}
	utf8Texts := []rune(text)
	utf8TextsLen := len(utf8Texts) // utf8 string quantity
	if utf8TextsLen == 0 {
		return nil, ErrEmptyString
	}
	lineText := make([]rune, 0, len(utf8Texts))
	lineTexts := make([]string, 0, 3)
	separatorWidth, err := gp.MeasureTextWidth(opt.Separator)
	if err != nil {
		return nil, err
	}
	// possible (not conflicting) position of the separator within the currently processed line
	separatorIdx := 0
	var lineWidth float64
	for i := 0; i < utf8TextsLen; i++ {
		var leftRune rune
		if len(lineText) > 0 {
			leftRune = lineText[len(lineText)-1]
		}
		runeWidth, err := gp.MeasureNextRuneWidth(utf8Texts[i], leftRune)
		if err != nil {
			return nil, err
		}
		// mid-word break required since the max width of the given rect is exceeded
		if lineWidth+runeWidth > width && utf8Texts[i] != '\n' {
			// forceBreak will be set to true in case an indicator sensitive break was not possible which will cause
			// strict break to not exceed the desired width
			forceBreak := false
			if opt.Mode == BreakModeIndicatorSensitive {
				forceBreak = !performIndicatorSensitiveLineBreak(&lineTexts, &lineText, &i, &lineWidth, opt)
			}
			// BreakModeStrict breaks immediately with an optionally available separator
			if opt.Mode == BreakModeStrict || forceBreak {
				performStrictLineBreak(&lineTexts, &lineText, &i, &lineWidth, separatorIdx, opt)
			}
			continue
		}
		// regular break due to a new line rune
		if utf8Texts[i] == '\n' {
			lineTexts = append(lineTexts, string(lineText))
			lineText = lineText[0:0]
			lineWidth = 0
			continue
		}
		// end of text
		if i == utf8TextsLen-1 {
			lineText = append(lineText, utf8Texts[i])
			lineTexts = append(lineTexts, string(lineText))
		}
		// store overall index when separator would still fit in the currently processed text-line
		if opt.HasSeparator() && lineWidth+runeWidth+separatorWidth <= width {
			separatorIdx = i
		}

		lineText = append(lineText, utf8Texts[i])
		lineWidth += runeWidth
	}
	return lineTexts, nil
}

func performIndicatorSensitiveLineBreak(lineTexts *[]string, lineText *[]rune, i *int, lineWidth *float64, opt *BreakOption) bool {
	brIdx := breakIndicatorIndex(*lineText, opt.BreakIndicator)
	if brIdx > 0 {
		diff := len(*lineText) - brIdx
		*lineText = (*lineText)[0:brIdx]
		*lineTexts = append(*lineTexts, string(*lineText))
		*lineText = (*lineText)[0:0]
		*i -= diff
		*lineWidth = 0
		return true
	}
	return false
}

func performStrictLineBreak(lineTexts *[]string, lineText *[]rune, i *int, lineWidth *float64, separatorIdx int, opt *BreakOption) {
	if opt.HasSeparator() && separatorIdx > -1 {
		// trim the line to the last possible index with an appended separator
		trimIdx := *i - separatorIdx
		*lineText = (*lineText)[0 : len(*lineText)-trimIdx]
		// append separator to the line
		*lineText = append(*lineText, []rune(opt.Separator)...)
		*lineTexts = append(*lineTexts, string(*lineText))
		*lineText = (*lineText)[0:0]
		*i = separatorIdx - 1
		*lineWidth = 0
		return
	}
	*lineTexts = append(*lineTexts, string(*lineText))
	*lineText = (*lineText)[0:0]
	*lineWidth = 0
	*i--
}

// breakIndicatorIndex returns the index where a text line (i.e. rune slice) can be split "gracefully" by checking on
// the break indicator.
// In case no possible break can be identified -1 is returned.
func breakIndicatorIndex(text []rune, bi rune) int {
	for i := len(text) - 1; i > 0; i-- {
		if text[i] == bi {
			return i
		}
	}
	return -1
}

// AddTTFFontByReader adds font data by reader.
func (gp *GoPdf) AddTTFFontData(family string, fontData []byte) error {
	return gp.AddTTFFontDataWithOption(family, fontData, defaultTtfFontOption())
}

// AddTTFFontDataWithOption adds font data with option.
func (gp *GoPdf) AddTTFFontDataWithOption(family string, fontData []byte, option TtfOption) error {
	subsetFont := new(SubsetFontObj)
	subsetFont.init(func() *GoPdf {
		return gp
	})
	subsetFont.SetTtfFontOption(option)
	subsetFont.SetFamily(family)
	err := subsetFont.SetTTFData(fontData)
	if err != nil {
		return err
	}

	return gp.setSubsetFontObject(subsetFont, family, option)
}

// setSubsetFontObject sets SubsetFontObj.
// The given SubsetFontObj is expected to be configured in advance.
func (gp *GoPdf) setSubsetFontObject(subsetFont *SubsetFontObj, family string, option TtfOption) error {
	_ = gp.addObj(subsetFont) //add หลังสุด
	return nil
}

// MeasureTextWidth : measure Width of text (use current font)
func (gp *GoPdf) MeasureTextWidth(text string) (float64, error) {

	text, err := gp.curr.FontISubset.AddChars(text) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	_, _, textWidthPdfUnit, err := createContent(gp.curr.FontISubset, text, gp.curr.FontSize, gp.curr.CharSpacing, nil)
	if err != nil {
		return 0, err
	}
	return pointsToUnits(gp.config.Unit, textWidthPdfUnit), nil
}

func (gp *GoPdf) MeasureNextRuneWidth(r, leftRune rune) (float64, error) {

	r, err := gp.curr.FontISubset.AddRune(r) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	textWidthPdfUnit, err := createNextRuneContent(gp.curr.FontISubset, r, leftRune, gp.curr.FontSize, gp.curr.CharSpacing)
	if err != nil {
		return 0, err
	}
	return pointsToUnits(gp.config.Unit, textWidthPdfUnit), nil
}

/*---private---*/

// init
func (gp *GoPdf) init() {
	//init curr
	gp.resetCurrXY()
	gp.curr = Current{}

	// change the unit type
	gp.config.PageSize = *gp.config.PageSize.unitsToPoints(gp.config.Unit)

	// init gofpdi free pdf document importer
	//gp.fpdi = importerOrDefault(importer...)

}

//func importerOrDefault(importer ...*gofpdi.Importer) *gofpdi.Importer {
//	if len(importer) != 0 {
//		return importer[len(importer)-1]
//	}
//	return gofpdi.NewImporter()
//}

func (gp *GoPdf) resetCurrXY() {
}

// UnitsToPoints converts the units to the documents unit type
func (gp *GoPdf) UnitsToPoints(u float64) float64 {
	return unitsToPoints(gp.config.Unit, u)
}

// UnitsToPointsVar converts the units to the documents unit type for all variables passed in
func (gp *GoPdf) UnitsToPointsVar(u ...*float64) {
	unitsToPointsVar(gp.config.Unit, u...)
}

// PointsToUnits converts the points to the documents unit type
func (gp *GoPdf) PointsToUnits(u float64) float64 {
	return pointsToUnits(gp.config.Unit, u)
}

// PointsToUnitsVar converts the points to the documents unit type for all variables passed in
func (gp *GoPdf) PointsToUnitsVar(u ...*float64) {
	pointsToUnitsVar(gp.config.Unit, u...)
}

func (gp *GoPdf) addObj(iobj *SubsetFontObj) int {
	index := len(gp.pdfObjs)
	gp.pdfObjs = append(gp.pdfObjs, iobj)
	return index
}

//tool for validate pdf https://www.pdf-online.com/osa/validate.aspx
