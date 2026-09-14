package gopdf

type TTFParser struct {
	ZFontData []byte

	ZTables map[string]TableDirectoryEntry

	ZUnitsPerEm uint
	ZXMin       int
	ZYMin       int
	ZXMax       int
	ZYMax       int

	ZNumberOfHMetrics uint
	ZAscender         int
	ZDescender        int

	ZNumGlyphs uint
	ZWidths    []uint

	ZTypoAscender  int
	ZTypoDescender int
	ZCapHeight     int
	ZXHeight       int

	ZItalicAngle int

	IsShortIndex  bool
	LocaTable     []uint
	SegCount      uint
	StartCount    []uint
	EndCount      []uint
	IdRangeOffset []uint
	IdDelta       []uint
	GlyphIdArray  []uint
	ZFlag         int

	ZGroupingTables []CmapFormat12GroupingTable
}

type TableDirectoryEntry struct {
	CheckSum uint
	Offset   uint
	Length   uint
}

func (t TableDirectoryEntry) PaddedLength() int {
	l := int(t.Length)
	return (l + 3) & ^3
}

type CmapFormat12GroupingTable struct {
	StartCharCode, EndCharCode, GlyphID uint
}

func (t *TTFParser) GroupingTables() []CmapFormat12GroupingTable {
	return t.ZGroupingTables
}

func (t *TTFParser) XHeight() int {
	return t.ZXHeight
}

func (t *TTFParser) XMin() int {
	return t.ZXMin
}

func (t *TTFParser) YMin() int {
	return t.ZYMin
}

func (t *TTFParser) XMax() int {
	return t.ZXMax
}

func (t *TTFParser) YMax() int {
	return t.ZYMax
}

func (t *TTFParser) ItalicAngle() int {
	return t.ZItalicAngle
}

func (t *TTFParser) Flag() int {
	return t.ZFlag
}

func (t *TTFParser) Ascender() int {
	return t.ZAscender
}

func (t *TTFParser) Descender() int {
	return t.ZDescender
}

func (t *TTFParser) TypoAscender() int {
	return t.ZTypoAscender
}

func (t *TTFParser) TypoDescender() int {
	return t.ZTypoDescender
}

func (t *TTFParser) CapHeight() int {
	return t.ZCapHeight
}

func (t *TTFParser) NumGlyphs() uint {
	return t.ZNumGlyphs
}

func (t *TTFParser) UnitsPerEm() uint {
	return t.ZUnitsPerEm
}

func (t *TTFParser) NumberOfHMetrics() uint {
	return t.ZNumberOfHMetrics
}

func (t *TTFParser) Widths() []uint {
	return t.ZWidths
}

func (t *TTFParser) GetTables() map[string]TableDirectoryEntry {
	return t.ZTables
}

func (t *TTFParser) FontData() []byte {
	return t.ZFontData
}

func Round(value float64) int {
	if value < 0.0 {
		value -= 0.5
	} else {
		value += 0.5
	}
	return int(value)
}
