package gopdf

import (
	"bytes"
	"fmt"
	"io"
)

// ImageObj image object
type ImageObj struct {
	//imagepath string
	rawImgReader *bytes.Reader
	imginfo      imgInfo
	//getRoot func() *GoPdf
}

func (i *ImageObj) init(funcGetRoot func() *GoPdf) {

}

func (i *ImageObj) write(w io.Writer, objID int) error {
	data := i.imginfo.data

	if err := writeImgProps(w, i.imginfo); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "\t/Length %d\n>>\n", len(data)); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "stream\n"); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "\nendstream\n"); err != nil {
		return err
	}

	return nil
}

func (i *ImageObj) isColspaceIndexed() bool {
	return isColspaceIndexed(i.imginfo)
}

func (i *ImageObj) createDeviceRGB() (*DeviceRGBObj, error) {
	var dRGB DeviceRGBObj
	dRGB.data = i.imginfo.pal
	return &dRGB, nil
}

func (i *ImageObj) getType() string {
	return "Image"
}

// SetImage set image
func (i *ImageObj) SetImage(data []byte) error {
	i.rawImgReader = bytes.NewReader(data)

	return nil
}

func (i *ImageObj) parse() error {

	i.rawImgReader.Seek(0, 0)
	imginfo, err := parseImg(i.rawImgReader)
	if err != nil {
		return err
	}
	i.imginfo = imginfo

	return nil
}
