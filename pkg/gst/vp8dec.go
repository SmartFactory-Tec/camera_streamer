package gst

type Vp8Dec struct {
	*BaseElement
}

func NewVp8Dec(name string) (Vp8Dec, error) {
	createdElement, err := NewGstElement("vp8dec", name)

	if err != nil {
		return Vp8Dec{}, err
	}

	return Vp8Dec{&createdElement}, nil
}
