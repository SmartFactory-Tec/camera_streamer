package gst

type H265Parse struct {
	Element
}

func NewH265Parse(name string) (*H265Parse, error) {
	element, err := makeElement(name, "h265parse")

	if err != nil {
		return nil, err
	}

	h265Parse := H265Parse{
		element,
	}
	enableGarbageCollection(&h265Parse)

	return &h265Parse, nil
}
