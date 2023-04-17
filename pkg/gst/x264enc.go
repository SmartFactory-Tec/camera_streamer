package gst

type X264Enc struct {
	Element
}

func NewX264Enc(name string) (*X264Enc, error) {
	element, err := makeElement(name, "x264enc")

	if err != nil {
		return nil, err
	}

	x264Enc := X264Enc{element}
	enableGarbageCollection(&x264Enc)

	return &x264Enc, nil
}
