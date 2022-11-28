package gst

type H264Parse struct {
	*BaseElement
}

func NewH264Parse(name string) (H264Parse, error) {
	createdElement, err := NewGstElement("h264parse", name)

	if err != nil {
		return H264Parse{}, err
	}

	return H264Parse{&createdElement}, nil
}
