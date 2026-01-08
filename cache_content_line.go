package gopdf

import (
	"fmt"
	"io"
)

type cacheContentLine struct {
	pageHeight float64
	x1         float64
	y1         float64
	x2         float64
	y2         float64
}

func (c *cacheContentLine) write(w io.Writer) error {
	fmt.Fprintf(w, "q\n")
	fmt.Fprintf(w, "%0.2f %0.2f m %0.2f %0.2f l S\n", c.x1, c.pageHeight-c.y1, c.x2, c.pageHeight-c.y2)
	fmt.Fprintf(w, "Q\n")
	return nil
}
