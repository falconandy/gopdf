package gopdf

import (
	"io"
)

type ICacheContent interface {
	write(w io.Writer) error
}
