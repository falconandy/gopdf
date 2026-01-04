package gopdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

func writeImgProps(w io.Writer, imginfo imgInfo) error {
	if err := writeBaseImgProps(w, imginfo, imginfo.colspace); err != nil {
		return err
	}

	if strings.TrimSpace(imginfo.decodeParms) != "" {
		if _, err := fmt.Fprintf(w, "\t/DecodeParms <<%s>>\n", imginfo.decodeParms); err != nil {
			return err
		}
	}

	if imginfo.trns != nil && len(imginfo.trns) > 0 {
		j := 0
		content := "\t/Mask ["
		max := len(imginfo.trns)

		for j < max {
			content += fmt.Sprintf("\t\t%d ", imginfo.trns[j])
			content += fmt.Sprintf("\t\t%d ", imginfo.trns[j])
			j++
		}

		content += "\t]\n"

		if _, err := io.WriteString(w, content); err != nil {
			return err
		}
	}

	return nil
}

func writeBaseImgProps(w io.Writer, imginfo imgInfo, colorSpace string) error {
	content := "<<\n"
	content += "\t/Type /XObject\n"
	content += "\t/Subtype /Image\n"
	content += fmt.Sprintf("\t/Width %d\n", imginfo.w)
	content += fmt.Sprintf("\t/Height %d\n", imginfo.h)

	if isColspaceIndexed(imginfo) {
		size := len(imginfo.pal)/3 - 1
		content += fmt.Sprintf("\t/ColorSpace [/Indexed /DeviceRGB %d %d 0 R]\n", size, imginfo.deviceRGBObjID+1)
	} else {
		content += fmt.Sprintf("\t/ColorSpace /%s\n", colorSpace)
		if imginfo.colspace == "DeviceCMYK" {
			content += "\t/Decode [1 0 1 0 1 0 1 0]\n"
		}
	}

	content += fmt.Sprintf("\t/BitsPerComponent %s\n", imginfo.bitsPerComponent)

	if strings.TrimSpace(imginfo.filter) != "" {
		content += fmt.Sprintf("\t/Filter /%s\n", imginfo.filter)
	}

	if _, err := io.WriteString(w, content); err != nil {
		return err
	}

	return nil
}

func isColspaceIndexed(imginfo imgInfo) bool {
	if imginfo.colspace == "Indexed" {
		return true
	}
	return false
}

func parseImg(raw *bytes.Reader) (imgInfo, error) {
	// fmt.Printf("----------\n")
	var info imgInfo

	err := parsePng(raw, &info)
	if err != nil {
		return info, err
	}

	// fmt.Printf("%#v\n", info)

	return info, nil
}

var pngMagicNumber = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
var pngIHDR = []byte{0x49, 0x48, 0x44, 0x52}

