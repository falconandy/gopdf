package gopdf

import (
	"fmt"
	"io"
)

// DeviceRGBObj  DeviceRGB
type DeviceRGBObj struct {
	data    []byte
	getRoot func() *GoPdf
}

func (d *DeviceRGBObj) init(funcGetRoot func() *GoPdf) {
	d.getRoot = funcGetRoot
}

func (d *DeviceRGBObj) getType() string {
	return "devicergb"
}

// สร้าง ข้อมูลใน pdf
func (d *DeviceRGBObj) write(w io.Writer, objID int) error {

	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "/Length %d\n", len(d.data))
	io.WriteString(w, ">>\n")
	io.WriteString(w, "stream\n")
	w.Write(d.data)
	io.WriteString(w, "endstream\n")

	return nil
}
