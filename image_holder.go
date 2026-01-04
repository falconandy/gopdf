package gopdf

// ImageHolder hold image data
type ImageHolder struct {
	Data []byte
}

// ImageHolderByBytes create ImageHolder by []byte
func ImageHolderByBytes(b []byte) (*ImageHolder, error) {
	return &ImageHolder{
		Data: b,
	}, nil
}
