package gopdf

// PageOption option of page
type PageOption struct {
	PageSize *Rect
}

func (p PageOption) isEmpty() bool {
	return p.PageSize == nil
}
