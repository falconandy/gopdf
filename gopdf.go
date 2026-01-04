package gopdf

import (
	"fmt"
	"io"
	"strconv"
)

// GoPdf : A simple library for generating PDF written in Go lang
type GoPdf struct {
	pdfObjs []IObj
	config  Config

	/*---index ของ obj สำคัญๆ เก็บเพื่อลด loop ตอนค้นหา---*/
	//index ของ obj pages
	indexOfPagesObj int

	indexOfContent int

	//index ของ procset ซึ่งควรจะมีอันเดียว
	indexOfProcSet int
}

type ImageOptions struct {
	X    float64
	Y    float64
	Rect *Rect
}

// ImageByHolder : draw image by ImageHolder
func (gp *GoPdf) ImageByHolder(img *ImageHolder, x float64, y float64, rect *Rect) error {
	gp.UnitsToPointsVar(&x, &y)

	rect = rect.UnitsToPoints(gp.config.Unit)

	imageOptions := ImageOptions{
		X:    x,
		Y:    y,
		Rect: rect,
	}

	return gp.imageByHolder(img, imageOptions)
}

func (gp *GoPdf) imageByHolder(img *ImageHolder, opts ImageOptions) error {
	//create img object
	imgobj := new(ImageObj)

	imgobj.init(func() *GoPdf {
		return gp
	})

	err := imgobj.SetImage(img.Data)
	if err != nil {
		return err
	}

	err = imgobj.parse()
	if err != nil {
		return err
	}
	index := gp.addObj(imgobj)
	if gp.indexOfProcSet != -1 {
		//ยัดรูป
		procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
		gp.getContent().AppendStreamImage(index, opts)
		procset.RelateXobjs = append(procset.RelateXobjs, RelateXobject{IndexOfObj: index})
	}

	if imgobj.isColspaceIndexed() {
		dRGB, err := imgobj.createDeviceRGB()
		if err != nil {
			return err
		}
		dRGB.getRoot = func() *GoPdf {
			return gp
		}
		imgobj.imginfo.deviceRGBObjID = gp.addObj(dRGB)
	}

	return nil
}

// AddPage : add new page
func (gp *GoPdf) AddPage() {
	page := new(PageObj)
	page.init(func() *GoPdf {
		return gp
	})

	page.ResourcesRelate = strconv.Itoa(gp.indexOfProcSet+1) + " 0 R"
	gp.addObj(page)

	//reset
	gp.indexOfContent = -1
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
	//init all basic obj
	catalog := new(CatalogObj)
	catalog.init(func() *GoPdf {
		return gp
	})
	pages := new(PagesObj)
	pages.init(func() *GoPdf {
		return gp
	})
	gp.addObj(catalog)
	gp.indexOfPagesObj = gp.addObj(pages)

	//indexOfProcSet
	procset := new(ProcSetObj)
	procset.init(func() *GoPdf {
		return gp
	})
	gp.indexOfProcSet = gp.addObj(procset)
}

// WriteTo implements the io.WriterTo interface and can
// be used to stream the PDF as it is compiled to an io.Writer.
func (gp *GoPdf) WriteTo(w io.Writer) (n int64, err error) {
	return gp.compilePdf(w)
}

func (gp *GoPdf) compilePdf(w io.Writer) (n int64, err error) {
	gp.prepare()

	max := len(gp.pdfObjs)
	writer := newCountingWriter(w)
	fmt.Fprint(writer, "%PDF-1.7\n%����\n\n")
	linelens := make([]int64, max)
	i := 0

	for i < max {
		objID := i + 1
		linelens[i] = writer.offset
		pdfObj := gp.pdfObjs[i]
		fmt.Fprintf(writer, "%d 0 obj\n", objID)
		pdfObj.write(writer, objID)
		io.WriteString(writer, "endobj\n\n")
		i++
	}
	gp.xref(writer, writer.offset, linelens, i)
	return writer.offset, nil
}

