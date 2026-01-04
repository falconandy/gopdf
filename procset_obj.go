package gopdf

import (
	"fmt"
	"io"
)

// ProcSetObj is a PDF procSet object.
type ProcSetObj struct {
	RelateXobjs RelateXobjects
	getRoot     func() *GoPdf
}

func (pr *ProcSetObj) init(funcGetRoot func() *GoPdf) {
	pr.getRoot = funcGetRoot
}

func (pr *ProcSetObj) write(w io.Writer, objID int) error {
	content := "<<\n"
	content += "\t/ProcSet [/PDF /Text /ImageB /ImageC /ImageI]\n"

	fonts := "\t/Font <<\n"
	fonts += "\t>>\n"

	content += fonts

	xobjects := "\t/XObject <<\n"
	for _, XObject := range pr.RelateXobjs {
		xobjects += fmt.Sprintf("\t\t/I%d %d 0 R\n", XObject.IndexOfObj+1, XObject.IndexOfObj+1)
	}
	xobjects += "\t>>\n"

	content += xobjects

	extGStates := "\t/ExtGState <<\n"
	extGStates += "\t>>\n"

	content += extGStates

	content += ">>\n"

	if _, err := io.WriteString(w, content); err != nil {
		return err
	}

	return nil
}

func (pr *ProcSetObj) getType() string {
	return "ProcSet"
}

// RelateXobjects is a slice of RelateXobject.
type RelateXobjects []RelateXobject

// RelateXobject is an index for ???
type RelateXobject struct {
	IndexOfObj int
}
