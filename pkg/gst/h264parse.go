package gst

type H264Parse struct {
	Element
}

func NewH264Parse(name string) (*H264Parse, error) {
	element, err := makeElement(name, "h264parse")

	if err != nil {
		return nil, err
	}

	h264Parse := H264Parse{
		element,
	}
	enableGarbageCollection(&h264Parse)

	return &h264Parse, nil
}
