package gst

type AvDecH264 struct {
	*BaseElement
}

func NewAvDecH264(name string) (AvDecH264, error) {
	createdElement, err := NewGstElement("avdec_h264", name)

	if err != nil {
		return AvDecH264{}, err
	}

	return AvDecH264{&createdElement}, nil
}
