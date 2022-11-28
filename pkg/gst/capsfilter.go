package gst

type CapsFilter struct {
	*BaseElement
}

func NewCapsFilter(name string) (CapsFilter, error) {
	createdElement, err := NewGstElement("capsfilter", name)

	if err != nil {
		return CapsFilter{}, err
	}

	return CapsFilter{&createdElement}, nil
}
