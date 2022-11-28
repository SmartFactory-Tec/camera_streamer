package gst

type Vp8Enc struct {
	*BaseElement
}

func NewVp8Enc(name string) (Vp8Enc, error) {
	createdElement, err := NewGstElement("vp8enc", name)

	if err != nil {
		return Vp8Enc{}, err
	}

	return Vp8Enc{&createdElement}, nil
}
