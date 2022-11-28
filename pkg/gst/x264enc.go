package gst

type X264Enc struct {
	*BaseElement
}

func NewX264Enc(name string) (X264Enc, error) {
	createdElement, err := NewGstElement("x264enc", name)

	if err != nil {
		return X264Enc{}, err
	}

	return X264Enc{&createdElement}, nil
}
