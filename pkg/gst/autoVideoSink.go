package gst

type AutoVideoSink struct {
	*BaseElement
}

func NewAutoVideoSink(name string) (AutoVideoSink, error) {
	createdElement, err := NewGstElement("autovideosink", name)

	if err != nil {
		return AutoVideoSink{}, err
	}

	return AutoVideoSink{&createdElement}, nil
}
