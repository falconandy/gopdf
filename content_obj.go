package gopdf

import (
	"fmt"
	"io"
)

// ContentObj content object
type ContentObj struct { //impl IObj
	listCache listCacheContent
	//text bytes.Buffer
	getRoot func() *GoPdf
}

func (c *ContentObj) init(funcGetRoot func() *GoPdf) {
	c.getRoot = funcGetRoot
}

func (c *ContentObj) write(w io.Writer, objID int) error {
	buff := GetBuffer()
	defer PutBuffer(buff)

	if err := c.listCache.write(buff); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "<<\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "/Length %d\n", buff.Len()); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ">>\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "stream\n"); err != nil {
		return err
	}

	if _, err := buff.WriteTo(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "endstream\n"); err != nil {
		return err
	}

	return nil
}

func (c *ContentObj) getType() string {
	return "Content"
}

func (c *ContentObj) GetCacheContentImage(index int, opts ImageOptions) *cacheContentImage {
	h := c.getRoot().config.PageSize.H

	return &cacheContentImage{
		pageHeight: h,
		index:      index,
		x:          opts.X,
		y:          opts.Y,
		rect:       *opts.Rect,
	}
}

// AppendStreamImage append image
func (c *ContentObj) AppendStreamImage(index int, opts ImageOptions) {
	cache := c.GetCacheContentImage(index, opts)
	c.listCache.append(cache)
}
