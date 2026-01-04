package gopdf

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestDeclarationImages(t *testing.T) {
	const (
		marginL = 5
		marginT = 5
		marginR = 5
		marginB = 5
	)

	pageSize := Rect{
		W: 210,
		H: 297,
	}

	pdf := &GoPdf{}
	pdf.Start(Config{
		Unit:     UnitMM,
		PageSize: pageSize,
	})

	image1, _ := os.ReadFile("1152017_5.08000_14-original-0-51fe1bcf5392e7eb48947fd8787449c0.png")
	image2, _ := os.ReadFile("1152017_5.08000_14-original-1-8396e688e93531cbb3fff970fd20baa6.png")
	image3, _ := os.ReadFile("1152017_5.08000_14-original-3-5446d6a184d994e6053a2cd0211b5abe.png")
	image4, _ := os.ReadFile("1152017_5.08000_14-original-4-3487ea53adb2ca26e36c27074ac2ada4.png")

	pages := [][]byte{image1, image2, image3, image4}

	for _, page := range pages {
		pdf.AddPage()
		imageHolder, err := ImageHolderByBytes(page)
		if err != nil {
			log.Fatal(err)
		}
		err = pdf.ImageByHolder(imageHolder, marginL, marginT, &Rect{
			W: pageSize.W - marginL - marginR,
			H: pageSize.H - marginT - marginB,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	b := bytes.NewBuffer(nil)
	_, err := pdf.WriteTo(b)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(b.Len())

	os.WriteFile("/tmp/out.pdf", b.Bytes(), 0644)
}
