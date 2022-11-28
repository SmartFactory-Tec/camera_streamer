package gst

type H265Parse struct {
	*BaseElement
}

func NewH265Parse(name string) (H265Parse, error) {
	createdElement, err := NewGstElement("h265parse", name)

	if err != nil {
		return H265Parse{}, err
	}

	return H265Parse{&createdElement}, nil
}
