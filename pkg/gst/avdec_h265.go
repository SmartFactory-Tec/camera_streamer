package gst

type AvDecH265 struct {
	*BaseElement
}

func NewAvDecH265(name string) (AvDecH265, error) {
	createdElement, err := NewGstElement("avdec_h265", name)

	if err != nil {
		return AvDecH265{}, err
	}

	return AvDecH265{&createdElement}, nil
}
