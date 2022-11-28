package gst

type VideoScale struct {
	*BaseElement
}

func NewVideoScale(name string) (VideoScale, error) {
	createdElement, err := NewGstElement("videoscale", name)

	if err != nil {
		return VideoScale{}, err
	}

	return VideoScale{&createdElement}, nil
}