type (
	countingWriter struct {
		offset int64
		writer io.Writer
	}
)

func newCountingWriter(w io.Writer) *countingWriter {
	return &countingWriter{writer: w}
}

func (cw *countingWriter) Write(b []byte) (int, error) {
	n, err := cw.writer.Write(b)
	cw.offset += int64(n)
	return n, err
}

/*---private---*/

// init
func (gp *GoPdf) init() {
	gp.pdfObjs = []IObj{}

	//init index
	gp.indexOfPagesObj = -1
	gp.indexOfContent = -1

	// change the unit type
	gp.config.PageSize = *gp.config.PageSize.UnitsToPoints(gp.config.Unit)

	// init gofpdi free pdf document importer
	//gp.fpdi = importerOrDefault(importer...)

}

//func importerOrDefault(importer ...*gofpdi.Importer) *gofpdi.Importer {
//	if len(importer) != 0 {
//		return importer[len(importer)-1]
//	}
//	return gofpdi.NewImporter()
//}

// UnitsToPointsVar converts the units to the documents unit type for all variables passed in
func (gp *GoPdf) UnitsToPointsVar(u ...*float64) {
	unitsToPointsVar(gp.config.Unit, u...)
}

func (gp *GoPdf) prepare() {

	if gp.indexOfPagesObj != -1 {
		indexCurrPage := -1
		pagesObj := gp.pdfObjs[gp.indexOfPagesObj].(*PagesObj)
		i := 0 //gp.indexOfFirstPageObj
		max := len(gp.pdfObjs)
		for i < max {
			objtype := gp.pdfObjs[i].getType()
			switch objtype {
			case "Page":
				pagesObj.Kids = fmt.Sprintf("%s %d 0 R ", pagesObj.Kids, i+1)
				pagesObj.PageCount++
				indexCurrPage = i
			case "Content":
				if indexCurrPage != -1 {
					gp.pdfObjs[indexCurrPage].(*PageObj).Contents = fmt.Sprintf("%s %d 0 R ", gp.pdfObjs[indexCurrPage].(*PageObj).Contents, i+1)
				}
			}
			i++
		}
	}
}

func (gp *GoPdf) xref(w io.Writer, xrefbyteoffset int64, linelens []int64, i int) error {

	io.WriteString(w, "xref\n")
	fmt.Fprintf(w, "0 %d\n", i+1)
	io.WriteString(w, "0000000000 65535 f \n")
	j := 0
	max := len(linelens)
	for j < max {
		linelen := linelens[j]
		fmt.Fprintf(w, "%s 00000 n \n", gp.formatXrefline(linelen))
		j++
	}
	io.WriteString(w, "trailer\n")
	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "/Size %d\n", max+1)
	io.WriteString(w, "/Root 1 0 R\n")
	io.WriteString(w, ">>\n")
	io.WriteString(w, "startxref\n")
	fmt.Fprintf(w, "%d", xrefbyteoffset)
	io.WriteString(w, "\n%%EOF\n")

	return nil
}

// ปรับ xref ให้เป็น 10 หลัก
func (gp *GoPdf) formatXrefline(n int64) string {
	str := strconv.FormatInt(n, 10)
	for len(str) < 10 {
		str = "0" + str
	}
	return str
}

func (gp *GoPdf) addObj(iobj IObj) int {
	index := len(gp.pdfObjs)
	gp.pdfObjs = append(gp.pdfObjs, iobj)
	return index
}

func (gp *GoPdf) getContent() *ContentObj {
	var content *ContentObj
	if gp.indexOfContent <= -1 {
		content = new(ContentObj)
		content.init(func() *GoPdf {
			return gp
		})
		gp.indexOfContent = gp.addObj(content)
	} else {
		content = gp.pdfObjs[gp.indexOfContent].(*ContentObj)
	}
	return content
}

//tool for validate pdf https://www.pdf-online.com/osa/validate.aspx
