package gst

type AvDecH265 struct {
	Element
}

func NewAvDecH265(name string) (*AvDecH265, error) {
	element, err := makeElement(name, "avdec_h265")

	if err != nil {
		return nil, err
	}

	avDecH265 := AvDecH265{element}
	enableGarbageCollection(&avDecH265)

	return &avDecH265, nil
}
