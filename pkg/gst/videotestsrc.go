package gst

type VideoTestSrc struct {
	Element
}

func NewVideoTestSrc(name string) (*VideoTestSrc, error) {
	element, err := makeElement(name, "videotestsrc")

	if err != nil {
		return nil, err
	}

	videoTestSrc := VideoTestSrc{element}
	enableGarbageCollection(&videoTestSrc)

	return &videoTestSrc, nil
}
