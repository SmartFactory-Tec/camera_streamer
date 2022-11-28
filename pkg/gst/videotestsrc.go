package gst

type VideoTestSrc struct {
	*BaseElement
}

func NewVideoTestSrc(name string) (VideoTestSrc, error) {
	createdElement, err := NewGstElement("videotestsrc", name)

	if err != nil {
		return VideoTestSrc{}, err
	}

	return VideoTestSrc{&createdElement}, nil
}