func parsePng(f *bytes.Reader, info *imgInfo) error {
	// f := bytes.NewReader(raw)
	f.Seek(0, 0)
	b, err := readBytes(f, 8)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, pngMagicNumber) {
		return errors.New("Not a PNG file")
	}

	f.Seek(4, 1) // skip header chunk
	b, err = readBytes(f, 4)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, pngIHDR) {
		return errors.New("Incorrect PNG file")
	}

	w, err := readInt(f)
	if err != nil {
		return err
	}
	h, err := readInt(f)
	if err != nil {
		return err
	}
	// fmt.Printf("w=%d h=%d\n", w, h)

	bpc, err := readBytes(f, 1)
	if err != nil {
		return err
	}

	if bpc[0] > 8 {
		return errors.New("16-bit depth not supported")
	}

	ct, err := readBytes(f, 1)
	if err != nil {
		return err
	}

	var colspace string
	switch ct[0] {
	case 0, 4:
		colspace = "DeviceGray"
	case 2, 6:
		colspace = "DeviceRGB"
	case 3:
		colspace = "Indexed"
	default:
		return errors.New("Unknown color type")
	}

	compressionMethod, err := readBytes(f, 1)
	if err != nil {
		return err
	}
	if compressionMethod[0] != 0 {
		return errors.New("Unknown compression method")
	}

	filterMethod, err := readBytes(f, 1)
	if err != nil {
		return err
	}
	if filterMethod[0] != 0 {
		return errors.New("Unknown filter method")
	}

	interlacing, err := readBytes(f, 1)
	if err != nil {
		return err
	}
	if interlacing[0] != 0 {
		return errors.New("Interlacing not supported")
	}

	_, err = f.Seek(4, 1) // skip
	if err != nil {
		return err
	}

	// decodeParms := "/Predictor 15 /Colors '.($colspace=='DeviceRGB' ? 3 : 1).' /BitsPerComponent '.$bpc.' /Columns '.$w;

	var pal []byte
	var trns []byte
	var data []byte
	for {
		un, err := readUInt(f)
		if err != nil {
			return err
		}
		n := int(un)
		typ, err := readBytes(f, 4)
		// fmt.Printf(">>>>%+v-%s-%d\n", typ, string(typ), n)
		if err != nil {
			return err
		}

		if string(typ) == "PLTE" {
			pal, err = readBytes(f, n)
			if err != nil {
				return err
			}
			_, err = f.Seek(int64(4), 1) // skip
			if err != nil {
				return err
			}
		} else if string(typ) == "tRNS" {

			var t []byte
			t, err = readBytes(f, n)
			if err != nil {
				return err
			}

			if ct[0] == 0 {
				trns = []byte{(t[1])}
			} else if ct[0] == 2 {
				trns = []byte{t[1], t[3], t[5]}
			} else {
				pos := strings.Index(string(t), "\x00")
				if pos >= 0 {
					trns = []byte{byte(pos)}
				}
			}

			_, err = f.Seek(int64(4), 1) // skip
			if err != nil {
				return err
			}

		} else if string(typ) == "IDAT" {
			// fmt.Printf("n=%d\n\n", n)
			var d []byte
			d, err = readBytes(f, n)
			if err != nil {
				return err
			}
			data = append(data, d...)
			_, err = f.Seek(int64(4), 1) // skip
			if err != nil {
				return err
			}
		} else if string(typ) == "IEND" {
			break
		} else {
			_, err = f.Seek(int64(n+4), 1) // skip
			if err != nil {
				return err
			}
		}

		if n <= 0 {
			break
		}
	} // end for

	// info.data = data //ok
	info.trns = trns
	info.pal = pal

	// fmt.Printf("data= %x", md5.Sum(data))

	if colspace == "Indexed" && strings.TrimSpace(string(pal)) == "" {
		return errors.New("Missing palette")
	}

	info.w = w
	info.h = h
	info.colspace = colspace
	info.bitsPerComponent = fmt.Sprintf("%d", int(bpc[0]))
	info.filter = "FlateDecode"

	colors := 1
	if colspace == "DeviceRGB" {
		colors = 3
	}
	info.decodeParms = fmt.Sprintf("/Predictor 15 /Colors  %d /BitsPerComponent %s /Columns %d", colors, info.bitsPerComponent, w)

	// fmt.Printf("%d = ct[0]\n", ct[0])
	// fmt.Printf("%x\n", md5.Sum(data))
	if ct[0] >= 4 {
		zipReader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer zipReader.Close()
		afterZipData, err := io.ReadAll(zipReader)
		if err != nil {
			return err
		}

		var color []byte
		var alpha []byte
		if ct[0] == 4 {
			// Gray image
			length := 2 * w
			i := 0
			for i < h {
				pos := (1 + length) * i
				color = append(color, afterZipData[pos])
				alpha = append(alpha, afterZipData[pos])
				line := afterZipData[pos+1 : pos+length+1]
				j := 0
				max := len(line)
				for j < max {
					color = append(color, line[j])
					j++
					alpha = append(alpha, line[j])
					j++
				}
				i++
			}
			// fmt.Print("aaaaa")

		} else {
			// RGB image
			length := 4 * w
			i := 0
			for i < h {
				pos := (1 + length) * i
				color = append(color, afterZipData[pos])
				alpha = append(alpha, afterZipData[pos])
				line := afterZipData[pos+1 : pos+length+1]
				j := 0
				max := len(line)
				for j < max {
					color = append(color, line[j:j+3]...)
					alpha = append(alpha, line[j+3])
					j = j + 4
				}

				i++
			}
		}

		info.data, err = compress(color)
		if err != nil {
			return err
		}

	} else {
		info.data = data
	}

	return nil
}

func compress(data []byte) ([]byte, error) {
	var results []byte
	var buff bytes.Buffer
	zwr, err := zlib.NewWriterLevel(&buff, zlib.BestSpeed)

	if err != nil {
		return results, err
	}
	_, err = zwr.Write(data)
	if err != nil {
		return nil, err
	}
	zwr.Close()
	return buff.Bytes(), nil
}

func readUInt(f *bytes.Reader) (uint, error) {
	buff, err := readBytes(f, 4)
	// fmt.Printf("%#v\n\n", buff)
	if err != nil {
		return 0, err
	}
	n := binary.BigEndian.Uint32(buff)
	return uint(n), nil
}

func readInt(f *bytes.Reader) (int, error) {
	u, err := readUInt(f)
	if err != nil {
		return 0, err
	}
	v := int(u)
	return v, nil
}

func readBytes(f *bytes.Reader, len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := f.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
