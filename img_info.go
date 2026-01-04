package gopdf

type imgInfo struct {
	w, h int
	//src              string
	colspace         string
	bitsPerComponent string
	filter           string
	decodeParms      string
	trns             []byte
	pal              []byte
	deviceRGBObjID   int
	data             []byte
}
