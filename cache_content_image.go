package gopdf

import (
	"fmt"
	"io"
)

type cacheContentImage struct {
	index      int
	x          float64
	y          float64
	pageHeight float64
	rect       Rect
}

func (c *cacheContentImage) write(writer io.Writer) error {
	width := c.rect.W
	height := c.rect.H

	contentStream := "q\n"

	x := c.x
	y := c.pageHeight - c.y

	y -= height

	var maskImageRotateMat string

	contentStream += fmt.Sprintf("q\n %s %0.2f 0 0\n %0.2f %0.2f %0.2f cm /I%d Do \nQ\n", maskImageRotateMat, width, height, x, y, c.index+1)

	contentStream += "Q\n"

	if _, err := io.WriteString(writer, contentStream); err != nil {
		return err
	}

	return nil
}
