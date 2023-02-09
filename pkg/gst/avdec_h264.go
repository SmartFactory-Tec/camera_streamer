package gst

type AvDecH264 struct {
	Element
}

func NewAvDecH264(name string) (*AvDecH264, error) {
	element, err := makeElement(name, "avdec_h264")

	if err != nil {
		return nil, err
	}

	avDecH264 := AvDecH264{element}
	enableGarbageCollection(&avDecH264)

	return &avDecH264, nil
}
